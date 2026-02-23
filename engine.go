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
	"database/sql"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"unsafe"

	"github.com/rpadovani/sqlx-v2/internal/bind"
	"github.com/rpadovani/sqlx-v2/internal/reflectx"
)

// scanPool pools []any slices used for scanning rows.
// This avoids heap-thrashing from allocating new slices for each row.

var scanPool = sync.Pool{
	New: func() any {
		var s []any
		return &s
	},
}

// getScanDest returns a pooled []any slice of the given size.
func getScanDest(size int) []any {
	sp := scanPool.Get().(*[]any)
	s := *sp
	if cap(s) < size {
		s = make([]any, size)
	} else {
		s = s[:size]
	}
	return s
}

// putScanDest returns a []any slice to the pool.
func putScanDest(s []any) {
	s = s[:cap(s)]
	for i := range s {
		s[i] = nil
	}
	scanPool.Put(&s)
}

// populateScanDest fills the scanDest slice with pointers to fields in the struct
// based on the pre-calculated fieldMeta.
func populateScanDest(base unsafe.Pointer, meta []fieldMeta, scanDest []any) {
	for _, m := range meta {
		if m.traversal == nil {
			scanDest[m.colIndex] = new(any)
			continue
		}

		var ptr unsafe.Pointer
		if m.isPtrPath {
			ptr = reflectx.AddrByTraversal(base, m.traversal)
		} else {
			ptr = unsafe.Add(base, m.offset)
		}

		scanDest[m.colIndex] = reflect.NewAt(m.fieldType, ptr).Interface()
	}
}

// hashColumns computes a zero-allocation FNV-1a hash of the column names.
func hashColumns(cols []string) uint64 {
	var hash uint64 = 14695981039346656037
	for _, col := range cols {
		for i := 0; i < len(col); i++ {
			hash ^= uint64(col[i])
			hash *= 1099511628211
		}
	}
	return hash
}

type metaCacheKey struct {
	t      reflect.Type
	hash   uint64
	unsafe bool
}

type metaCacheValue struct {
	columns []string
	meta    []fieldMeta
}

var metaCache sync.Map

// buildMeta constructs the fieldMeta slice that maps each SQL column to its
// corresponding struct field. This is the single source of truth for the
// column→field mapping logic used by selectScan, getScan, Rows.StructScan,
// and all generic scan functions.
func buildMeta(t reflect.Type, columns []string, tm *reflectx.StructMap, isUnsafe bool) ([]fieldMeta, error) {
	key := metaCacheKey{
		t:      t,
		hash:   hashColumns(columns),
		unsafe: isUnsafe,
	}

	if val, ok := metaCache.Load(key); ok {
		cached := val.(metaCacheValue)
		if len(cached.columns) == len(columns) {
			match := true
			for i := range columns {
				if cached.columns[i] != columns[i] {
					match = false
					break
				}
			}
			if match {
				return cached.meta, nil
			}
		}
	}

	meta := make([]fieldMeta, len(columns))
	for i, col := range columns {
		if fi, ok := tm.Names[col]; ok {
			meta[i] = fieldMeta{
				traversal: fi.Traversal,
				fieldType: fi.TargetType,
				offset:    fi.Offset,
				isPtrPath: fi.IsPtrPath,
				colIndex:  i,
			}
		} else if !isUnsafe {
			return nil, fmt.Errorf("sqlx: missing destination name %q in struct %s: %w", col, t.Name(), bind.ErrColumnNotFound)
		}
	}

	colsCopy := make([]string, len(columns))
	copy(colsCopy, columns)

	// Sort meta by offset for L1 cache sequential access in the hot loop
	importSort := func(m []fieldMeta) {
		// simple insertion sort since n is usually small (columns < 100)
		for i := 1; i < len(m); i++ {
			j := i
			for j > 0 && m[j-1].offset < m[j].offset {
				m[j-1], m[j] = m[j], m[j-1]
				j--
			}
		}
	}
	importSort(meta)

	metaCache.Store(key, metaCacheValue{
		columns: colsCopy,
		meta:    meta,
	})

	return meta, nil
}

// isScannable returns true if the given type is directly scannable by database/sql.
// Types like string, int, float64, sql.NullString, etc. are scannable.
func isScannable(t reflect.Type) bool {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String, reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return true
		}
	}

	scannerType := reflect.TypeFor[sql.Scanner]()
	if t.Implements(scannerType) || reflect.PointerTo(t).Implements(scannerType) {
		return true
	}

	return false
}

var ErrStopIter = fmt.Errorf("stop iteration")

// iterateScan is the unified generic row processing engine.
// It iterates over `rows`, calling `alloc` to get a reflect.Value for the destination pointer,
// scans into it, and then calls `yield`. If `yield` returns ErrStopIter, iteration halts without an error.
func iterateScan(rows *sql.Rows, elemType reflect.Type, isUnsafe, strictTagParsing bool, mapper *reflectx.Mapper, alloc func() reflect.Value, yield func(reflect.Value) error) (err error) {
	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("sqlx: error getting columns: %w", err)
	}

	if isScannable(elemType) {
		if len(columns) != 1 {
			return fmt.Errorf("sqlx: scannable dest type %s with >1 columns (%d)", elemType, len(columns))
		}
		for rows.Next() {
			err := func() error {
				vp := alloc()
				defer runtime.KeepAlive(vp)
				if err := rows.Scan(vp.Interface()); err != nil {
					return fmt.Errorf("sqlx: scannable scan error: %w", err)
				}
				return yield(vp)
			}()
			if err != nil {
				if err == ErrStopIter {
					return nil
				}
				return err
			}
		}
		return rows.Err()
	}

	tm := mapper.TypeMap(elemType)

	if strictTagParsing && len(tm.Errors) > 0 {
		return fmt.Errorf("sqlx: strict tag error: %w", tm.Errors[0])
	}

	meta, err := buildMeta(elemType, columns, tm, isUnsafe)
	if err != nil {
		return fmt.Errorf("sqlx: error building meta: %w", err)
	}

	scanDest := getScanDest(len(columns))
	defer putScanDest(scanDest)

	for rows.Next() {
		err := func() error {
			vp := alloc()
			defer runtime.KeepAlive(vp)
			base := unsafe.Pointer(vp.Pointer())

			populateScanDest(base, meta, scanDest)

			if err := rows.Scan(scanDest...); err != nil {
				return fmt.Errorf("sqlx: struct scan error: %w", err)
			}

			return yield(vp)
		}()
		if err != nil {
			if err == ErrStopIter {
				return nil
			}
			return err
		}
	}

	return rows.Err()
}

// selectScan is the core implementation of Select. It scans rows into a slice of structs.
func selectScan(rows *sql.Rows, dest any, isUnsafe, strictTagParsing bool, mapper *reflectx.Mapper) error {
	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Pointer || dv.Elem().Kind() != reflect.Slice {
		if cerr := rows.Close(); cerr != nil {
			return fmt.Errorf("sqlx: dest must be a pointer to a slice, got %T; close error: %v", dest, cerr)
		}
		return fmt.Errorf("sqlx: dest must be a pointer to a slice, got %T", dest)
	}

	slice := dv.Elem()
	elemType := slice.Type().Elem()
	isPtr := elemType.Kind() == reflect.Pointer
	if isPtr {
		elemType = elemType.Elem()
	}

	alloc := func() reflect.Value { return reflect.New(elemType) }

	return iterateScan(rows, elemType, isUnsafe, strictTagParsing, mapper, alloc, func(vp reflect.Value) error {
		if isPtr {
			slice.Set(reflect.Append(slice, vp))
		} else {
			slice.Set(reflect.Append(slice, vp.Elem()))
		}
		return nil
	})
}

// getScan is the core implementation of Get. It scans a single row into a dest.
func getScan(rows *sql.Rows, dest any, isUnsafe, strictTagParsing bool, mapper *reflectx.Mapper) error {
	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Pointer {
		if cerr := rows.Close(); cerr != nil {
			return fmt.Errorf("sqlx: dest must be a pointer, got %T; close error: %v", dest, cerr)
		}
		return fmt.Errorf("sqlx: dest must be a pointer, got %T", dest)
	}

	elemType := dv.Type().Elem()
	alloc := func() reflect.Value { return dv }
	found := false

	err := iterateScan(rows, elemType, isUnsafe, strictTagParsing, mapper, alloc, func(vp reflect.Value) error {
		found = true
		return ErrStopIter
	})

	if err != nil {
		return err
	}
	if !found {
		return sql.ErrNoRows
	}
	return nil
}
