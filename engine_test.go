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
	"database/sql"
	"testing"

	sqlx "github.com/rpadovani/sqlx-v2"
	_ "github.com/rpadovani/sqlx-v2/internal/mockdb"
)

func TestSelect_SmallStruct(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")

	var results []SmallStruct
	err = db.Select(&results, "SELECT small:5")
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
	if results[0].Email != "john@example.com" {
		t.Errorf("expected Email='john@example.com', got '%s'", results[0].Email)
	}
}

func TestSelect_MediumStruct(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")

	var results []MediumStruct
	err = db.Select(&results, "SELECT medium:10")
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

func TestSelect_LargeStruct(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")

	var results []LargeStruct
	err = db.Select(&results, "SELECT large:5")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 5 {
		t.Fatalf("expected 5 rows, got %d", len(results))
	}
	// Validate first and last field values to ensure the full
	// 50-field struct was mapped correctly end-to-end.
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

func TestGet_SmallStruct(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")

	var result SmallStruct
	err = db.Get(&result, "SELECT small:1")
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

func TestGet_MediumStruct(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")

	var result MediumStruct
	err = db.Get(&result, "SELECT medium:1")
	if err != nil {
		t.Fatal(err)
	}
	if result.FirstName != "John" {
		t.Errorf("expected FirstName='John', got '%s'", result.FirstName)
	}
}

func TestSelect_PointerSlice(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")

	var results []*SmallStruct
	err = db.Select(&results, "SELECT small:3")
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

func TestSelect_EmptyResult(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")

	var results []SmallStruct
	err = db.Select(&results, "SELECT small:0")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 rows, got %d", len(results))
	}
}

func TestGet_NoRows(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")

	var result SmallStruct
	err = db.Get(&result, "SELECT small:0")
	if err != sql.ErrNoRows {
		t.Fatalf("expected sql.ErrNoRows, got %v", err)
	}
}

func TestSelect_WithTransaction(t *testing.T) {
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
	err = tx.Select(&results, "SELECT small:3")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(results))
	}
}

func TestSelect_InvalidDest(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")

	var result SmallStruct
	err = db.Select(&result, "SELECT small:1")
	if err == nil {
		t.Fatal("expected error for non-slice dest")
	}
}

// EmbeddedStruct tests
type Address struct {
	Street string `db:"street"`
	City   string `db:"city"`
}

type Person struct {
	Address
	Name string `db:"name"`
}

func TestSelect_EmbeddedStruct(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")

	var results []Person
	err = db.Select(&results, "SELECT embedded:2")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(results))
	}
	// Verify the flattened embedded struct fields are scanned correctly
	if results[0].Name != "Alice" {
		t.Errorf("expected Name='Alice', got '%s'", results[0].Name)
	}
	if results[0].Street != "123 Main St" {
		t.Errorf("expected Street='123 Main St', got '%s'", results[0].Street)
	}
	if results[0].City != "Springfield" {
		t.Errorf("expected City='Springfield', got '%s'", results[0].City)
	}
}
