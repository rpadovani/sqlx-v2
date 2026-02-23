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

	"github.com/rpadovani/sqlx-v2/internal/reflectx"
)

// Stmt wraps sql.Stmt with extra functionality.
type Stmt struct {
	*sql.Stmt
	Mapper           *reflectx.Mapper
	Unsafe           bool
	StrictTagParsing bool
}

// Select using this prepared statement.
func (s *Stmt) Select(dest any, args ...any) error {
	return s.SelectContext(context.Background(), dest, args...)
}

// SelectContext using this prepared statement.
func (s *Stmt) SelectContext(ctx context.Context, dest any, args ...any) error {
	rows, err := s.QueryContext(ctx, args...)
	if err != nil {
		return err
	}
	return selectScan(rows, dest, s.Unsafe, s.StrictTagParsing, s.Mapper)
}

// Get using this prepared statement.
func (s *Stmt) Get(dest any, args ...any) error {
	return s.GetContext(context.Background(), dest, args...)
}

// GetContext using this prepared statement.
func (s *Stmt) GetContext(ctx context.Context, dest any, args ...any) error {
	rows, err := s.QueryContext(ctx, args...)
	if err != nil {
		return err
	}
	return getScan(rows, dest, s.Unsafe, s.StrictTagParsing, s.Mapper)
}

// MustExec (panic) runs MustExec using this statement.
func (s *Stmt) MustExec(args ...any) sql.Result {
	return s.MustExecContext(context.Background(), args...)
}

// MustExecContext (panic) runs MustExecContext using this statement.
func (s *Stmt) MustExecContext(ctx context.Context, args ...any) sql.Result {
	res, err := s.ExecContext(ctx, args...)
	if err != nil {
		panic(err)
	}
	return res
}

// QueryRowx using this prepared statement.
func (s *Stmt) QueryRowx(args ...any) *Row {
	return s.QueryRowxContext(context.Background(), args...)
}

// QueryRowxContext using this prepared statement.
func (s *Stmt) QueryRowxContext(ctx context.Context, args ...any) *Row {
	qs, err := s.QueryContext(ctx, args...)
	return &Row{rows: qs, err: err, Mapper: s.Mapper, Unsafe: s.Unsafe, StrictTagParsing: s.StrictTagParsing}
}

// Queryx using this prepared statement.
func (s *Stmt) Queryx(args ...any) (*Rows, error) {
	return s.QueryxContext(context.Background(), args...)
}

// QueryxContext using this prepared statement.
func (s *Stmt) QueryxContext(ctx context.Context, args ...any) (*Rows, error) {
	r, err := s.QueryContext(ctx, args...)
	if err != nil {
		return nil, err
	}
	return &Rows{Rows: r, Mapper: s.Mapper, Unsafe: s.Unsafe, StrictTagParsing: s.StrictTagParsing}, nil
}

// Preparex prepares a statement and returns an *sqlx.Stmt.
func Preparex(p Preparer, query string) (*Stmt, error) {
	s, err := p.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &Stmt{Stmt: s, Mapper: defaultMapper}, nil
}

// PreparexContext prepares a statement with context and returns an *sqlx.Stmt.
func PreparexContext(ctx context.Context, p PreparerContext, query string) (*Stmt, error) {
	s, err := p.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &Stmt{Stmt: s, Mapper: defaultMapper}, nil
}
