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
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"unsafe"

	"github.com/rpadovani/sqlx-v2/internal/bind"
	"github.com/rpadovani/sqlx-v2/internal/reflectx"
)

var defaultMapper = reflectx.NewMapperFunc("db", NameMapper)

// Select executes a query using the provided Queryer, and StructScans
// each row into dest, which must be a slice.
func Select(q Queryer, dest any, query string, args ...any) error {
	rows, err := q.Query(query, args...)
	if err != nil {
		return fmt.Errorf("sqlx: error in Select query: %w", err)
	}
	isUnsafe, strictTagParsing, mapper := extractMeta(q)
	return selectScan(rows, dest, isUnsafe, strictTagParsing, mapper)
}

// SelectContext executes a query using the provided QueryerContext,
// and StructScans each row into dest, which must be a slice.
func SelectContext(ctx context.Context, q QueryerContext, dest any, query string, args ...any) error {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("sqlx: error in SelectContext query: %w", err)
	}
	isUnsafe, strictTagParsing, mapper := extractMeta(q)
	return selectScan(rows, dest, isUnsafe, strictTagParsing, mapper)
}

// Get does a QueryRow using the provided Queryer, and scans the
// resulting row to dest. If dest is scannable, the result must
// only have one column. Otherwise, StructScan is used.
func Get(q Queryer, dest any, query string, args ...any) error {
	rows, err := q.Query(query, args...)
	if err != nil {
		return fmt.Errorf("sqlx: error in Get query: %w", err)
	}
	isUnsafe, strictTagParsing, mapper := extractMeta(q)
	return getScan(rows, dest, isUnsafe, strictTagParsing, mapper)
}

// GetContext does a QueryRowContext using the provided QueryerContext,
// and scans the resulting row to dest.
func GetContext(ctx context.Context, q QueryerContext, dest any, query string, args ...any) error {
	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("sqlx: error in GetContext query: %w", err)
	}
	isUnsafe, strictTagParsing, mapper := extractMeta(q)
	return getScan(rows, dest, isUnsafe, strictTagParsing, mapper)
}

// LoadFile exec's every statement in a file. Used for initializing schemas.
func LoadFile(e Execer, path string) (*sql.Result, error) {
	realpath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	contents, err := os.ReadFile(realpath)
	if err != nil {
		return nil, err
	}
	res, err := e.Exec(string(contents))
	return &res, err
}

// LoadFileContext exec's every statement in a file with context.
func LoadFileContext(ctx context.Context, e ExecerContext, path string) (*sql.Result, error) {
	realpath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	contents, err := os.ReadFile(realpath)
	if err != nil {
		return nil, err
	}
	res, err := e.ExecContext(ctx, string(contents))
	return &res, err
}

// MustExec execs the query using e and panics if there was an error.
func MustExec(e Execer, query string, args ...any) sql.Result {
	res, err := e.Exec(query, args...)
	if err != nil {
		panic(err)
	}
	return res
}

// MustExecContext execs the query using e and panics if there was an error.
func MustExecContext(ctx context.Context, e ExecerContext, query string, args ...any) sql.Result {
	res, err := e.ExecContext(ctx, query, args...)
	if err != nil {
		panic(err)
	}
	return res
}

// NamedQuery binds a named query and then runs Query on the result using the
// provided Ext (which is typically a *DB or *Tx).
func NamedQuery(e Ext, query string, arg any) (*Rows, error) {
	q, args, err := bind.BindNamed(bind.BindType(extractDriverName(e)), query, arg)
	if err != nil {
		return nil, err
	}
	return e.Queryx(q, args...)
}

// NamedQueryContext binds a named query and then runs Query on the result using the
// provided ExtContext.
func NamedQueryContext(ctx context.Context, e ExtContext, query string, arg any) (*Rows, error) {
	q, args, err := bind.BindNamed(bind.BindType(extractDriverName(e)), query, arg)
	if err != nil {
		return nil, err
	}
	return e.QueryxContext(ctx, q, args...)
}

// NamedExec uses BindNamed to bind a query and then runs Exec on the result.
func NamedExec(e Ext, query string, arg any) (sql.Result, error) {
	q, args, err := bind.BindNamed(bind.BindType(extractDriverName(e)), query, arg)
	if err != nil {
		return nil, err
	}
	return e.Exec(q, args...)
}

// NamedExecContext uses BindNamed to bind a query and then runs Exec on the result.
func NamedExecContext(ctx context.Context, e ExtContext, query string, arg any) (sql.Result, error) {
	q, args, err := bind.BindNamed(bind.BindType(extractDriverName(e)), query, arg)
	if err != nil {
		return nil, err
	}
	return e.ExecContext(ctx, q, args...)
}

// MapScan scans a single row into a map[string]any.
func MapScan(r ColScanner, dest map[string]any) error {
	columns, err := r.Columns()
	if err != nil {
		return err
	}

	values := make([]any, len(columns))
	valPtrs := make([]any, len(columns))
	for i := range values {
		valPtrs[i] = &values[i]
	}

	err = r.Scan(valPtrs...)
	if err != nil {
		return err
	}

	for i, col := range columns {
		dest[col] = values[i]
	}
	return nil
}

// SliceScan using this Rows.
func SliceScan(r ColScanner) ([]any, error) {
	columns, err := r.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]any, len(columns))
	valPtrs := make([]any, len(columns))
	for i := range values {
		valPtrs[i] = &values[i]
	}

	err = r.Scan(valPtrs...)
	if err != nil {
		return nil, err
	}
	return values, nil
}

// StructScan scans a ColScanner (Row or Rows) into a struct.
// It respects the unsafe flag from Row or Rows; if not in unsafe mode,
// unmapped columns will cause an error.
func StructScan(r ColScanner, dest any) error {
	columns, err := r.Columns()
	if err != nil {
		return err
	}

	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Pointer || dv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("sqlx: dest must be a pointer to a struct, got %T", dest)
	}

	elemType := dv.Elem().Type()

	var isUnsafe bool
	var strictTagParsing bool
	mapper := defaultMapper
	switch v := r.(type) {
	case *Row:
		isUnsafe = v.Unsafe
		strictTagParsing = v.StrictTagParsing
		if v.Mapper != nil {
			mapper = v.Mapper
		}
	case *Rows:
		isUnsafe = v.Unsafe
		strictTagParsing = v.StrictTagParsing
		if v.Mapper != nil {
			mapper = v.Mapper
		}
	}

	tm := mapper.TypeMap(elemType)

	if strictTagParsing && len(tm.Errors) > 0 {
		return fmt.Errorf("sqlx: strict tag error: %w", tm.Errors[0])
	}

	vp := reflect.ValueOf(dest)
	base := unsafe.Pointer(vp.Pointer())

	scanDest := getScanDest(len(columns))
	defer putScanDest(scanDest)

	for i, col := range columns {
		if fi, ok := tm.Names[col]; ok {
			if fi.IsPtrPath {
				ptr := reflectx.AddrByTraversal(base, fi.Traversal)
				scanDest[i] = reflect.NewAt(fi.TargetType, ptr).Interface()
			} else {
				ptr := unsafe.Add(base, fi.Offset)
				scanDest[i] = reflect.NewAt(fi.TargetType, ptr).Interface()
			}
		} else if !isUnsafe {
			return fmt.Errorf("sqlx: missing destination name %q in %s: %w", col, elemType, bind.ErrColumnNotFound)
		} else {
			scanDest[i] = new(any)
		}
	}

	err = r.Scan(scanDest...)
	runtime.KeepAlive(vp)
	if err != nil {
		return fmt.Errorf("sqlx: struct scan error: %w", err)
	}
	return nil
}

// extractMeta safely extracts configuration from known types, or returns defaults.
func extractMeta(q any) (bool, bool, *reflectx.Mapper) {
	switch v := q.(type) {
	case *DB:
		if v.Mapper == nil {
			return v.Unsafe, v.StrictTagParsing, defaultMapper
		}
		return v.Unsafe, v.StrictTagParsing, v.Mapper
	case *Tx:
		if v.Mapper == nil {
			return v.Unsafe, v.StrictTagParsing, defaultMapper
		}
		return v.Unsafe, v.StrictTagParsing, v.Mapper
	case *Conn:
		if v.Mapper == nil {
			return v.Unsafe, v.StrictTagParsing, defaultMapper
		}
		return v.Unsafe, v.StrictTagParsing, v.Mapper
	default:
		return false, false, defaultMapper
	}
}

// extractDriverName attempts to locate the driverName for bind parsing rules.
func extractDriverName(e any) string {
	switch v := e.(type) {
	case *DB:
		return v.DriverName
	case *Tx:
		return v.DriverName
	case *Conn:
		return v.DriverName
	default:
		return "unknown"
	}
}
