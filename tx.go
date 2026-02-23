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
	"reflect"

	"github.com/rpadovani/sqlx-v2/internal/bind"
	"github.com/rpadovani/sqlx-v2/internal/reflectx"
)

// Tx wraps sql.Tx with extra functionality.
type Tx struct {
	*sql.Tx
	DriverName       string
	Mapper           *reflectx.Mapper
	Unsafe           bool
	StrictTagParsing bool
}

// Rebind a query within a transaction's bindvar type.
func (tx *Tx) Rebind(query string) string {
	return bind.Rebind(bind.BindType(tx.DriverName), query)
}

// BindNamed binds a query within this transaction's bindvar type.
func (tx *Tx) BindNamed(query string, arg any) (string, []any, error) {
	return bind.BindNamed(bind.BindType(tx.DriverName), query, arg)
}

// NamedQuery within this transaction.
func (tx *Tx) NamedQuery(query string, arg any) (*Rows, error) {
	return tx.NamedQueryContext(context.Background(), query, arg)
}

// NamedQueryContext within this transaction.
func (tx *Tx) NamedQueryContext(ctx context.Context, query string, arg any) (*Rows, error) {
	return NamedQueryContext(ctx, tx, query, arg)
}

// NamedExec a named query within this transaction.
func (tx *Tx) NamedExec(query string, arg any) (sql.Result, error) {
	return tx.NamedExecContext(context.Background(), query, arg)
}

// NamedExecContext a named query within this transaction.
func (tx *Tx) NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error) {
	return NamedExecContext(ctx, tx, query, arg)
}

// Select within this transaction.
func (tx *Tx) Select(dest any, query string, args ...any) error {
	return tx.SelectContext(context.Background(), dest, query, args...)
}

// SelectContext within this transaction.
func (tx *Tx) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return SelectContext(ctx, tx, dest, query, args...)
}

// Get within this transaction.
func (tx *Tx) Get(dest any, query string, args ...any) error {
	return tx.GetContext(context.Background(), dest, query, args...)
}

// GetContext within this transaction.
func (tx *Tx) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return GetContext(ctx, tx, dest, query, args...)
}

// MustExec runs MustExec within this transaction.
func (tx *Tx) MustExec(query string, args ...any) sql.Result {
	return tx.MustExecContext(context.Background(), query, args...)
}

// MustExecContext runs MustExecContext within this transaction.
func (tx *Tx) MustExecContext(ctx context.Context, query string, args ...any) sql.Result {
	return MustExecContext(ctx, tx, query, args...)
}

// Queryx within this transaction.
func (tx *Tx) Queryx(query string, args ...any) (*Rows, error) {
	return tx.QueryxContext(context.Background(), query, args...)
}

// QueryxContext within this transaction.
func (tx *Tx) QueryxContext(ctx context.Context, query string, args ...any) (*Rows, error) {
	r, err := tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{Rows: r, Mapper: tx.Mapper, Unsafe: tx.Unsafe, StrictTagParsing: tx.StrictTagParsing}, nil
}

// QueryRowx within this transaction.
func (tx *Tx) QueryRowx(query string, args ...any) *Row {
	return tx.QueryRowxContext(context.Background(), query, args...)
}

// QueryRowxContext within this transaction.
func (tx *Tx) QueryRowxContext(ctx context.Context, query string, args ...any) *Row {
	rows, err := tx.QueryContext(ctx, query, args...)
	return &Row{rows: rows, err: err, Mapper: tx.Mapper, Unsafe: tx.Unsafe, StrictTagParsing: tx.StrictTagParsing}
}

// Preparex returns an sqlx.Stmt instead of a sql.Stmt.
func (tx *Tx) Preparex(query string) (*Stmt, error) {
	return tx.PreparexContext(context.Background(), query)
}

// PreparexContext returns an sqlx.Stmt instead of a sql.Stmt.
func (tx *Tx) PreparexContext(ctx context.Context, query string) (*Stmt, error) {
	return PreparexContext(ctx, tx, query)
}

// PrepareNamed returns an sqlx.NamedStmt.
func (tx *Tx) PrepareNamed(query string) (*NamedStmt, error) {
	return tx.PrepareNamedContext(context.Background(), query)
}

// PrepareNamedContext returns an sqlx.NamedStmt.
func (tx *Tx) PrepareNamedContext(ctx context.Context, query string) (*NamedStmt, error) {
	bindType := bind.BindType(tx.DriverName)
	q, args, err := bind.CompileNamedQuery([]byte(query), bindType)
	if err != nil {
		return nil, err
	}
	stmt, err := tx.PreparexContext(ctx, q)
	if err != nil {
		return nil, err
	}
	return &NamedStmt{
		QueryString:      q,
		Params:           args,
		Stmt:             stmt.Stmt,
		Mapper:           stmt.Mapper,
		Unsafe:           tx.Unsafe,
		StrictTagParsing: tx.StrictTagParsing,
	}, nil
}

// Stmtx returns a version of the prepared statement which runs within a transaction.
func (tx *Tx) Stmtx(stmt any) *Stmt {
	var s *sql.Stmt
	switch v := stmt.(type) {
	case Stmt:
		s = v.Stmt
	case *Stmt:
		s = v.Stmt
	case *sql.Stmt:
		s = v
	default:
		panic(fmt.Sprintf("non-statement type %v passed to Stmtx", reflect.ValueOf(stmt).Type()))
	}
	return &Stmt{Stmt: tx.Stmt(s), Mapper: tx.Mapper, Unsafe: tx.Unsafe, StrictTagParsing: tx.StrictTagParsing}
}

// StmtxContext returns a version of the prepared statement which runs within a transaction.
func (tx *Tx) StmtxContext(ctx context.Context, stmt any) *Stmt {
	var s *sql.Stmt
	switch v := stmt.(type) {
	case Stmt:
		s = v.Stmt
	case *Stmt:
		s = v.Stmt
	case *sql.Stmt:
		s = v
	default:
		panic(fmt.Sprintf("non-statement type %v passed to StmtxContext", reflect.ValueOf(stmt).Type()))
	}
	return &Stmt{Stmt: tx.StmtContext(ctx, s), Mapper: tx.Mapper, Unsafe: tx.Unsafe, StrictTagParsing: tx.StrictTagParsing}
}

// NamedStmt returns a version of the prepared statement which runs within a transaction.
func (tx *Tx) NamedStmt(stmt *NamedStmt) *NamedStmt {
	return &NamedStmt{
		QueryString:      stmt.QueryString,
		Params:           stmt.Params,
		Stmt:             tx.Stmt(stmt.Stmt),
		Mapper:           tx.Mapper,
		Unsafe:           tx.Unsafe,
		StrictTagParsing: tx.StrictTagParsing,
	}
}

// NamedStmtContext returns a version of the prepared statement which runs within a transaction.
func (tx *Tx) NamedStmtContext(ctx context.Context, stmt *NamedStmt) *NamedStmt {
	return &NamedStmt{
		QueryString:      stmt.QueryString,
		Params:           stmt.Params,
		Stmt:             tx.StmtContext(ctx, stmt.Stmt),
		Mapper:           tx.Mapper,
		Unsafe:           tx.Unsafe,
		StrictTagParsing: tx.StrictTagParsing,
	}
}
