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

func TestNamedStmt_Close(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("SELECT default WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	err = stmt.Close()
	if err != nil {
		t.Fatal(err)
	}
}

func TestNamedStmt_Exec(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("INSERT INTO test VALUES (:id)")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1}
	res, err := stmt.Exec(arg)
	if err != nil {
		t.Fatal(err)
	}
	if res == nil {
		t.Error("expected non-nil result")
	}
}

func TestNamedStmt_ExecContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("INSERT INTO test VALUES (:id)")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1}
	ctx := context.Background()
	res, err := stmt.ExecContext(ctx, arg)
	if err != nil {
		t.Fatal(err)
	}
	if res == nil {
		t.Error("expected non-nil result")
	}
}

func TestNamedStmt_Query(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("SELECT default WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1, "1": 1}
	rows, err := stmt.Query(arg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
}

func TestNamedStmt_QueryContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("SELECT default WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1, "1": 1}
	ctx := context.Background()
	rows, err := stmt.QueryContext(ctx, arg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
}

func TestNamedStmt_QueryRow(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("SELECT default WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1, "1": 1}
	row := stmt.QueryRow(arg)
	if row == nil {
		t.Error("expected non-nil row")
	}
}

func TestNamedStmt_QueryRowContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("SELECT default WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1, "1": 1}
	ctx := context.Background()
	row := stmt.QueryRowContext(ctx, arg)
	if row == nil {
		t.Error("expected non-nil row")
	}
}

func TestNamedStmt_QueryRowx(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("SELECT default WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1, "1": 1}
	row := stmt.QueryRowx(arg)
	if row == nil {
		t.Error("expected non-nil row")
	}
}

func TestNamedStmt_QueryRowxContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("SELECT default WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1, "1": 1}
	ctx := context.Background()
	row := stmt.QueryRowxContext(ctx, arg)
	if row == nil {
		t.Error("expected non-nil row")
	}
}

func TestNamedStmt_Queryx(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("SELECT default WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1, "1": 1}
	rows, err := stmt.Queryx(arg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
}

func TestNamedStmt_QueryxContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("SELECT default WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1, "1": 1}
	ctx := context.Background()
	rows, err := stmt.QueryxContext(ctx, arg)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
}

func TestNamedStmt_MustExec(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("INSERT INTO test VALUES (:id)")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1}
	res := stmt.MustExec(arg)
	if res == nil {
		t.Error("expected non-nil result")
	}
}

func TestNamedStmt_MustExecContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("INSERT INTO test VALUES (:id)")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1}
	ctx := context.Background()
	res := stmt.MustExecContext(ctx, arg)
	if res == nil {
		t.Error("expected non-nil result")
	}
}

func TestNamedStmt_Select(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("SELECT default WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1, "1": 1}
	var results []SmallStruct
	err = stmt.Select(&results, arg)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNamedStmt_SelectContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("SELECT default WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1, "1": 1}
	ctx := context.Background()
	var results []SmallStruct
	err = stmt.SelectContext(ctx, &results, arg)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNamedStmt_Get(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("SELECT small:1 WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1, "1": 1}
	var res SmallStruct
	err = stmt.Get(&res, arg)
	if err != nil {
		t.Fatal(err)
	}
	if res.ID != 1 {
		t.Errorf("expected id 1, got %d", res.ID)
	}
}

func TestNamedStmt_GetContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.PrepareNamed("SELECT small:1 WHERE id = :id")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	arg := map[string]any{"id": 1, "1": 1}
	ctx := context.Background()
	var res SmallStruct
	err = stmt.GetContext(ctx, &res, arg)
	if err != nil {
		t.Fatal(err)
	}
	if res.ID != 1 {
		t.Errorf("expected id 1, got %d", res.ID)
	}
}
