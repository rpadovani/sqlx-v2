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
	"reflect"

	"github.com/rpadovani/sqlx-v2/internal/bind"
	"github.com/rpadovani/sqlx-v2/internal/reflectx"
)

// NamedStmt is a prepared statement that executes named queries.
// Prepare it how you would execute a NamedQuery, but pass in a struct or
// map when executing.
type NamedStmt struct {
	Stmt             *sql.Stmt
	QueryString      string
	Params           []string
	Mapper           *reflectx.Mapper
	Unsafe           bool
	StrictTagParsing bool
}

// Close closes the named statement.
func (n *NamedStmt) Close() error {
	return n.Stmt.Close()
}

// Exec executes a named statement using the struct passed.
// If the struct is a slice or array, it iterates and executes for each element.
func (n *NamedStmt) Exec(arg any) (sql.Result, error) {
	return n.ExecContext(context.Background(), arg)
}

// ExecContext executes a named statement using the struct passed.
// If the struct is a slice or array, it iterates and executes for each element.
func (n *NamedStmt) ExecContext(ctx context.Context, arg any) (sql.Result, error) {
	v := reflect.ValueOf(arg)
	if v.Kind() == reflect.Array || v.Kind() == reflect.Slice {
		if v.Len() == 0 {
			return nil, nil
		}
		var res sql.Result
		var err error
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i).Interface()
			args, errBind := bind.BindAnyArgs(n.Params, elem, n.Mapper)
			if errBind != nil {
				return res, errBind
			}
			res, err = n.Stmt.ExecContext(ctx, args...)
			if err != nil {
				return res, err
			}
		}
		return res, nil
	}

	args, err := bind.BindAnyArgs(n.Params, arg, n.Mapper)
	if err != nil {
		return nil, err
	}
	return n.Stmt.ExecContext(ctx, args...)
}

// Query executes a named statement using the struct passed, returns *sql.Rows.
func (n *NamedStmt) Query(arg any) (*sql.Rows, error) {
	return n.QueryContext(context.Background(), arg)
}

// QueryContext executes a named statement using the struct passed, returns *sql.Rows.
func (n *NamedStmt) QueryContext(ctx context.Context, arg any) (*sql.Rows, error) {
	args, err := bind.BindAnyArgs(n.Params, arg, n.Mapper)
	if err != nil {
		return nil, err
	}
	return n.Stmt.QueryContext(ctx, args...)
}

// QueryRow executes a named statement using the struct passed, returns *sqlx.Row.
func (n *NamedStmt) QueryRow(arg any) *Row {
	return n.QueryRowContext(context.Background(), arg)
}

// QueryRowContext executes a named statement returning an *sqlx.Row.
func (n *NamedStmt) QueryRowContext(ctx context.Context, arg any) *Row {
	args, err := bind.BindAnyArgs(n.Params, arg, n.Mapper)
	if err != nil {
		return &Row{err: err}
	}
	qs, err := n.Stmt.QueryContext(ctx, args...)
	return &Row{rows: qs, err: err, Mapper: n.Mapper, Unsafe: n.Unsafe, StrictTagParsing: n.StrictTagParsing}
}

// QueryRowx executes a named statement using the struct passed, returns *sqlx.Row.
func (n *NamedStmt) QueryRowx(arg any) *Row {
	return n.QueryRow(arg)
}

// QueryRowxContext executes a named statement returning an *sqlx.Row.
func (n *NamedStmt) QueryRowxContext(ctx context.Context, arg any) *Row {
	return n.QueryRowContext(ctx, arg)
}

// Queryx executes a named statement returning an *sqlx.Rows.
func (n *NamedStmt) Queryx(arg any) (*Rows, error) {
	return n.QueryxContext(context.Background(), arg)
}

// QueryxContext executes a named statement returning an *sqlx.Rows.
func (n *NamedStmt) QueryxContext(ctx context.Context, arg any) (*Rows, error) {
	r, err := n.QueryContext(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &Rows{Rows: r, Mapper: n.Mapper, Unsafe: n.Unsafe, StrictTagParsing: n.StrictTagParsing}, err
}

// MustExec (panic) runs MustExec using this NamedStmt.
func (n *NamedStmt) MustExec(arg any) sql.Result {
	return n.MustExecContext(context.Background(), arg)
}

// MustExecContext (panic) runs MustExecContext using this NamedStmt.
func (n *NamedStmt) MustExecContext(ctx context.Context, arg any) sql.Result {
	res, err := n.ExecContext(ctx, arg)
	if err != nil {
		panic(err)
	}
	return res
}

// Select using this NamedStmt.
func (n *NamedStmt) Select(dest any, arg any) error {
	return n.SelectContext(context.Background(), dest, arg)
}

// SelectContext using this NamedStmt.
func (n *NamedStmt) SelectContext(ctx context.Context, dest any, arg any) error {
	args, err := bind.BindAnyArgs(n.Params, arg, n.Mapper)
	if err != nil {
		return err
	}
	rows, err := n.Stmt.QueryContext(ctx, args...)
	if err != nil {
		return err
	}
	return selectScan(rows, dest, n.Unsafe, n.StrictTagParsing, n.Mapper)
}

// Get using this NamedStmt.
func (n *NamedStmt) Get(dest any, arg any) error {
	return n.GetContext(context.Background(), dest, arg)
}

// GetContext using this NamedStmt.
func (n *NamedStmt) GetContext(ctx context.Context, dest any, arg any) error {
	args, err := bind.BindAnyArgs(n.Params, arg, n.Mapper)
	if err != nil {
		return err
	}
	rows, err := n.Stmt.QueryContext(ctx, args...)
	if err != nil {
		return err
	}
	return getScan(rows, dest, n.Unsafe, n.StrictTagParsing, n.Mapper)
}
