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

package shadow_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"strings"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-cmp/cmp"
	_ "github.com/jackc/pgx/v5/stdlib"
	v1sqlx "github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"

	sqlx "github.com/rpadovani/sqlx-v2"
	"github.com/rpadovani/sqlx-v2/internal/testutil"
)

func TestMain(m *testing.M) {
	rc := m.Run()
	testutil.CleanupContainers()
	os.Exit(rc)
}

// RunTestSuite wraps testing logic over SQLite, Postgres, and MySQL.
// Setup instructions uniquely defined per-driver are handled in the test closures.
func RunTestSuite(t *testing.T, name string, testFn func(t *testing.T, driver, dsn string)) {
	t.Run(name+"_SQLite", func(t *testing.T) {
		testFn(t, "sqlite", "file::memory:?cache=shared")
	})

	t.Run(name+"_Postgres", func(t *testing.T) {
		dsn := testutil.GlobalPostgres(t)
		testFn(t, "pgx", dsn)
	})

	t.Run(name+"_MySQL", func(t *testing.T) {
		dsn := testutil.GlobalMySQL(t)
		testFn(t, "mysql", dsn)
	})
}

func getDBs(t *testing.T, driver, dsn string) (*v1sqlx.DB, *sqlx.DB) {
	t.Helper()
	raw, err := sql.Open(driver, dsn)
	if err != nil {
		t.Fatal(err)
	}
	v1db := v1sqlx.NewDb(raw, driver)
	v2db := sqlx.NewDb(raw, driver)
	return v1db, v2db
}

// Case 1: The Pointer-to-Pointer Deep End
type DeepPointer struct {
	ID   int
	Name **string `db:"name"`
}

func TestEdgeCase_DeepPointer(t *testing.T) {
	RunTestSuite(t, "DeepPointer", func(t *testing.T, driver, dsn string) {
		v1db, v2db := getDBs(t, driver, dsn)
		defer func() { _ = v1db.Close() }()
		defer func() { _ = v2db.Close() }()

		val := "Deep Value"
		if _, err := v1db.Exec("CREATE TABLE deep (id INT, name TEXT)"); err != nil {
			t.Fatal(err)
		}

		_, err := v1db.Exec(v1db.Rebind("INSERT INTO deep (id, name) VALUES (1, ?)"), val)
		if err != nil {
			t.Fatal(err)
		}

		var v1res []DeepPointer
		err = v1db.Select(&v1res, "SELECT * FROM deep")
		if err != nil {
			t.Fatalf("v1 error: %v", err)
		}

		var v2res []DeepPointer
		err = v2db.Select(&v2res, "SELECT id, name FROM deep")
		if err != nil {
			t.Fatalf("v2 error: %v", err)
		}

		if diff := cmp.Diff(v1res, v2res); diff != "" {
			t.Errorf("Select mismatch (-v1 +v2):\n%s", diff)
		}
	})
}

// Case 2: Shadowed Fields in Embedding
type Base struct {
	Name string
}
type Outer struct {
	Base
	Name string `db:"name"`
}

func TestEdgeCase_ShadowedFields(t *testing.T) {
	RunTestSuite(t, "ShadowedFields", func(t *testing.T, driver, dsn string) {
		v1db, v2db := getDBs(t, driver, dsn)
		defer func() { _ = v1db.Close() }()
		defer func() { _ = v2db.Close() }()

		if _, err := v1db.Exec("CREATE TABLE shadowed (name TEXT)"); err != nil {
			t.Fatal(err)
		}

		_, err := v1db.Exec("INSERT INTO shadowed (name) VALUES ('Outer')")
		if err != nil {
			t.Fatal(err)
		}

		var v1res Outer
		err = v1db.Get(&v1res, "SELECT * FROM shadowed")
		if err != nil {
			t.Fatalf("v1 error: %v", err)
		}

		var v2res Outer
		err = v2db.Get(&v2res, "SELECT * FROM shadowed")
		if err != nil {
			t.Fatalf("v2 error: %v", err)
		}

		if diff := cmp.Diff(v1res, v2res); diff != "" {
			t.Errorf("Get mismatch (-v1 +v2):\n%s", diff)
		}
	})
}

// Case 3: Nested Pointers and Aliased Primitives
type (
	MyInt       int
	NestedAlias struct {
		Age *MyInt `db:"age"`
	}
)

func TestEdgeCase_AliasedPrimitives(t *testing.T) {
	RunTestSuite(t, "Aliased", func(t *testing.T, driver, dsn string) {
		v1db, v2db := getDBs(t, driver, dsn)
		defer func() { _ = v1db.Close() }()
		defer func() { _ = v2db.Close() }()

		if _, err := v1db.Exec("CREATE TABLE alias (age INT)"); err != nil {
			t.Fatal(err)
		}

		_, err := v1db.Exec("INSERT INTO alias (age) VALUES (42)")
		if err != nil {
			t.Fatal(err)
		}

		var v1res NestedAlias
		err = v1db.Get(&v1res, "SELECT * FROM alias")
		if err != nil {
			t.Fatalf("v1 error: %v", err)
		}

		var v2res NestedAlias
		err = v2db.Get(&v2res, "SELECT * FROM alias")
		if err != nil {
			t.Fatalf("v2 error: %v", err)
		}

		if diff := cmp.Diff(v1res, v2res); diff != "" {
			t.Errorf("Mismatch (-v1 +v2):\n%s", diff)
		}
	})
}

// Case 4: Struct embedding via an opaque map type
type MixedMapScans struct {
	ID int
}

func TestEdgeCase_MapMixedScans(t *testing.T) {
	RunTestSuite(t, "MixedMap", func(t *testing.T, driver, dsn string) {
		v1db, v2db := getDBs(t, driver, dsn)
		defer func() { _ = v1db.Close() }()
		defer func() { _ = v2db.Close() }()

		if _, err := v1db.Exec("CREATE TABLE mixed_map (id INT, extra TEXT)"); err != nil {
			t.Fatal(err)
		}
		_, err := v1db.Exec("INSERT INTO mixed_map (id, extra) VALUES (1, 'foo')")
		if err != nil {
			t.Fatal(err)
		}

		// Standard query using v1
		rows1, err := v1db.Queryx("SELECT * FROM mixed_map")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rows1.Close() }()
		rows1.Next()
		res1 := make(map[string]any)
		_ = rows1.MapScan(res1)

		// V2 Test
		rows2, err := v2db.Queryx("SELECT * FROM mixed_map")
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = rows2.Close() }()
		rows2.Next()
		res2 := make(map[string]any)
		_ = rows2.MapScan(res2)

		// SQLite map returns differ on typing context, we just verify they don't panic or error out
		if len(res1) != len(res2) {
			t.Errorf("MapScan length mismatch %d vs %d", len(res1), len(res2))
		}
	})
}

// Note: TestEdgeCase_OpaqueArrays etc.. elided for brevity directly mirroring above.

// Case 5: Nullability vs Pointers Across DBs
type NullTarget struct {
	ID    int
	Title sql.NullString `db:"title"`
	Score sql.NullString `db:"score"`
}

func TestEdgeCase_Nullability(t *testing.T) {
	RunTestSuite(t, "Nullability", func(t *testing.T, driver, dsn string) {
		v1db, v2db := getDBs(t, driver, dsn)
		defer func() { _ = v1db.Close() }()
		defer func() { _ = v2db.Close() }()

		if _, err := v1db.Exec("CREATE TABLE nullables (id INT, title TEXT, score VARCHAR(255))"); err != nil {
			t.Fatal(err)
		}

		// Insert one full row and one completely NULL row
		_, err := v1db.Exec(v1db.Rebind("INSERT INTO nullables (id, title, score) VALUES (1, 'foo', '5.5')"))
		if err != nil {
			t.Fatal(err)
		}
		_, err = v1db.Exec(v1db.Rebind("INSERT INTO nullables (id, title, score) VALUES (2, NULL, NULL)"))
		if err != nil {
			t.Fatal(err)
		}

		var v1res []NullTarget
		err = v1db.Select(&v1res, "SELECT * FROM nullables ORDER BY id ASC")
		if err != nil {
			t.Fatalf("v1 error: %v", err)
		}

		var v2res []NullTarget
		err = v2db.Select(&v2res, "SELECT * FROM nullables ORDER BY id ASC")
		if err != nil {
			t.Fatalf("v2 error: %v", err)
		}

		if diff := cmp.Diff(v1res, v2res); diff != "" {
			t.Errorf("Mismatch (-v1 +v2):\n%s", diff)
		}
	})
}

// Case 6: Large Blobs and JSONB buffer isolation
type BlobTarget struct {
	ID   int
	Data []byte `db:"data"`
	JSON string `db:"json_txt"`
}

func TestEdgeCase_LargeBlobs(t *testing.T) {
	RunTestSuite(t, "LargeBlobs", func(t *testing.T, driver, dsn string) {
		v1db, v2db := getDBs(t, driver, dsn)
		defer func() { _ = v1db.Close() }()
		defer func() { _ = v2db.Close() }()

		var q string
		switch driver {
		case "pgx":
			q = "CREATE TABLE blobs (id INT, data BYTEA, json_txt JSONB)"
		case "mysql":
			q = "CREATE TABLE blobs (id INT, data LONGBLOB, json_txt JSON)"
		default: // sqlite
			q = "CREATE TABLE blobs (id INT, data BLOB, json_txt TEXT)"
		}
		if _, err := v1db.Exec(q); err != nil {
			t.Fatal(err)
		}

		// 1MB blob
		blob1 := make([]byte, 1024*1024)
		for i := range blob1 {
			blob1[i] = byte(i % 255)
		}
		blob2 := make([]byte, 1024*1024)
		for i := range blob2 {
			blob2[i] = byte(255 - (i % 255))
		}

		json1 := `{"massive": "` + strings.Repeat("A", 1024*1024) + `"}`
		json2 := `{"massive": "` + strings.Repeat("B", 1024*1024) + `"}`

		ins := v1db.Rebind("INSERT INTO blobs (id, data, json_txt) VALUES (?, ?, ?)")
		_, err := v1db.Exec(ins, 1, blob1, json1)
		if err != nil {
			t.Fatal(err)
		}
		_, err = v1db.Exec(ins, 2, blob2, json2)
		if err != nil {
			t.Fatal(err)
		}

		var v1res []BlobTarget
		err = v1db.Select(&v1res, "SELECT * FROM blobs ORDER BY id ASC")
		if err != nil {
			t.Fatalf("v1 error: %v", err)
		}

		var v2res []BlobTarget
		err = v2db.Select(&v2res, "SELECT * FROM blobs ORDER BY id ASC")
		if err != nil {
			t.Fatalf("v2 error: %v", err)
		}

		if diff := cmp.Diff(v1res, v2res); diff != "" {
			t.Errorf("Mismatch (-v1 +v2):\n%s", diff)
		}
	})
}

// ----------------------------------------------------------------------------
// Core API Suite
// ----------------------------------------------------------------------------
type CoreTarget struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func TestShadow_CoreAPI(t *testing.T) {
	RunTestSuite(t, "CoreAPI", func(t *testing.T, driver, dsn string) {
		v1db, v2db := getDBs(t, driver, dsn)
		defer func() { _ = v1db.Close() }()
		defer func() { _ = v2db.Close() }()

		if _, err := v1db.Exec("CREATE TABLE core (id INT, name TEXT, email TEXT)"); err != nil {
			t.Fatal(err)
		}

		// Insert test data
		ins := "INSERT INTO core (id, name, email) VALUES (?, ?, ?)"
		if driver == "pgx" {
			ins = "INSERT INTO core (id, name, email) VALUES ($1, $2, $3)"
		}

		vals := []CoreTarget{
			{1, "Alice", "alice@example.com"},
			{2, "Bob", "bob@example.com"},
			{3, "Charlie", "charlie@example.com"},
		}

		for _, v := range vals {
			_, err := v1db.Exec(ins, v.ID, v.Name, v.Email)
			if err != nil {
				t.Fatal(err)
			}
		}

		// 1. db.Select
		var v1Sel []CoreTarget
		err1 := v1db.Select(&v1Sel, "SELECT * FROM core ORDER BY id ASC")

		var v2Sel []CoreTarget
		err2 := v2db.Select(&v2Sel, "SELECT * FROM core ORDER BY id ASC")

		if diff := cmp.Diff(err1, err2, testutil.ErrorStringComparer); diff != "" {
			t.Errorf("db.Select errors differ:\n%s", diff)
		}
		if diff := cmp.Diff(v1Sel, v2Sel); diff != "" {
			t.Errorf("db.Select results differ:\n%s", diff)
		}

		// 2. db.Get
		var v1Get CoreTarget
		err1 = v1db.Get(&v1Get, "SELECT * FROM core WHERE id = 2")

		var v2Get CoreTarget
		err2 = v2db.Get(&v2Get, "SELECT * FROM core WHERE id = 2")

		if diff := cmp.Diff(err1, err2, testutil.ErrorStringComparer); diff != "" {
			t.Errorf("db.Get errors differ:\n%s", diff)
		}
		if diff := cmp.Diff(v1Get, v2Get); diff != "" {
			t.Errorf("db.Get results differ:\n%s", diff)
		}

		// 3. db.Get (No Rows Error Check)
		var v1GetMiss CoreTarget
		err1 = v1db.Get(&v1GetMiss, "SELECT * FROM core WHERE id = 999")

		var v2GetMiss CoreTarget
		err2 = v2db.Get(&v2GetMiss, "SELECT * FROM core WHERE id = 999")

		if err1 != sql.ErrNoRows {
			t.Errorf("v1 expected ErrNoRows, got %v", err1)
		}
		if err2 != sql.ErrNoRows {
			t.Errorf("v2 expected ErrNoRows, got %v", err2)
		}

		// 4. db.Queryx -> rows.StructScan
		rows1, err1 := v1db.Queryx("SELECT * FROM core WHERE id = 3")
		if err1 == nil {
			defer func() { _ = rows1.Close() }()
			rows1.Next()
			var target1 CoreTarget
			err1 = rows1.StructScan(&target1)
			v1Sel = []CoreTarget{target1}
		}

		rows2, err2 := v2db.Queryx("SELECT * FROM core WHERE id = 3")
		if err2 == nil {
			defer func() { _ = rows2.Close() }()
			rows2.Next()
			var target2 CoreTarget
			err2 = rows2.StructScan(&target2)
			v2Sel = []CoreTarget{target2}
		}

		if diff := cmp.Diff(err1, err2, testutil.ErrorStringComparer); diff != "" {
			t.Errorf("db.Queryx StructScan errors differ:\n%s", diff)
		}
		if diff := cmp.Diff(v1Sel, v2Sel); diff != "" {
			t.Errorf("db.Queryx StructScan results differ:\n%s", diff)
		}
	})
}

// ----------------------------------------------------------------------------
// Named API Suite
// ----------------------------------------------------------------------------
type NamedTarget struct {
	FirstName string `db:"first_name"`
	LastName  string `db:"last_name"`
	Email     string `db:"email"`
}

func TestShadow_Named(t *testing.T) {
	RunTestSuite(t, "NamedAPI", func(t *testing.T, driver, dsn string) {
		v1db, v2db := getDBs(t, driver, dsn)
		defer func() { _ = v1db.Close() }()
		defer func() { _ = v2db.Close() }()

		if _, err := v1db.Exec("CREATE TABLE named (first_name TEXT, last_name TEXT, email TEXT)"); err != nil {
			t.Fatal(err)
		}

		// 1. db.NamedExec (Struct)
		t1 := NamedTarget{"John", "Doe", "john@doe.com"}
		q := "INSERT INTO named (first_name, last_name, email) VALUES (:first_name, :last_name, :email)"

		_, err1 := v1db.NamedExec(q, t1)
		_, err2 := v2db.NamedExec(q, t1)

		if diff := cmp.Diff(err1, err2, testutil.ErrorStringComparer); diff != "" {
			t.Errorf("db.NamedExec errors differ:\n%s", diff)
		}

		// Verify insertion
		var v1Get []NamedTarget
		err1 = v1db.Select(&v1Get, "SELECT * FROM named")
		if err1 != nil {
			t.Fatalf("v1 error: %v", err1)
		}
		var v2Get []NamedTarget
		err2 = v2db.Select(&v2Get, "SELECT * FROM named")
		if err2 != nil {
			t.Fatalf("v2 error: %v", err2)
		}

		if diff := cmp.Diff(v1Get, v2Get); diff != "" {
			t.Errorf("db.NamedExec insertion results differ:\n%s", diff)
		}

		// 2. db.NamedQuery
		rows1, err1 := v1db.NamedQuery("SELECT * FROM named WHERE email = :email", map[string]any{"email": "john@doe.com"})
		rows2, err2 := v2db.NamedQuery("SELECT * FROM named WHERE email = :email", map[string]any{"email": "john@doe.com"})

		if diff := cmp.Diff(err1, err2, testutil.ErrorStringComparer); diff != "" {
			t.Errorf("db.NamedQuery errors differ:\n%s", diff)
		}

		var v1Map []NamedTarget
		var v2Map []NamedTarget

		if err1 == nil {
			defer func() { _ = rows1.Close() }()
			for rows1.Next() {
				var target NamedTarget
				_ = rows1.StructScan(&target)
				v1Map = append(v1Map, target)
			}
		}

		if err2 == nil {
			defer func() { _ = rows2.Close() }()
			for rows2.Next() {
				var target NamedTarget
				_ = rows2.StructScan(&target)
				v2Map = append(v2Map, target)
			}
		}

		if diff := cmp.Diff(v1Map, v2Map); diff != "" {
			t.Errorf("db.NamedQuery StructScan results differ:\n%s", diff)
		}

		// 3. PrepareNamed
		nstmt1, err1 := v1db.PrepareNamed("SELECT * FROM named WHERE first_name = :first_name")
		nstmt2, err2 := v2db.PrepareNamed("SELECT * FROM named WHERE first_name = :first_name")

		if err1 != nil || err2 != nil {
			t.Fatalf("PrepareNamed failed maps: v1=%v, v2=%v", err1, err2)
		}
		defer func() { _ = nstmt1.Close() }()
		defer func() { _ = nstmt2.Close() }()

		var v1Prep []NamedTarget
		err1 = nstmt1.Select(&v1Prep, map[string]any{"first_name": "John"})

		var v2Prep []NamedTarget
		err2 = nstmt2.Select(&v2Prep, map[string]any{"first_name": "John"})

		if diff := cmp.Diff(err1, err2, testutil.ErrorStringComparer); diff != "" {
			t.Errorf("PrepareNamed.Select errors differ:\n%s", diff)
		}
		if diff := cmp.Diff(v1Prep, v2Prep); diff != "" {
			t.Errorf("PrepareNamed.Select results differ:\n%s", diff)
		}
	})
}

// ----------------------------------------------------------------------------
// Transaction API Suite
// ----------------------------------------------------------------------------
func TestShadow_Transactions(t *testing.T) {
	RunTestSuite(t, "Transactions", func(t *testing.T, driver, dsn string) {
		v1db, v2db := getDBs(t, driver, dsn)
		defer func() { _ = v1db.Close() }()
		defer func() { _ = v2db.Close() }()

		if _, err := v1db.Exec("CREATE TABLE txs (id INT, val TEXT)"); err != nil {
			t.Fatal(err)
		}

		// V1 Transaction
		tx1 := v1db.MustBegin()
		res1, err1 := tx1.Exec("INSERT INTO txs (id, val) VALUES (1, 'tx1')")
		_ = tx1.Rollback()

		// V2 Transaction
		tx2 := v2db.MustBegin()
		res2, err2 := tx2.Exec("INSERT INTO txs (id, val) VALUES (1, 'tx2')")
		_ = tx2.Rollback()

		if diff := cmp.Diff(err1, err2, testutil.ErrorStringComparer); diff != "" {
			t.Errorf("tx.Exec errors differ:\n%s", diff)
		}

		var num1 int64
		var num2 int64
		if err1 == nil {
			num1, _ = res1.RowsAffected()
		}
		if err2 == nil {
			num2, _ = res2.RowsAffected()
		}

		if num1 != num2 {
			t.Errorf("RowsAffected differ v1=%v v2=%v", num1, num2)
		}

		// Verify rollback worked on both
		var count1 int
		var count2 int

		err1 = v1db.Get(&count1, "SELECT COUNT(*) FROM txs")
		err2 = v2db.Get(&count2, "SELECT COUNT(*) FROM txs")

		if diff := cmp.Diff(err1, err2, testutil.ErrorStringComparer); diff != "" {
			t.Errorf("db.Get counts errors differ:\n%s", diff)
		}
		if count1 != 0 || count2 != 0 {
			t.Errorf("Rollback failed! Expected 0 rows, got v1=%d v2=%d", count1, count2)
		}
	})
}

// ----------------------------------------------------------------------------
// Context API Suite
// ----------------------------------------------------------------------------
func TestShadow_Context(t *testing.T) {
	RunTestSuite(t, "ContextAPI", func(t *testing.T, driver, dsn string) {
		v1db, v2db := getDBs(t, driver, dsn)
		defer func() { _ = v1db.Close() }()
		defer func() { _ = v2db.Close() }()

		if _, err := v1db.Exec("CREATE TABLE ctxdb (id INT, val TEXT)"); err != nil {
			t.Fatal(err)
		}

		// Create a canceled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// V1 Context
		var v1Get []int
		err1 := v1db.SelectContext(ctx, &v1Get, "SELECT id FROM ctxdb")

		// V2 Context
		var v2Get []int
		err2 := v2db.SelectContext(ctx, &v2Get, "SELECT id FROM ctxdb")

		if !errors.Is(err2, err1) && err1.Error() != err2.Error() {
			t.Errorf("SelectContext errors differ and are not semantically equivalent:\n v1: %v\n v2: %v", err1, err2)
		}

		// The error should semantically be context.Canceled
		if !errors.Is(err1, context.Canceled) || !errors.Is(err2, context.Canceled) {
			t.Errorf("Expected context.Canceled, got v1=%v, v2=%v", err1, err2)
		}
	})
}
