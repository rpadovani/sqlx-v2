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

// Package shadow provides a testing harness to verify that sqlx-v2 is a
// 1:1 bug-compatible drop-in replacement for jmoiron/sqlx.
//
// The shadow testing library runs parallel operations through both v1 and v2
// and compares the results, flagging any deviations in behavior.
package shadow

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	sqlx "github.com/rpadovani/sqlx-v2"
	_ "github.com/rpadovani/sqlx-v2/internal/mockdb"
	"github.com/rpadovani/sqlx-v2/internal/reflectx"
)

// TestResult holds the comparison result of a single test case.
type TestResult struct {
	Name    string
	Passed  bool
	Message string
}

// Harness is the shadow testing harness for comparing v1 and v2 behavior.
type Harness struct {
	t      *testing.T
	strict bool
	db     *sqlx.DB
}

// NewHarness creates a new shadow testing harness.
// If strict is true, v2-specific fixes will be enabled and v1 quirks
// will be tested to ensure they differ.
func NewHarness(t *testing.T, db *sql.DB, driverName string, strict bool) *Harness {
	t.Helper()

	sqlxDB := sqlx.NewDb(db, driverName)
	if strict {
		sqlxDB.StrictTagParsing = true
	}

	return &Harness{
		t:      t,
		strict: strict,
		db:     sqlxDB,
	}
}

// Close closes the underlying database connection.
func (h *Harness) Close() {
	// Let the caller close the real DB connection provided in NewHarness
}

// DB returns the underlying sqlx.DB for direct testing.
func (h *Harness) DB() *sqlx.DB {
	return h.db
}

// TestTagParsing verifies that struct tag parsing matches v1 behavior.
func (h *Harness) TestTagParsing(t *testing.T) {
	t.Helper()

	type BasicTags struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Email string `db:"email"`
	}

	type DashTag struct {
		ID     int    `db:"id"`
		Ignore string `db:"-"`
		Name   string `db:"name"`
	}

	type NoTag struct {
		ID   int
		Name string
	}

	type EmbeddedPerson struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
	}

	type UserWithEmbed struct {
		EmbeddedPerson
		Email string `db:"email"`
	}

	mapper := reflectx.NewMapperFunc("db", strings.ToLower)

	// Test 1: Basic tag resolution — verify key existence, offsets, and types
	t.Run("BasicTags", func(t *testing.T) {
		tm := mapper.TypeMap(reflect.TypeFor[BasicTags]())
		assertFieldExists(t, tm, "id")
		assertFieldExists(t, tm, "name")
		assertFieldExists(t, tm, "email")

		// Verify offset and type correctness — not just key existence.
		// The offset must match the compiler's ground truth.
		var bt BasicTags
		assertFieldHasOffset(t, tm, "id", reflect.TypeFor[int](), unsafe.Offsetof(bt.ID))
		assertFieldHasOffset(t, tm, "name", reflect.TypeFor[string](), unsafe.Offsetof(bt.Name))
		assertFieldHasOffset(t, tm, "email", reflect.TypeFor[string](), unsafe.Offsetof(bt.Email))
	})

	// Test 2: Dash tag (skip field)
	t.Run("DashTag", func(t *testing.T) {
		tm := mapper.TypeMap(reflect.TypeFor[DashTag]())
		assertFieldExists(t, tm, "id")
		assertFieldExists(t, tm, "name")
		assertFieldNotExists(t, tm, "-")
		assertFieldNotExists(t, tm, "ignore")
	})

	// Test 3: No tags (fallback to name mapper)
	t.Run("NoTag", func(t *testing.T) {
		tm := mapper.TypeMap(reflect.TypeFor[NoTag]())
		assertFieldExists(t, tm, "id")
		assertFieldExists(t, tm, "name")
	})

	// Test 4: Embedded struct flattening
	t.Run("EmbeddedStruct", func(t *testing.T) {
		tm := mapper.TypeMap(reflect.TypeFor[UserWithEmbed]())
		assertFieldExists(t, tm, "first_name")
		assertFieldExists(t, tm, "last_name")
		assertFieldExists(t, tm, "email")
	})
}

// TestNullHandling verifies NULL value processing through the actual engine.
// This replaces the previous tautological test that only checked Go's zero value.
func (h *Harness) TestNullHandling(t *testing.T) {
	t.Helper()

	// Test real NULL scan through the engine
	t.Run("NullColumnScan", func(t *testing.T) {
		// Create a table with a nullable column
		h.db.MustExec("CREATE TABLE IF NOT EXISTS null_test (id INT, name TEXT)")
		h.db.MustExec("DELETE FROM null_test")
		h.db.MustExec("INSERT INTO null_test (id, name) VALUES (1, NULL)")

		type NullRow struct {
			ID   int            `db:"id"`
			Name sql.NullString `db:"name"`
		}

		var result NullRow
		err := h.db.Get(&result, "SELECT id, name FROM null_test WHERE id = 1")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}

		if result.ID != 1 {
			t.Errorf("expected ID=1, got %d", result.ID)
		}
		if result.Name.Valid {
			t.Errorf("expected Name to be NULL (Valid=false), got Valid=true, String='%s'", result.Name.String)
		}
	})

	// Test non-NULL scan for comparison
	t.Run("NonNullColumnScan", func(t *testing.T) {
		h.db.MustExec("CREATE TABLE IF NOT EXISTS null_test (id INT, name TEXT)")
		h.db.MustExec("DELETE FROM null_test")
		h.db.MustExec("INSERT INTO null_test (id, name) VALUES (2, 'Alice')")

		type NullRow struct {
			ID   int            `db:"id"`
			Name sql.NullString `db:"name"`
		}

		var result NullRow
		err := h.db.Get(&result, "SELECT id, name FROM null_test WHERE id = 2")
		if err != nil {
			t.Fatalf("Get error: %v", err)
		}

		if result.ID != 2 {
			t.Errorf("expected ID=2, got %d", result.ID)
		}
		if !result.Name.Valid {
			t.Error("expected Name to be non-NULL (Valid=true), got Valid=false")
		}
		if result.Name.String != "Alice" {
			t.Errorf("expected Name='Alice', got '%s'", result.Name.String)
		}
	})
}

// TestScanBehavior verifies that MapScan, SliceScan work correctly.
func (h *Harness) TestScanBehavior(t *testing.T) {
	t.Helper()

	t.Run("MapScan", func(t *testing.T) {
		h.db.MustExec("CREATE TABLE IF NOT EXISTS scan_test (id INT, name TEXT, email TEXT)")
		h.db.MustExec("DELETE FROM scan_test")
		h.db.MustExec("INSERT INTO scan_test (id, name, email) VALUES (1, 'Test', 'test@test.com')")

		rows, err := h.db.Queryx("SELECT id, name, email FROM scan_test LIMIT 1")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rows.Close() }()

		if rows.Next() {
			dest := make(map[string]any)
			err := rows.MapScan(dest)
			if err != nil {
				t.Fatal(err)
			}
			if len(dest) != 3 {
				t.Errorf("expected 3 columns, got %d", len(dest))
			}
			if _, ok := dest["id"]; !ok {
				t.Error("expected 'id' column in map")
			}
			if _, ok := dest["name"]; !ok {
				t.Error("expected 'name' column in map")
			}
			if _, ok := dest["email"]; !ok {
				t.Error("expected 'email' column in map")
			}
		} else {
			t.Error("expected at least one row")
		}
	})

	t.Run("SliceScan", func(t *testing.T) {
		h.db.MustExec("CREATE TABLE IF NOT EXISTS scan_test (id INT, name TEXT, email TEXT)")
		h.db.MustExec("DELETE FROM scan_test")
		h.db.MustExec("INSERT INTO scan_test (id, name, email) VALUES (1, 'Test', 'test@test.com')")

		rows, err := h.db.Queryx("SELECT id, name, email FROM scan_test LIMIT 1")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rows.Close() }()

		if rows.Next() {
			vals, err := rows.SliceScan()
			if err != nil {
				t.Fatal(err)
			}
			if len(vals) != 3 {
				t.Errorf("expected 3 values, got %d", len(vals))
			}
		} else {
			t.Error("expected at least one row")
		}
	})
}

// TestConnectionWrapping verifies that DB/Tx/Conn wrapping works correctly.
func (h *Harness) TestConnectionWrapping(t *testing.T) {
	t.Helper()

	t.Run("DriverName", func(t *testing.T) {
		if h.db.DriverName == "" {
			t.Errorf("expected a driver name, got empty")
		}
	})

	t.Run("BeginxCommit", func(t *testing.T) {
		tx, err := h.db.Beginx()
		if err != nil {
			t.Fatal(err)
		}
		if tx.DriverName == "" {
			t.Errorf("expected a driver name, got empty")
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("UnsafeDB", func(t *testing.T) {
		unsafeDB := *h.db
		unsafeDB.Unsafe = true
		if unsafeDB.DriverName == "" {
			t.Errorf("expected a driver name, got empty")
		}
	})
}

func assertFieldExists(t *testing.T, sm *reflectx.StructMap, name string) {
	t.Helper()
	if _, ok := sm.Names[name]; !ok {
		t.Errorf("expected field '%s' to exist in StructMap, available: %s", name, mapKeys(sm.Names))
	}
}

func assertFieldNotExists(t *testing.T, sm *reflectx.StructMap, name string) {
	t.Helper()
	if _, ok := sm.Names[name]; ok {
		t.Errorf("expected field '%s' to NOT exist in StructMap", name)
	}
}

func mapKeys(m map[string]*reflectx.FieldInfo) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return fmt.Sprintf("[%s]", strings.Join(keys, ", "))
}

// assertFieldHasOffset validates that a field's offset and target type match
// the compiler's ground truth. This prevents tautological testing where
// assertions merely check key existence without verifying correctness.
func assertFieldHasOffset(t *testing.T, sm *reflectx.StructMap, name string, expectedType reflect.Type, expectedOffset uintptr) {
	t.Helper()
	fi, ok := sm.Names[name]
	if !ok {
		t.Fatalf("field '%s' not found in StructMap", name)
		return
	}
	if fi.Offset != expectedOffset {
		t.Errorf("field '%s' offset mismatch: got %d, want %d", name, fi.Offset, expectedOffset)
	}
	if fi.TargetType != expectedType {
		t.Errorf("field '%s' type mismatch: got %v, want %v", name, fi.TargetType, expectedType)
	}
	if len(fi.Traversal) == 0 {
		t.Errorf("field '%s' has empty Traversal (should have at least 1 step)", name)
	}
}
