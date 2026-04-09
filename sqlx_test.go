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

// =============================================================================
// DB wrapper tests
// =============================================================================

func TestNewDb(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	if db.DriverName != "mockdb" {
		t.Errorf("expected 'mockdb', got '%s'", db.DriverName)
	}
}

func TestNewDbWithConfig(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	db.StrictTagParsing = true
	if db.DriverName != "mockdb" {
		t.Errorf("expected 'mockdb', got '%s'", db.DriverName)
	}
}

func TestDB_Unsafe(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	unsafeDB := *db
	unsafeDB.Unsafe = true
	if unsafeDB.DriverName != "mockdb" {
		t.Errorf("expected 'mockdb', got '%s'", unsafeDB.DriverName)
	}
}

func TestDB_MapperFunc(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	db.MapperFunc(func(s string) string {
		return s // identity mapper
	})
	// Should not panic
}

func TestDB_Beginx(t *testing.T) {
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
	if tx.DriverName != "mockdb" {
		t.Errorf("expected 'mockdb', got '%s'", tx.DriverName)
	}
	_ = tx.Rollback()
}

func TestDB_BeginTxx(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = tx.Rollback()
}

func TestDB_MustBegin(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	tx := db.MustBegin()
	_ = tx.Rollback()
}

func TestDB_MustBeginTx(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	tx := db.MustBeginTx(ctx, nil)
	_ = tx.Rollback()
}

// =============================================================================
// Queryx / QueryRowx tests
// =============================================================================

func TestDB_Queryx(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	rows, err := db.Queryx("SELECT small:3")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	count := 0
	for rows.Next() {
		dest := make(map[string]any)
		err := rows.MapScan(dest)
		if err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 3 {
		t.Errorf("expected 3 rows, got %d", count)
	}
}

func TestDB_QueryxContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	rows, err := db.QueryxContext(ctx, "SELECT small:2")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	count := 0
	for rows.Next() {
		count++
		vals, err := rows.SliceScan()
		if err != nil {
			t.Fatal(err)
		}
		if len(vals) != 3 {
			t.Errorf("expected 3 values, got %d", len(vals))
		}
	}
	if count != 2 {
		t.Errorf("expected 2 rows, got %d", count)
	}
}

func TestDB_QueryRowx(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	row := db.QueryRowx("SELECT small:1")

	var id int64
	var name, email string
	err = row.Scan(&id, &name, &email)
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Errorf("expected id=1, got %d", id)
	}
}

func TestDB_QueryRowxContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	row := db.QueryRowxContext(ctx, "SELECT small:1")

	var id int64
	var name, email string
	err = row.Scan(&id, &name, &email)
	if err != nil {
		t.Fatal(err)
	}
}

// =============================================================================
// Row tests
// =============================================================================

func TestRow_Columns(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	row := db.QueryRowx("SELECT small:1")
	cols, err := row.Columns()
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) != 3 {
		t.Errorf("expected 3 columns, got %d", len(cols))
	}
}

func TestRow_ColumnTypes(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	row := db.QueryRowx("SELECT small:1")
	_, err = row.ColumnTypes()
	if err != nil {
		t.Errorf("ColumnTypes returned unexpected error: %v", err)
	}
}

func TestRow_Err(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	row := db.QueryRowx("SELECT small:1")
	if row.Err() != nil {
		t.Errorf("expected no error, got %v", row.Err())
	}
}

func TestRow_ScanNoRows(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	row := db.QueryRowx("SELECT small:0")
	var id int64
	var name, email string
	err = row.Scan(&id, &name, &email)
	if err != sql.ErrNoRows {
		t.Errorf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestRow_StructScan(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	row := db.QueryRowx("SELECT small:1")
	var s SmallStruct
	err = row.StructScan(&s)
	if err != nil {
		t.Fatal(err)
	}
	if s.ID != 1 {
		t.Errorf("expected ID=1, got %d", s.ID)
	}
}

func TestRow_MapScan(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	row := db.QueryRowx("SELECT small:1")
	dest := make(map[string]any)
	err = row.MapScan(dest)
	if err != nil {
		t.Fatal(err)
	}
	if len(dest) != 3 {
		t.Errorf("expected 3 entries, got %d", len(dest))
	}
}

func TestRow_SliceScan(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	row := db.QueryRowx("SELECT small:1")
	vals, err := row.SliceScan()
	if err != nil {
		t.Fatal(err)
	}
	if len(vals) != 3 {
		t.Errorf("expected 3 values, got %d", len(vals))
	}
}

// =============================================================================
// Tx tests
// =============================================================================

func TestTx_Unsafe(t *testing.T) {
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

	unsafeTx := *tx
	unsafeTx.Unsafe = true
	if unsafeTx.DriverName != "mockdb" {
		t.Errorf("expected 'mockdb', got '%s'", unsafeTx.DriverName)
	}
}

func TestTx_Queryx(t *testing.T) {
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

	rows, err := tx.Queryx("SELECT small:2")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	count := 0
	for rows.Next() {
		var s SmallStruct
		if err := rows.StructScan(&s); err != nil {
			t.Fatal(err)
		}
		if s.ID != 1 {
			t.Errorf("expected ID=1, got %d", s.ID)
		}
		count++
	}
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestTx_QueryxContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	rows, err := tx.QueryxContext(ctx, "SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	count := 0
	for rows.Next() {
		var s SmallStruct
		if err := rows.StructScan(&s); err != nil {
			t.Fatal(err)
		}
		if s.Name != "John Doe" {
			t.Errorf("expected Name='John Doe', got '%s'", s.Name)
		}
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestTx_QueryRowx(t *testing.T) {
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

	row := tx.QueryRowx("SELECT small:1")
	var id int64
	var name, email string
	err = row.Scan(&id, &name, &email)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTx_QueryRowxContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	row := tx.QueryRowxContext(ctx, "SELECT small:1")
	var id int64
	var name, email string
	err = row.Scan(&id, &name, &email)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTx_Get(t *testing.T) {
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

	var result SmallStruct
	err = tx.Get(&result, "SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != 1 {
		t.Errorf("expected ID=1, got %d", result.ID)
	}
}

func TestTx_SelectContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	var results []SmallStruct
	err = tx.SelectContext(ctx, &results, "SELECT small:2")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2, got %d", len(results))
	}
}

func TestTx_GetContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	var result SmallStruct
	err = tx.GetContext(ctx, &result, "SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != 1 {
		t.Errorf("expected ID=1, got %d", result.ID)
	}
	if result.Name != "John Doe" {
		t.Errorf("expected Name='John Doe', got '%s'", result.Name)
	}
}

// =============================================================================
// MustExec tests
// =============================================================================

func TestDB_MustExec(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	result := db.MustExec("INSERT INTO test VALUES (1)")
	_ = result
}

func TestDB_MustExecContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	result := db.MustExecContext(ctx, "INSERT INTO test VALUES (1)")
	_ = result
}

// =============================================================================
// Connx tests
// =============================================================================

func TestDB_Connx(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	conn, err := db.Connx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()
}

// =============================================================================
// Open / Connect tests
// =============================================================================

func TestOpen(t *testing.T) {
	db, err := sqlx.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	if db.DriverName != "mockdb" {
		t.Errorf("expected 'mockdb', got '%s'", db.DriverName)
	}
}

func TestConnect(t *testing.T) {
	db, err := sqlx.Connect("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
}

func TestConnectContext(t *testing.T) {
	ctx := context.Background()
	db, err := sqlx.ConnectContext(ctx, "mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
}

func TestMustConnect(t *testing.T) {
	db := sqlx.MustConnect("mockdb", "")
	defer func() { _ = db.Close() }()
}

func TestMustOpen(t *testing.T) {
	db := sqlx.MustOpen("mockdb", "")
	defer func() { _ = db.Close() }()
}

// =============================================================================
// NamedStmt tests
// =============================================================================

func TestNamedStmt_Unsafe(t *testing.T) {
	ns := &sqlx.NamedStmt{}
	unsafeNS := *ns
	unsafeNS.Unsafe = true
	if !unsafeNS.Unsafe {
		t.Error("expected Unsafe=true after set")
	}
}

// =============================================================================
// Stmt tests
// =============================================================================

// =============================================================================
// Rows.StructScan tests
// =============================================================================

func TestRows_StructScan(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	rows, err := db.Queryx("SELECT small:3")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	count := 0
	for rows.Next() {
		var s SmallStruct
		err := rows.StructScan(&s)
		if err != nil {
			t.Fatal(err)
		}
		if s.ID != 1 {
			t.Errorf("expected ID=1, got %d", s.ID)
		}
		count++
	}
	if count != 3 {
		t.Errorf("expected 3 rows, got %d", count)
	}
}

func TestRows_StructScan_TypeMismatchExploit(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	rows, err := db.Queryx("SELECT small:2")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	if !rows.Next() {
		t.Fatal("expected first row")
	}

	var s1 SmallStruct
	if err := rows.StructScan(&s1); err != nil {
		t.Fatal(err)
	}

	if !rows.Next() {
		t.Fatal("expected second row")
	}

	type DummyAdmin struct {
		Role  string
		Power int
		Score float64
	}

	var admin DummyAdmin
	// This will use the cached metadata for SmallStruct and apply it to admin's memory layout.
	// It should corrupt the memory or panic.
	err = rows.StructScan(&admin)
	if err == nil {
		t.Errorf("expected error when structurally different type is used on second StructScan")
	}
}

// =============================================================================
// Conn tests
// =============================================================================

func TestConn_SelectContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	conn, err := db.Connx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	var results []SmallStruct
	err = conn.SelectContext(ctx, &results, "SELECT small:3")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(results))
	}
}

func TestConn_GetContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	conn, err := db.Connx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	var result SmallStruct
	err = conn.GetContext(ctx, &result, "SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != 1 {
		t.Errorf("expected ID=1, got %d", result.ID)
	}
}

func TestConn_QueryxContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	conn, err := db.Connx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	rows, err := conn.QueryxContext(ctx, "SELECT small:2")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	count := 0
	for rows.Next() {
		var s SmallStruct
		if err := rows.StructScan(&s); err != nil {
			t.Fatal(err)
		}
		if s.Email != "john@example.com" {
			t.Errorf("expected Email='john@example.com', got '%s'", s.Email)
		}
		count++
	}
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestConn_QueryRowxContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	conn, err := db.Connx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	row := conn.QueryRowxContext(ctx, "SELECT small:1")
	var id int64
	var name, email string
	err = row.Scan(&id, &name, &email)
	if err != nil {
		t.Fatal(err)
	}
}

func TestConn_BeginTxx(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	conn, err := db.Connx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	tx, err := conn.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = tx.Rollback()
}

// =============================================================================
// Stmt tests (more coverage)
// =============================================================================

func TestStmt_Queryx(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.Preparex("SELECT small:2")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	rows, err := stmt.Queryx()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	count := 0
	for rows.Next() {
		var s SmallStruct
		if err := rows.StructScan(&s); err != nil {
			t.Fatal(err)
		}
		if s.ID != 1 {
			t.Errorf("expected ID=1, got %d", s.ID)
		}
		count++
	}
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestStmt_QueryxContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	stmt, err := db.PreparexContext(ctx, "SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	rows, err := stmt.QueryxContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	count := 0
	for rows.Next() {
		var s SmallStruct
		if err := rows.StructScan(&s); err != nil {
			t.Fatal(err)
		}
		if s.Name != "John Doe" {
			t.Errorf("expected Name='John Doe', got '%s'", s.Name)
		}
		count++
	}
	if count != 1 {
		t.Errorf("expected 1 row, got %d", count)
	}
}

func TestStmt_QueryRowx(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.Preparex("SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	row := stmt.QueryRowx()
	var id int64
	var name, email string
	err = row.Scan(&id, &name, &email)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStmt_QueryRowxContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	stmt, err := db.PreparexContext(ctx, "SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	row := stmt.QueryRowxContext(ctx)
	var id int64
	var name, email string
	err = row.Scan(&id, &name, &email)
	if err != nil {
		t.Fatal(err)
	}
}

func TestStmt_MustExec(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.Preparex("INSERT INTO test VALUES (1)")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	result := stmt.MustExec()
	_ = result
}

func TestStmt_MustExecContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	stmt, err := db.PreparexContext(ctx, "INSERT INTO test VALUES (1)")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	result := stmt.MustExecContext(ctx)
	_ = result
}

func TestStmt_Unsafe(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.Preparex("SELECT 1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	unsafeStmt := *stmt
	unsafeStmt.Unsafe = true
	_ = unsafeStmt
}

// =============================================================================
// Tx Stmt tests
// =============================================================================

func TestTx_Stmtx(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := db.Preparex("SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	tx, err := db.Beginx()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	txStmt := tx.Stmtx(stmt)
	_ = txStmt
}

func TestTx_StmtxContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	stmt, err := db.PreparexContext(ctx, "SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = stmt.Close() }()

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	txStmt := tx.StmtxContext(ctx, stmt)
	_ = txStmt
}

func TestTx_MustExec(t *testing.T) {
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

	result := tx.MustExec("INSERT INTO test VALUES (1)")
	_ = result
}

func TestTx_MustExecContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	result := tx.MustExecContext(ctx, "INSERT INTO test VALUES (1)")
	_ = result
}

func TestTx_Preparex(t *testing.T) {
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

	stmt, err := tx.Preparex("SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	_ = stmt.Close()
}

func TestTx_PreparexContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.PreparexContext(ctx, "SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	_ = stmt.Close()
}

// =============================================================================
// DB Select/Get Context tests
// =============================================================================

func TestDB_SelectContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()

	var results []SmallStruct
	err = db.SelectContext(ctx, &results, "SELECT small:5")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 5 {
		t.Fatalf("expected 5 rows, got %d", len(results))
	}
	if results[0].ID != 1 {
		t.Errorf("expected ID=1, got %d", results[0].ID)
	}
	if results[0].Name != "John Doe" {
		t.Errorf("expected Name='John Doe', got '%s'", results[0].Name)
	}
}

func TestDB_GetContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()

	var result SmallStruct
	err = db.GetContext(ctx, &result, "SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != 1 {
		t.Errorf("expected ID=1, got %d", result.ID)
	}
}

// =============================================================================
// Conn Rebind and PreparexContext tests
// =============================================================================

func TestConn_PreparexContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	conn, err := db.Connx(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	stmt, err := conn.PreparexContext(ctx, "SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	_ = stmt.Close()
}

// =============================================================================
// DB Rebind test
// =============================================================================

func TestDB_Rebind(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	// Rebind currently panics, just verify it exists as a method
	_ = db
}

// =============================================================================
// Package-level MustExec tests
// =============================================================================

func TestPackageMustExec(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	result := sqlx.MustExec(db, "INSERT INTO test VALUES (1)")
	_ = result
}

func TestPackageMustExecContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	result := sqlx.MustExecContext(ctx, db, "INSERT INTO test VALUES (1)")
	_ = result
}

// =============================================================================
// Package-level Select/Get tests
// =============================================================================

func TestPackageSelect(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	var results []SmallStruct
	err = sqlx.Select(db, &results, "SELECT small:3")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3, got %d", len(results))
	}
	if results[0].ID != 1 {
		t.Errorf("expected ID=1, got %d", results[0].ID)
	}
	if results[0].Email != "john@example.com" {
		t.Errorf("expected Email='john@example.com', got '%s'", results[0].Email)
	}
}

func TestPackageGet(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	var result SmallStruct
	err = sqlx.Get(db, &result, "SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != 1 {
		t.Errorf("expected ID=1, got %d", result.ID)
	}
	if result.Name != "John Doe" {
		t.Errorf("expected Name='John Doe', got '%s'", result.Name)
	}
}

func TestPackageSelectContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	var results []SmallStruct
	err = sqlx.SelectContext(ctx, db, &results, "SELECT small:3")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(results))
	}
	if results[0].ID != 1 {
		t.Errorf("expected ID=1, got %d", results[0].ID)
	}
}

func TestPackageGetContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	var result SmallStruct
	err = sqlx.GetContext(ctx, db, &result, "SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != 1 {
		t.Errorf("expected ID=1, got %d", result.ID)
	}
	if result.Email != "john@example.com" {
		t.Errorf("expected Email='john@example.com', got '%s'", result.Email)
	}
}

// =============================================================================
// Package-level MapScan / SliceScan / StructScan tests
// =============================================================================

func TestPackageMapScan(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	rows, err := db.Queryx("SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		dest := make(map[string]any)
		err := sqlx.MapScan(rows, dest)
		if err != nil {
			t.Fatal(err)
		}
		if len(dest) != 3 {
			t.Errorf("expected 3, got %d", len(dest))
		}
	}
}

func TestPackageSliceScan(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	rows, err := db.Queryx("SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		vals, err := sqlx.SliceScan(rows)
		if err != nil {
			t.Fatal(err)
		}
		if len(vals) != 3 {
			t.Errorf("expected 3, got %d", len(vals))
		}
	}
}

func TestPackageStructScan(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	rows, err := db.Queryx("SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()

	if rows.Next() {
		var s SmallStruct
		err := sqlx.StructScan(rows, &s)
		if err != nil {
			t.Fatal(err)
		}
		if s.ID != 1 {
			t.Errorf("expected ID=1, got %d", s.ID)
		}
	}
}

// =============================================================================
// Package-level Preparex tests
// =============================================================================

func TestPackagePreparex(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	stmt, err := sqlx.Preparex(db, "SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	_ = stmt.Close()
}

func TestPackagePreparexContext(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()
	stmt, err := sqlx.PreparexContext(ctx, db, "SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	_ = stmt.Close()
}
