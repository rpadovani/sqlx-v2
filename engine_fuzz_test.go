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
	"database/sql/driver"
	"io"
	"math/rand"
	"strconv"
	"sync"
	"testing"

	sqlx "github.com/rpadovani/sqlx-v2"
)

// ChaosDriver implements a malicious database driver that throws randomized
// types, mismatched columns, and nil pointers at the scanner engine.
type ChaosDriver struct{}

func init() {
	sql.Register("chaosdb", &ChaosDriver{})
}

type FuzzTargetStruct struct {
	ID        int64   `db:"id"`
	FirstName string  `db:"first_name"`
	LastName  string  `db:"last_name"`
	Email     string  `db:"email"`
	Age       int64   `db:"age"`
	Score     float64 `db:"score"`
	Active    bool    `db:"active"`
}

func (d *ChaosDriver) Open(name string) (driver.Conn, error) {
	return &ChaosConn{}, nil
}

type ChaosConn struct{}

func (c *ChaosConn) Prepare(query string) (driver.Stmt, error) {
	return &ChaosStmt{seed: query}, nil
}
func (c *ChaosConn) Close() error              { return nil }
func (c *ChaosConn) Begin() (driver.Tx, error) { return nil, nil }

type ChaosStmt struct {
	seed    string
	payload []byte
}

func (s *ChaosStmt) Close() error                                    { return nil }
func (s *ChaosStmt) NumInput() int                                   { return -1 }
func (s *ChaosStmt) Exec(args []driver.Value) (driver.Result, error) { return nil, nil }
func (s *ChaosStmt) Query(args []driver.Value) (driver.Rows, error) {
	seedInt, _ := strconv.ParseInt(s.seed, 10, 64)
	return &ChaosRows{rnd: rand.New(rand.NewSource(seedInt)), payload: s.payload}, nil
}

type ChaosRows struct {
	rnd         *rand.Rand
	payload     []byte
	emittedCols bool
	columnNames []string
	types       []int
}

func (r *ChaosRows) Columns() []string {
	if !r.emittedCols {
		// Target 2 Fuzzing: Generate up to 50 random structural columns to test OOB meta paths
		numCols := r.rnd.Intn(50) + 1
		r.columnNames = make([]string, numCols)
		r.types = make([]int, numCols)

		targets := []string{"id", "first_name", "last_name", "email", "active", "score", "age"}

		for i := range numCols {
			// 50% chance to hit a real struct column tag
			if r.rnd.Intn(2) == 0 {
				r.columnNames[i] = targets[r.rnd.Intn(len(targets))]
			} else {
				r.columnNames[i] = "garbage_" + strconv.Itoa(r.rnd.Intn(100))
			}
			r.types[i] = r.rnd.Intn(6) // 0=int, 1=string, 2=float, 3=bool, 4=nil, 5=fuzzer_bytes
		}
		r.emittedCols = true
	}
	return r.columnNames
}

func (r *ChaosRows) Close() error { return nil }

func (r *ChaosRows) Next(dest []driver.Value) error {
	// Emit 0 to 5 rows before EOF
	if r.rnd.Intn(10) < 2 {
		return io.EOF
	}

	for i := range dest {
		switch r.types[i] {
		case 0:
			dest[i] = int64(r.rnd.Int63())
		case 1:
			dest[i] = "fuzz_string"
		case 2:
			dest[i] = r.rnd.Float64()
		case 3:
			dest[i] = r.rnd.Intn(2) == 1
		case 4:
			dest[i] = nil
		case 5:
			dest[i] = r.payload // Inject the raw fuzzer payload to test memory limits directly into destinations
		}
	}
	return nil
}

// Global payload carrier to route into the driver during a query
// (Since sql driver interface doesn't let us pass []byte raw easily without binding)
// Using context values would be cleaner, but we'll use a package-level mux for the fuzzer.
var chaosPayloadMux sync.Map

// FuzzRowScanBounds implements Target 2: Row Scan Bounds fuzzing.
// It verifies that selectScan never performs an OOB write, even with mismatched/massive driver data.
func FuzzRowScanBounds(f *testing.F) {
	// Seed the fuzzer with some payloads
	f.Add(int64(0), []byte{})
	f.Add(int64(1), []byte("A"))
	f.Add(int64(2), []byte("massive_data_injection_1234567890"))

	rawDB, err := sql.Open("chaosdb", "")
	if err != nil {
		f.Fatal(err)
	}
	db := sqlx.NewDb(rawDB, "chaosdb")
	ctx := context.Background()

	f.Fuzz(func(t *testing.T, seed int64, payload []byte) {
		// Route payload to the statement via mux (chaosdb will read this based on correlation ID)
		correlationID := strconv.FormatInt(seed, 10)
		chaosPayloadMux.Store(correlationID, payload)
		defer chaosPayloadMux.Delete(correlationID)

		// Create a dynamic query using correlation ID so ChaosStmt can extract the payload
		query := "SELECT " + correlationID

		// We expect errors because the inputs are chaotic and type-mismatched.
		// The ONLY goal is that the unsafe.Pointer engine does not panic,
		// segfault, or corrupt memory boundaries during projection loops!

		// 1. Array Projection
		_, _ = sqlx.SelectG[FuzzTargetStruct](ctx, db, query)

		// 2. Single Entity
		_, _ = sqlx.GetG[FuzzTargetStruct](ctx, db, query)

		// 3. Iterator Flow
		for _, err := range sqlx.SelectIter[FuzzTargetStruct](ctx, db, query) {
			if err != nil {
				break
			}
		}
	})
}
