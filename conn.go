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
	"github.com/rpadovani/sqlx-v2/internal/reflectx"
)

// Conn wraps sql.Conn with extra functionality.
type Conn struct {
	*sql.Conn
	DriverName       string
	Mapper           *reflectx.Mapper
	Unsafe           bool
	StrictTagParsing bool
}

// Rebind a query within a Conn's bindvar type.
func (c *Conn) Rebind(query string) string {
	return bind.Rebind(bind.BindType(c.DriverName), query)
}

// SelectContext using this Conn.
func (c *Conn) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return SelectContext(ctx, c, dest, query, args...)
}

// GetContext using this Conn.
func (c *Conn) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return GetContext(ctx, c, dest, query, args...)
}

// QueryxContext queries the database and returns an *sqlx.Rows.
func (c *Conn) QueryxContext(ctx context.Context, query string, args ...any) (*Rows, error) {
	r, err := c.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{Rows: r, Mapper: c.Mapper, Unsafe: c.Unsafe, StrictTagParsing: c.StrictTagParsing}, nil
}

// QueryRowxContext queries the database and returns an *sqlx.Row.
func (c *Conn) QueryRowxContext(ctx context.Context, query string, args ...any) *Row {
	rows, err := c.QueryContext(ctx, query, args...)
	return &Row{rows: rows, err: err, Mapper: c.Mapper, Unsafe: c.Unsafe, StrictTagParsing: c.StrictTagParsing}
}

// PreparexContext returns an sqlx.Stmt instead of a sql.Stmt.
func (c *Conn) PreparexContext(ctx context.Context, query string) (*Stmt, error) {
	return PreparexContext(ctx, c, query)
}

// BeginTxx begins a transaction and returns an *sqlx.Tx.
func (c *Conn) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := c.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: tx, DriverName: c.DriverName, Mapper: c.Mapper, Unsafe: c.Unsafe, StrictTagParsing: c.StrictTagParsing}, nil
}
