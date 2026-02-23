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

	"github.com/rpadovani/sqlx-v2/internal/bind"
	"github.com/rpadovani/sqlx-v2/internal/reflectx"
)

// DB wraps sql.DB with extra functionality.
type DB struct {
	*sql.DB
	DriverName       string
	Mapper           *reflectx.Mapper
	Unsafe           bool
	StrictTagParsing bool
}

// NewDb returns a new sqlx DB wrapper for a pre-existing *sql.DB.
// The driverName of the original database is required for named query support.
func NewDb(db *sql.DB, driverName string) *DB {
	return &DB{
		DB:         db,
		DriverName: driverName,
		Mapper:     reflectx.NewMapperFunc("db", NameMapper),
	}
}

// Connect to a database and verify with a ping.
func Connect(driverName, dataSourceName string) (*DB, error) {
	db, err := Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		if cerr := db.Close(); cerr != nil {
			return nil, fmt.Errorf("sqlx: error pinging database: %w (close error: %v)", err, cerr)
		}
		return nil, fmt.Errorf("sqlx: error pinging database: %w", err)
	}
	return db, nil
}

// ConnectContext to a database and verify with a ping.
func ConnectContext(ctx context.Context, driverName, dataSourceName string) (*DB, error) {
	db, err := Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	err = db.PingContext(ctx)
	if err != nil {
		if cerr := db.Close(); cerr != nil {
			return nil, fmt.Errorf("sqlx: error pinging database: %w (close error: %v)", err, cerr)
		}
		return nil, err
	}
	return db, nil
}

// MustConnect connects to a database and panics on error.
func MustConnect(driverName, dataSourceName string) *DB {
	db, err := Connect(driverName, dataSourceName)
	if err != nil {
		panic(err)
	}
	return db
}

// MustOpen is the same as sql.Open, but returns an *sqlx.DB instead and panics on error.
func MustOpen(driverName, dataSourceName string) *DB {
	db, err := Open(driverName, dataSourceName)
	if err != nil {
		panic(err)
	}
	return db
}

// Open is the same as sql.Open, but returns an *sqlx.DB instead.
func Open(driverName, dataSourceName string) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	return NewDb(db, driverName), nil
}

// MapperFunc sets a new mapper for this db using the default sqlx struct tag
// and the provided mapper function.
func (db *DB) MapperFunc(mf func(string) string) {
	db.Mapper = reflectx.NewMapperFunc("db", mf)
}

// Rebind transforms a query from QUESTION to the DB driver's bindvar type.
func (db *DB) Rebind(query string) string {
	return bind.Rebind(bind.BindType(db.DriverName), query)
}

// BindNamed binds a query using the DB driver's bindvar type.
func (db *DB) BindNamed(query string, arg any) (string, []any, error) {
	return bind.BindNamed(bind.BindType(db.DriverName), query, arg)
}

// NamedQuery using this DB.
// Any named placeholder parameters are replaced with fields from arg.
func (db *DB) NamedQuery(query string, arg any) (*Rows, error) {
	return db.NamedQueryContext(context.Background(), query, arg)
}

// NamedQueryContext using this DB.
func (db *DB) NamedQueryContext(ctx context.Context, query string, arg any) (*Rows, error) {
	return NamedQueryContext(ctx, db, query, arg)
}

// NamedExec using this DB.
// Any named placeholder parameters are replaced with fields from arg.
func (db *DB) NamedExec(query string, arg any) (sql.Result, error) {
	return db.NamedExecContext(context.Background(), query, arg)
}

// NamedExecContext using this DB.
func (db *DB) NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error) {
	return NamedExecContext(ctx, db, query, arg)
}

// Select using this DB.
// Any placeholder parameters are replaced with supplied args.
func (db *DB) Select(dest any, query string, args ...any) error {
	return SelectContext(context.Background(), db, dest, query, args...)
}

// SelectContext using this DB.
func (db *DB) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return SelectContext(ctx, db, dest, query, args...)
}

// Get using this DB.
// Any placeholder parameters are replaced with supplied args.
// An error is returned if the result set is empty.
func (db *DB) Get(dest any, query string, args ...any) error {
	return db.GetContext(context.Background(), dest, query, args...)
}

// GetContext using this DB.
func (db *DB) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	return GetContext(ctx, db, dest, query, args...)
}

// MustBegin starts a transaction, and panics on error.
func (db *DB) MustBegin() *Tx {
	tx, err := db.Beginx()
	if err != nil {
		panic(err)
	}
	return tx
}

// MustBeginTx starts a transaction with context, and panics on error.
func (db *DB) MustBeginTx(ctx context.Context, opts *sql.TxOptions) *Tx {
	tx, err := db.BeginTxx(ctx, opts)
	if err != nil {
		panic(err)
	}
	return tx
}

// Beginx begins a transaction and returns an *sqlx.Tx instead of an *sql.Tx.
func (db *DB) Beginx() (*Tx, error) {
	return db.BeginTxx(context.Background(), nil)
}

// BeginTxx begins a transaction and returns an *sqlx.Tx instead of an *sql.Tx.
func (db *DB) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.BeginTx(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("sqlx: error beginning transaction: %w", err)
	}
	return &Tx{Tx: tx, DriverName: db.DriverName, Mapper: db.Mapper, Unsafe: db.Unsafe, StrictTagParsing: db.StrictTagParsing}, nil
}

// Queryx queries the database and returns an *sqlx.Rows.
func (db *DB) Queryx(query string, args ...any) (*Rows, error) {
	return db.QueryxContext(context.Background(), query, args...)
}

// QueryxContext queries the database and returns an *sqlx.Rows.
func (db *DB) QueryxContext(ctx context.Context, query string, args ...any) (*Rows, error) {
	r, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("sqlx: error in query: %w", err)
	}
	return &Rows{Rows: r, Mapper: db.Mapper, Unsafe: db.Unsafe, StrictTagParsing: db.StrictTagParsing}, nil
}

// QueryRowx queries the database and returns an *sqlx.Row.
func (db *DB) QueryRowx(query string, args ...any) *Row {
	return db.QueryRowxContext(context.Background(), query, args...)
}

// QueryRowxContext queries the database and returns an *sqlx.Row.
func (db *DB) QueryRowxContext(ctx context.Context, query string, args ...any) *Row {
	rows, err := db.QueryContext(ctx, query, args...)
	return &Row{rows: rows, err: err, Mapper: db.Mapper, Unsafe: db.Unsafe, StrictTagParsing: db.StrictTagParsing}
}

// MustExec (panic) runs MustExec using this database.
func (db *DB) MustExec(query string, args ...any) sql.Result {
	return db.MustExecContext(context.Background(), query, args...)
}

// MustExecContext (panic) runs MustExecContext using this database.
func (db *DB) MustExecContext(ctx context.Context, query string, args ...any) sql.Result {
	return MustExecContext(ctx, db, query, args...)
}

// Preparex returns an sqlx.Stmt instead of a sql.Stmt.
func (db *DB) Preparex(query string) (*Stmt, error) {
	return db.PreparexContext(context.Background(), query)
}

// PreparexContext returns an sqlx.Stmt instead of a sql.Stmt.
func (db *DB) PreparexContext(ctx context.Context, query string) (*Stmt, error) {
	return PreparexContext(ctx, db, query)
}

// PrepareNamed returns an sqlx.NamedStmt.
func (db *DB) PrepareNamed(query string) (*NamedStmt, error) {
	return db.PrepareNamedContext(context.Background(), query)
}

// PrepareNamedContext returns an sqlx.NamedStmt.
func (db *DB) PrepareNamedContext(ctx context.Context, query string) (*NamedStmt, error) {
	bindType := bind.BindType(db.DriverName)
	q, args, err := bind.CompileNamedQuery([]byte(query), bindType)
	if err != nil {
		return nil, fmt.Errorf("sqlx: error compiling named query: %w", err)
	}
	stmt, err := db.PreparexContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("sqlx: error preparing statement: %w", err)
	}
	return &NamedStmt{
		QueryString:      q,
		Params:           args,
		Stmt:             stmt.Stmt,
		Mapper:           stmt.Mapper,
		Unsafe:           db.Unsafe,
		StrictTagParsing: db.StrictTagParsing,
	}, nil
}

// Connx returns an *sqlx.Conn instead of an *sql.Conn.
func (db *DB) Connx(ctx context.Context) (*Conn, error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("sqlx: error getting connection: %w", err)
	}
	return &Conn{Conn: conn, DriverName: db.DriverName, Mapper: db.Mapper, Unsafe: db.Unsafe, StrictTagParsing: db.StrictTagParsing}, nil
}
