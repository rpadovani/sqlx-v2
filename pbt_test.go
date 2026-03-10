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
	"reflect"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	_ "modernc.org/sqlite"

	sqlx "github.com/rpadovani/sqlx-v2"
	"github.com/rpadovani/sqlx-v2/internal/testutil"
)

// PBTStruct represents a generated database row containing various supported scalar types.
type PBTStruct struct {
	ID      int64   `db:"id"`
	Name    string  `db:"name"`
	IsOk    bool    `db:"is_ok"`
	Balance float64 `db:"balance"`
	Age     int     `db:"age"` // Note: SQLite driver might return int64 for ints, so we stick to strictly mapping it.
}

// genPBTStruct generates a random struct.
func genPBTStruct() gopter.Gen {
	return gen.Struct(reflect.TypeFor[PBTStruct](), map[string]gopter.Gen{
		"ID":      gen.Int64(),
		"Name":    gen.AlphaString(),
		"IsOk":    gen.Bool(),
		"Balance": gen.Float64(),
		"Age":     gen.IntRange(0, 120),
	})
}

// TestPropertySymmetry validates that writing a struct to the database
// and reading it back yields structurally identical values.
func TestPropertySymmetry(t *testing.T) {
	drivers := []struct {
		name       string
		getDSN     func(t *testing.T) string
		createStmt string
	}{
		{
			name:   "sqlite",
			getDSN: func(t *testing.T) string { return ":memory:" },
			createStmt: `CREATE TABLE pbt_test (
				id INTEGER PRIMARY KEY,
				name TEXT,
				is_ok BOOLEAN,
				balance REAL,
				age INTEGER
			)`,
		},
		{
			name:   "pgx", // The registered name for github.com/jackc/pgx/v5/stdlib
			getDSN: testutil.GlobalPostgres,
			createStmt: `CREATE TABLE pbt_test (
				id BIGINT PRIMARY KEY,
				name TEXT,
				is_ok BOOLEAN,
				balance DOUBLE PRECISION,
				age INTEGER
			)`,
		},
		{
			name:   "mysql",
			getDSN: testutil.GlobalMySQL,
			createStmt: `CREATE TABLE pbt_test (
				id BIGINT PRIMARY KEY,
				name TEXT,
				is_ok BOOLEAN,
				balance DOUBLE,
				age INTEGER
			)`,
		},
	}

	for _, d := range drivers {
		t.Run(d.name, func(t *testing.T) {
			dsn := d.getDSN(t)
			if dsn == "skip" {
				t.Skip("Skipping test; Docker not available")
			}

			db := sqlx.MustConnect(d.name, dsn)
			defer func() { _ = db.Close() }()

			// Reset table for each run to avoid state leakage
			_, _ = db.Exec("DROP TABLE IF EXISTS pbt_test")
			_, err := db.Exec(d.createStmt)
			if err != nil {
				t.Fatalf("Failed to create table: %v", err)
			}

			params := gopter.DefaultTestParameters()
			params.MinSuccessfulTests = 500
			properties := gopter.NewProperties(params)

			properties.Property("Struct Symmetry", prop.ForAll(
				func(in PBTStruct) bool {
					// Clear table inside the loop to ensure isolation per generated case
					_, err := db.Exec("DELETE FROM pbt_test")
					if err != nil {
						return false
					}

					// Write
					var insertQuery string
					if d.name == "pgx" || d.name == "mysql" {
						insertQuery = db.Rebind(`
							INSERT INTO pbt_test (id, name, is_ok, balance, age) 
							VALUES (?, ?, ?, ?, ?)
						`)
						_, err = db.Exec(insertQuery, in.ID, in.Name, in.IsOk, in.Balance, in.Age)
					} else {
						// SQLite fallback uses native named args
						_, err = db.NamedExec(`
							INSERT INTO pbt_test (id, name, is_ok, balance, age) 
							VALUES (:id, :name, :is_ok, :balance, :age)
						`, in)
					}

					if err != nil {
						return false
					}

					// Read
					var out PBTStruct
					err = db.Get(&out, db.Rebind("SELECT * FROM pbt_test WHERE id = ?"), in.ID)
					if err != nil {
						return false
					}

					// Assert
					return reflect.DeepEqual(in, out)
				},
				genPBTStruct(),
			))

			properties.TestingRun(t)
		})
	}
}
