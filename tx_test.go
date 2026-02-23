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

package sqlx_test

import (
	"context"
	"database/sql"
	"testing"

	sqlx "github.com/rpadovani/sqlx-v2"
)

func TestTx_Rebind(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	q := tx.Rebind("SELECT * FROM test WHERE id = ?")
	if q != "SELECT * FROM test WHERE id = ?" {
		t.Errorf("expected '?', got '%s'", q)
	}
}

func TestTx_BindNamed(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	arg := map[string]any{"id": 1}
	q, args, err := tx.BindNamed("SELECT * FROM test WHERE id = :id", arg)
	if err != nil {
		t.Fatal(err)
	}
	if q != "SELECT * FROM test WHERE id = ?" {
		t.Errorf("expected '?', got '%s'", q)
	}
	if len(args) != 1 || args[0] != 1 {
		t.Errorf("expected [1], got %v", args)
	}
}

func TestTx_NamedQuery(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	arg := map[string]any{"id": 1, "1": 1}
	rows, err := tx.NamedQuery("SELECT default WHERE id = :id", arg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
}

func TestTx_NamedQueryContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	arg := map[string]any{"id": 2, "2": 2}
	ctx := context.Background()
	rows, err := tx.NamedQueryContext(ctx, "SELECT default WHERE id = :id", arg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
}

func TestTx_NamedExec(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	arg := map[string]any{"id": 1}
	res, err := tx.NamedExec("INSERT INTO test VALUES (:id)", arg)
	if err != nil {
		t.Fatal(err)
	}
	if res == nil {
		t.Error("expected non-nil result")
	}
}

func TestTx_NamedExecContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	arg := map[string]any{"id": 1}
	ctx := context.Background()
	res, err := tx.NamedExecContext(ctx, "INSERT INTO test VALUES (:id)", arg)
	if err != nil {
		t.Fatal(err)
	}
	if res == nil {
		t.Error("expected non-nil result")
	}
}

func TestTx_Select(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	var results []SmallStruct
	err = tx.Select(&results, "SELECT small:2")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2, got %d", len(results))
	}
}

func TestTx_PrepareNamed(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PrepareNamed("SELECT small:2 WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()
	if stmt.QueryString == "" {
		t.Error("expected non-empty QueryString")
	}
}

func TestTx_PrepareNamedContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	ctx := context.Background()
	stmt, err := tx.PrepareNamedContext(ctx, "SELECT small:2 WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()
	if stmt.QueryString == "" {
		t.Error("expected non-empty QueryString")
	}
}

func TestTx_NamedStmt(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := db.PrepareNamed("SELECT small:2 WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	txStmt := tx.NamedStmt(stmt)
	if txStmt == nil || txStmt.Stmt == nil {
		t.Error("expected non-nil txStmt")
	}
}

func TestTx_NamedStmtContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	ctx := context.Background()
	stmt, err := db.PrepareNamedContext(ctx, "SELECT small:2 WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	txStmt := tx.NamedStmtContext(ctx, stmt)
	if txStmt == nil || txStmt.Stmt == nil {
		t.Error("expected non-nil txStmt")
	}
}
