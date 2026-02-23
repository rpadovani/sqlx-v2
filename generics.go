// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sqlx

import (
	"context"
	"database/sql"
	"fmt"
	"iter"
	"reflect"
	"runtime"
	"unsafe"
)

// SelectG executes a query and returns a slice of T, using generics for type safety.
// It uses pre-calculated struct field offsets for efficient scanning.
//
//	users, err := sqlx.SelectG[User](ctx, db, "SELECT * FROM users WHERE active = ?", true)
func SelectG[T any](ctx context.Context, q QueryerContext, query string, args ...any) (res []T, err error) {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("sqlx: error in SelectG: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var zero T
	elemType := reflect.TypeOf(zero)
	isPtr := false
	if elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
		isPtr = true
	}

	isUnsafe, strictTagParsing, mapper := extractMeta(q)

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("sqlx: error getting columns: %w", err)
	}

	if isScannable(elemType) {
		if len(columns) != 1 {
			return nil, fmt.Errorf("sqlx: scannable dest type %s with >1 columns (%d)", elemType, len(columns))
		}
		var results []T
		for rows.Next() {
			var val T
			var vp any
			if isPtr {
				vp = reflect.New(elemType).Interface()
			} else {
				vp = &val
			}
			if err := rows.Scan(vp); err != nil {
				return nil, fmt.Errorf("sqlx: scannable scan error: %w", err)
			}
			if isPtr {
				results = append(results, vp.(T))
			} else {
				results = append(results, val)
			}
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("sqlx: row error: %w", err)
		}
		return results, nil
	}

	tm := mapper.TypeMap(elemType)
	if strictTagParsing && len(tm.Errors) > 0 {
		return nil, fmt.Errorf("sqlx: strict tag error: %w", tm.Errors[0])
	}

	meta, err := buildMeta(elemType, columns, tm, isUnsafe)
	if err != nil {
		return nil, fmt.Errorf("sqlx: error building meta: %w", err)
	}

	scanDest := getScanDest(len(columns))
	defer putScanDest(scanDest)

	var results []T
	if !isPtr {
		for rows.Next() {
			if len(results) == cap(results) {
				newCap := cap(results) * 2
				if newCap == 0 {
					newCap = 8
				}
				newResults := make([]T, len(results), newCap)
				copy(newResults, results)
				results = newResults
			}
			i := len(results)
			results = append(results, zero)

			base := unsafe.Pointer(&results[i])
			populateScanDest(base, meta, scanDest)

			if err := rows.Scan(scanDest...); err != nil {
				return nil, fmt.Errorf("sqlx: struct scan error: %w", err)
			}
			runtime.KeepAlive(results)
		}
	} else {
		for rows.Next() {
			vp := reflect.New(elemType)
			base := unsafe.Pointer(vp.Pointer())
			populateScanDest(base, meta, scanDest)

			if err := rows.Scan(scanDest...); err != nil {
				return nil, fmt.Errorf("sqlx: struct scan error: %w", err)
			}
			results = append(results, vp.Interface().(T))
			runtime.KeepAlive(vp)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlx: row error: %w", err)
	}

	return results, nil
}

// GetG executes a query that is expected to return at most one row,
// scanning it into a value of type T. It uses pre-calculated struct field
// offsets for efficient scanning.
//
//	user, err := sqlx.GetG[User](ctx, db, "SELECT * FROM users WHERE id = ?", 1)
func GetG[T any](ctx context.Context, q QueryerContext, query string, args ...any) (res T, err error) {
	var zero T
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return zero, fmt.Errorf("sqlx: error in GetG: %w", err)
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	elemType := reflect.TypeOf(zero)
	isPtr := false
	if elemType.Kind() == reflect.Pointer {
		elemType = elemType.Elem()
		isPtr = true
	}

	isUnsafe, strictTagParsing, mapper := extractMeta(q)

	columns, err := rows.Columns()
	if err != nil {
		return zero, fmt.Errorf("sqlx: error getting columns: %w", err)
	}

	if isScannable(elemType) {
		if len(columns) != 1 {
			return zero, fmt.Errorf("sqlx: scannable dest type %s with >1 columns (%d)", elemType, len(columns))
		}
		if !rows.Next() {
			if err := rows.Err(); err != nil {
				return zero, fmt.Errorf("sqlx: row error: %w", err)
			}
			return zero, sql.ErrNoRows
		}
		var result T
		var vp any
		if isPtr {
			vp = reflect.New(elemType).Interface()
		} else {
			vp = &result
		}
		if err := rows.Scan(vp); err != nil {
			return zero, fmt.Errorf("sqlx: scannable scan error: %w", err)
		}
		if isPtr {
			return vp.(T), nil
		}
		return result, nil
	}

	tm := mapper.TypeMap(elemType)
	if strictTagParsing && len(tm.Errors) > 0 {
		return zero, fmt.Errorf("sqlx: strict tag error: %w", tm.Errors[0])
	}

	meta, err := buildMeta(elemType, columns, tm, isUnsafe)
	if err != nil {
		return zero, fmt.Errorf("sqlx: error building meta: %w", err)
	}

	scanDest := getScanDest(len(columns))
	defer putScanDest(scanDest)

	if !rows.Next() {
		if err := rows.Err(); err != nil {
			return zero, fmt.Errorf("sqlx: row error: %w", err)
		}
		return zero, sql.ErrNoRows
	}

	var result T
	if isPtr {
		vp := reflect.New(elemType)
		base := unsafe.Pointer(vp.Pointer())
		populateScanDest(base, meta, scanDest)
		if err := rows.Scan(scanDest...); err != nil {
			return zero, fmt.Errorf("sqlx: struct scan error: %w", err)
		}
		runtime.KeepAlive(vp)
		return vp.Interface().(T), nil
	} else {
		base := unsafe.Pointer(&result)
		populateScanDest(base, meta, scanDest)
		if err := rows.Scan(scanDest...); err != nil {
			return zero, fmt.Errorf("sqlx: struct scan error: %w", err)
		}
		runtime.KeepAlive(result)
		return result, nil
	}
}

// SelectIter returns an iterator (iter.Seq2) that streams rows one at a time,
// scanning each into a value of type T using pre-calculated struct field
// offsets. This avoids loading all results into memory at once.
//
//	for user, err := range sqlx.SelectIter[User](ctx, db, "SELECT * FROM users") {
//	    if err != nil {
//	        return err
//	    }
//	    process(user)
//	}
func SelectIter[T any](ctx context.Context, q QueryerContext, query string, args ...any) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		var zero T
		rows, err := q.QueryContext(ctx, query, args...)
		if err != nil {
			yield(zero, fmt.Errorf("sqlx: error in SelectIter rows: %w", err))
			return
		}

		elemType := reflect.TypeOf(zero)
		if elemType.Kind() == reflect.Pointer {
			elemType = reflect.TypeOf(zero).Elem()
		}

		isUnsafe, strictTagParsing, mapper := extractMeta(q)
		alloc := func() reflect.Value { return reflect.New(elemType) }

		err = iterateScan(rows, elemType, isUnsafe, strictTagParsing, mapper, alloc, func(vp reflect.Value) error {
			if !yield(vp.Elem().Interface().(T), nil) {
				return ErrStopIter
			}
			return nil
		})

		if err != nil {
			yield(zero, fmt.Errorf("sqlx: error in SelectIter scan: %w", err))
		}
	}
}
