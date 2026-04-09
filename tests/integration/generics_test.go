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
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/rpadovani/sqlx-v2/internal/mockdb"
)

// =============================================================================
// SelectG[T] tests
// =============================================================================

func TestSelect_Small(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()
	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()

	results, err := sqlx.SelectG[SmallStruct](ctx, db, "SELECT small:5")
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

func TestSelect_Medium(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()
	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()

	results, err := sqlx.SelectG[MediumStruct](ctx, db, "SELECT medium:10")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 10 {
		t.Fatalf("expected 10 rows, got %d", len(results))
	}
	if results[0].FirstName != "John" {
		t.Errorf("expected FirstName='John', got '%s'", results[0].FirstName)
	}
}

func TestSelect_Large(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()
	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()

	results, err := sqlx.SelectG[LargeStruct](ctx, db, "SELECT large:3")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(results))
	}
	if results[0].Field00 != 0 {
		t.Errorf("expected Field00=0, got %d", results[0].Field00)
	}
	if results[0].Field01 != "value_1" {
		t.Errorf("expected Field01='value_1', got '%s'", results[0].Field01)
	}
	if results[0].Field49 != "value_49" {
		t.Errorf("expected Field49='value_49', got '%s'", results[0].Field49)
	}
	if results[0].Field48 != 48 {
		t.Errorf("expected Field48=48, got %d", results[0].Field48)
	}
}

func TestSelect_Empty(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()
	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()

	results, err := sqlx.SelectG[SmallStruct](ctx, db, "SELECT small:0")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(results))
	}
}

// =============================================================================
// GetG[T] tests
// =============================================================================

func TestGet_Small(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()
	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()

	result, err := sqlx.GetG[SmallStruct](ctx, db, "SELECT small:1")
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

func TestGetG_NoRows(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()
	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()

	_, err = sqlx.GetG[SmallStruct](ctx, db, "SELECT small:0")
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

// =============================================================================
// SelectIter[T] (iter.Seq2) tests
// =============================================================================

func TestSelectIter_Small(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()
	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()

	count := 0
	for row, err := range sqlx.SelectIter[SmallStruct](ctx, db, "SELECT small:5") {
		if err != nil {
			t.Fatal(err)
		}
		count++
		if row.ID != 1 {
			t.Errorf("expected ID=1, got %d", row.ID)
		}
	}
	if count != 5 {
		t.Fatalf("expected 5 iterations, got %d", count)
	}
}

func TestSelectIter_EarlyBreak(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()
	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()

	count := 0
	for _, err := range sqlx.SelectIter[SmallStruct](ctx, db, "SELECT small:100") {
		if err != nil {
			t.Fatal(err)
		}
		count++
		if count >= 3 {
			break // Should cleanly stop iteration
		}
	}
	if count != 3 {
		t.Fatalf("expected 3 iterations before break, got %d", count)
	}
}

func TestSelectIter_Empty(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()
	db := sqlx.NewDb(rawDB, "mockdb")
	ctx := context.Background()

	count := 0
	for _, err := range sqlx.SelectIter[SmallStruct](ctx, db, "SELECT small:0") {
		if err != nil {
			t.Fatal(err)
		}
		count++
	}
	if count != 0 {
		t.Fatalf("expected 0 iterations, got %d", count)
	}
}

// =============================================================================
// Transaction generics tests
// =============================================================================

func TestSelectTx(t *testing.T) {
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

	results, err := sqlx.SelectG[SmallStruct](ctx, tx, "SELECT small:3")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(results))
	}
}

func TestGetTx(t *testing.T) {
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

	result, err := sqlx.GetG[SmallStruct](ctx, tx, "SELECT small:1")
	if err != nil {
		t.Fatal(err)
	}
	if result.ID != 1 {
		t.Errorf("expected ID=1, got %d", result.ID)
	}
}

func TestGetTx_NoRows(t *testing.T) {
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

	_, err = sqlx.GetG[SmallStruct](ctx, tx, "SELECT small:0")
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

// =============================================================================
// Vulnerability Tests
// =============================================================================

type AliasTrap struct {
	ID   int     `db:"id"`
	Name *string `db:"name"`
}

func TestSelectIter_SyncPool_PointerAlias(t *testing.T) {
	rawDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	_, err = rawDB.Exec(`CREATE TABLE alias_trap (id INTEGER, name TEXT);
		INSERT INTO alias_trap (id, name) VALUES (1, 'Alice'), (2, 'Bob'), (3, 'Charlie');`)
	if err != nil {
		t.Fatal(err)
	}

	db := sqlx.NewDb(rawDB, "sqlite3")
	ctx := context.Background()

	var names []*string
	for row, err := range sqlx.SelectIter[AliasTrap](ctx, db, "SELECT id, name FROM alias_trap ORDER BY id") {
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, row.Name)
	}

	if len(names) != 3 {
		t.Fatalf("expected 3 names, got %d", len(names))
	}

	// If pointer aliasing bug exists, all pointers will point to the last row's value ("Charlie")
	if *names[0] != "Alice" || *names[1] != "Bob" || *names[2] != "Charlie" {
		t.Errorf("pointer aliasing detected: expected [Alice, Bob, Charlie], got [%s, %s, %s]",
			*names[0], *names[1], *names[2])
	}
}

