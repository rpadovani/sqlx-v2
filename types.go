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

	"github.com/rpadovani/sqlx-v2/internal/bind"
)

// Bindvar type constants. These are re-exported from the internal bind
// package to maintain API compatibility with jmoiron/sqlx.
const (
	UNKNOWN  = bind.UNKNOWN
	QUESTION = bind.QUESTION
	DOLLAR   = bind.DOLLAR
	NAMED    = bind.NAMED
	AT       = bind.AT
)

// BindType returns the bindtype for a given database given a drivername.
func BindType(driverName string) int {
	return bind.BindType(driverName)
}

// BindDriver sets the BindType for driverName to bindType.
func BindDriver(driverName string, bindType int) {
	bind.BindDriver(driverName, bindType)
}

// ColScanner defines the interface for MapScan and SliceScan.
type ColScanner interface {
	Columns() ([]string, error)
	Scan(dest ...any) error
	Err() error
}

// Queryer defines the interface for Get and Select.
type Queryer interface {
	Query(query string, args ...any) (*sql.Rows, error)
	Queryx(query string, args ...any) (*Rows, error)
	QueryRowx(query string, args ...any) *Row
}

// QueryerContext defines the interface for GetContext and SelectContext.
type QueryerContext interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryxContext(ctx context.Context, query string, args ...any) (*Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...any) *Row
}

// Execer defines the interface for MustExec and LoadFile.
type Execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

// ExecerContext defines the interface for MustExecContext and LoadFileContext.
type ExecerContext interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Ext defines a union interface which can Bind, Query, and Exec.
type Ext interface {
	Queryer
	Execer
}

// ExtContext defines a union interface which can Bind, Query, and Exec with context.
type ExtContext interface {
	QueryerContext
	ExecerContext
}

// Preparer defines the interface for Preparex.
type Preparer interface {
	Prepare(query string) (*sql.Stmt, error)
}

// PreparerContext defines the interface for PreparexContext.
type PreparerContext interface {
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}
