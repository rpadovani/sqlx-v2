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
	"unsafe"

	"github.com/rpadovani/sqlx-v2/internal/reflectx"
)

// Row reimplements sql.Row to provide access to the underlying sql.Rows object.
type Row struct {
	err              error
	rows             *sql.Rows
	Mapper           *reflectx.Mapper
	Unsafe           bool
	StrictTagParsing bool
}

// Scan implements sql.Row.Scan without discarding underlying errors.
func (r *Row) Scan(dest ...any) (err error) {
	if r.err != nil {
		return r.err
	}
	defer func() {
		if cerr := r.rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	for _, dp := range dest {
		if _, ok := dp.(*sql.RawBytes); ok {
			return fmt.Errorf("sql: RawBytes isn't allowed on Row.Scan")
		}
	}
	if !r.rows.Next() {
		if err = r.rows.Err(); err != nil {
			return err
		}
		return sql.ErrNoRows
	}
	if err = r.rows.Scan(dest...); err != nil {
		return err
	}
	// Make sure the query can be processed to completion with no errors.
	return nil
}

// Columns returns the underlying sql.Rows.Columns(), or the deferred error.
func (r *Row) Columns() ([]string, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.rows.Columns()
}

// ColumnTypes returns the underlying sql.Rows.ColumnTypes(), or the deferred error.
func (r *Row) ColumnTypes() ([]*sql.ColumnType, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.rows.ColumnTypes()
}

// Err returns the error encountered, if any.
func (r *Row) Err() error {
	return r.err
}

// SliceScan a row, returning a []any with values similar to MapScan.
func (r *Row) SliceScan() ([]any, error) {
	return SliceScan(r)
}

// MapScan scans a single Row into the dest map[string]any.
func (r *Row) MapScan(dest map[string]any) error {
	return MapScan(r, dest)
}

// StructScan a single Row into dest.
func (r *Row) StructScan(dest any) error {
	return StructScan(r, dest)
}

// Rows wraps sql.Rows to cache costly reflection operations.
type Rows struct {
	*sql.Rows
	Mapper           *reflectx.Mapper
	Unsafe           bool
	StrictTagParsing bool
	// started, meta, and baseType are used for StructScan caching
	started  bool
	meta     []fieldMeta
	scanDest []any
	baseType reflect.Type
}

type fieldMeta struct {
	traversal []reflectx.Step
	fieldType reflect.Type
	offset    uintptr
	isPtrPath bool
	colIndex  int
}

// SliceScan a row, returning a []any with values similar to MapScan.
func (r *Rows) SliceScan() ([]any, error) {
	return SliceScan(r)
}

// MapScan scans a single Row from Rows into the dest map[string]any.
func (r *Rows) MapScan(dest map[string]any) error {
	return MapScan(r, dest)
}

// StructScan scans the current row into dest, which must be a pointer to a struct.
// Field metadata is cached after the first call; subsequent calls must use the same type.
func (r *Rows) StructScan(dest any) error {
	dv := reflect.ValueOf(dest)
	if dv.Kind() != reflect.Pointer || dv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("sqlx: dest must be a pointer to a struct, got %T", dest)
	}

	elemType := dv.Elem().Type()

	if !r.started {
		columns, err := r.Columns()
		if err != nil {
			return err
		}

		mapper := r.Mapper
		if mapper == nil {
			mapper = defaultMapper
			r.Mapper = mapper
		}

		tm := mapper.TypeMap(elemType)

		meta, err := buildMeta(elemType, columns, tm, r.Unsafe)
		if err != nil {
			return fmt.Errorf("sqlx: error building meta: %w", err)
		}
		r.meta = meta
		r.scanDest = make([]any, len(meta))

		r.started = true
		r.baseType = elemType
	} else if r.baseType != elemType {
		return fmt.Errorf("sqlx: StructScan called with structurally different type than previous iteration (expected %s, got %s)", r.baseType, elemType)
	}

	vp := reflect.ValueOf(dest)
	base := unsafe.Pointer(vp.Pointer())

	populateScanDest(base, r.meta, r.scanDest)

	defer runtime.KeepAlive(vp)
	return r.Scan(r.scanDest...)
}
