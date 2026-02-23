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
	"runtime"
	"runtime/debug"
	"testing"

	sqlx "github.com/rpadovani/sqlx-v2"
	_ "github.com/rpadovani/sqlx-v2/internal/mockdb"
)

// TestGCSafety_SelectUnderPressure verifies that scanned struct values survive
// garbage collection during the scan loop. This catches missing runtime.KeepAlive
// calls that could allow the GC to collect reflect.Value allocations before Scan
// writes through the derived unsafe.Pointer.
//
// Remediation for: Mutation 2 (KeepAlive Removal) in the Deep Truth audit.
func TestGCSafety_SelectUnderPressure(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mockdb")

	// Maximize GC pressure: collect after every allocation
	oldGC := debug.SetGCPercent(1)
	defer debug.SetGCPercent(oldGC)

	const numRows = 50000

	// Test 1: Classic Select path (iterateScan → selectScan)
	t.Run("Select", func(t *testing.T) {
		for iter := range 3 {
			// Force GC + return memory to OS between iterations
			runtime.GC()
			debug.FreeOSMemory()

			var results []MediumStruct
			err := db.Select(&results, "SELECT medium:50000")
			if err != nil {
				t.Fatalf("iteration %d: Select error: %v", iter, err)
			}
			if len(results) != numRows {
				t.Fatalf("iteration %d: expected %d rows, got %d", iter, numRows, len(results))
			}

			// Validate every row — a GC-collected struct would have zero values
			for i, r := range results {
				if r.FirstName != "John" {
					t.Fatalf("iteration %d, row %d: expected FirstName='John', got '%s' (possible GC corruption)", iter, i, r.FirstName)
				}
				if r.Email != "john@example.com" {
					t.Fatalf("iteration %d, row %d: expected Email='john@example.com', got '%s' (possible GC corruption)", iter, i, r.Email)
				}
				if r.ID != 1 {
					t.Fatalf("iteration %d, row %d: expected ID=1, got %d (possible GC corruption)", iter, i, r.ID)
				}
				if r.Score != 95.5 {
					t.Fatalf("iteration %d, row %d: expected Score=95.5, got %f (possible GC corruption)", iter, i, r.Score)
				}
				if r.LastName != "Doe" {
					t.Fatalf("iteration %d, row %d: expected LastName='Doe', got '%s' (possible GC corruption)", iter, i, r.LastName)
				}
				if r.Active != true {
					t.Fatalf("iteration %d, row %d: expected Active=true, got false (possible GC corruption)", iter, i)
				}
			}
		}
	})

	// Test 2: Generic SelectG path
	t.Run("SelectG", func(t *testing.T) {
		ctx := context.Background()
		for iter := range 3 {
			runtime.GC()
			debug.FreeOSMemory()

			results, err := sqlx.SelectG[MediumStruct](ctx, db, "SELECT medium:50000")
			if err != nil {
				t.Fatalf("iteration %d: SelectG error: %v", iter, err)
			}
			if len(results) != numRows {
				t.Fatalf("iteration %d: expected %d rows, got %d", iter, numRows, len(results))
			}

			for i, r := range results {
				if r.FirstName != "John" {
					t.Fatalf("iteration %d, row %d: expected FirstName='John', got '%s' (possible GC corruption)", iter, i, r.FirstName)
				}
				if r.Score != 95.5 {
					t.Fatalf("iteration %d, row %d: expected Score=95.5, got %f (possible GC corruption)", iter, i, r.Score)
				}
				if r.Active != true {
					t.Fatalf("iteration %d, row %d: expected Active=true, got false (possible GC corruption)", iter, i)
				}
			}
		}
	})

	// Test 3: Get path (single row) — high iteration count to stress KeepAlive
	t.Run("Get", func(t *testing.T) {
		for iter := range 100 {
			runtime.GC()
			debug.FreeOSMemory()

			var result MediumStruct
			err := db.Get(&result, "SELECT medium:1")
			if err != nil {
				t.Fatalf("iteration %d: Get error: %v", iter, err)
			}
			if result.FirstName != "John" {
				t.Fatalf("iteration %d: expected FirstName='John', got '%s' (possible GC corruption)", iter, result.FirstName)
			}
			if result.Active != true {
				t.Fatalf("iteration %d: expected Active=true, got false (possible GC corruption)", iter)
			}
		}
	})

	// Test 4: StructScan path via Rows — GC inside the scan loop
	t.Run("StructScan", func(t *testing.T) {
		for iter := range 5 {
			runtime.GC()
			debug.FreeOSMemory()

			rows, err := db.Queryx("SELECT medium:10000")
			if err != nil {
				t.Fatalf("iteration %d: Queryx error: %v", iter, err)
			}

			count := 0
			for rows.Next() {
				var result MediumStruct
				if err := rows.StructScan(&result); err != nil {
					_ = rows.Close()
					t.Fatalf("iteration %d, row %d: StructScan error: %v", iter, count, err)
				}
				if result.FirstName != "John" {
					_ = rows.Close()
					t.Fatalf("iteration %d, row %d: expected FirstName='John', got '%s' (possible GC corruption)", iter, count, result.FirstName)
				}
				count++
				// Force GC + FreeOSMemory every 100 rows to stress KeepAlive
				if count%100 == 0 {
					runtime.GC()
					debug.FreeOSMemory()
				}
			}
			_ = rows.Close()

			if count != 10000 {
				t.Fatalf("iteration %d: expected 10000 rows, got %d", iter, count)
			}
		}
	})
}
