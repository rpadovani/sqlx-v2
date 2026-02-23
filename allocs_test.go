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
	"fmt"
	"testing"

	"github.com/rpadovani/sqlx-v2"
	_ "github.com/rpadovani/sqlx-v2/internal/mockdb"
)

// LargeStruct and SmallStruct are used from bench_test.go

// TestAllocations_SelectIter asserts that SelectIter maintains a low,
// constant number of allocations per row, regardless of struct size.
func TestAllocations_SelectIter(t *testing.T) {
	db, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	sdb := sqlx.NewDb(db, "mockdb")
	ctx := context.Background()

	// Warm up to populate reflection caches and pool
	for range sqlx.SelectIter[LargeStruct](ctx, sdb, "SELECT large:1") {
	}

	// We run 1 call with many rows to see amortized cost
	numRows := 100
	allocs := testing.AllocsPerRun(10, func() {
		for s, err := range sqlx.SelectIter[LargeStruct](ctx, sdb, fmt.Sprintf("SELECT large:%d", numRows)) {
			if err != nil {
				t.Fatal(err)
			}
			_ = s
		}
	})

	perRow := allocs / float64(numRows)
	fmt.Printf("SelectIter LargeStruct (50 fields) allocs per row: %v\n", perRow)

	// Threshold: ~6 allocations per row.
	// Reflection-based mapping would be 50+ per row.
	if perRow > 6.0 {
		t.Errorf("SelectIter too many allocations: %v per row", perRow)
	}
}

// TestAllocations_SelectG asserts that SelectG (slice-based) is efficient.
func TestAllocations_SelectG(t *testing.T) {
	db, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	sdb := sqlx.NewDb(db, "mockdb")
	ctx := context.Background()

	// Warm up
	_, _ = sqlx.SelectG[LargeStruct](ctx, sdb, "SELECT large:1")

	numRows := 100
	allocs := testing.AllocsPerRun(10, func() {
		res, err := sqlx.SelectG[LargeStruct](ctx, sdb, fmt.Sprintf("SELECT large:%d", numRows))
		if err != nil {
			t.Fatal(err)
		}
		_ = res
	})

	perRow := allocs / float64(numRows)
	fmt.Printf("SelectG LargeStruct (50 fields) allocs per row: %v\n", perRow)

	// Threshold: SelectG should also be around ~3 per row.
	// Write barrier fix adds ~0.4 allocs per row.
	if perRow > 6.0 {
		t.Errorf("SelectG too many allocations: %v per row", perRow)
	}
}

// TestAllocations_StructScan acts as an anti-regression test for the "Zero-Allocation Pointer Hopping" logic.
// It verifies that rows.StructScan does not allocate intermediate reflect.Value wrappers during the hot loop.
func TestAllocations_StructScan(t *testing.T) {
	db, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	sdb := sqlx.NewDb(db, "mockdb")

	// Warm up
	rows, _ := sdb.Queryx("SELECT large:1")
	for rows.Next() {
		var result LargeStruct
		_ = rows.StructScan(&result)
	}
	_ = rows.Close()

	numRows := 100
	allocs := testing.AllocsPerRun(10, func() {
		rows, err := sdb.Queryx(fmt.Sprintf("SELECT large:%d", numRows))
		if err != nil {
			t.Fatal(err)
		}
		for rows.Next() {
			var result LargeStruct
			if err := rows.StructScan(&result); err != nil {
				t.Fatal(err)
			}
		}
		_ = rows.Close()
	})

	perRow := allocs / float64(numRows)
	fmt.Printf("Rows.StructScan LargeStruct (50 fields) allocs per row: %v\n", perRow)

	// Threshold: Due to the zero-allocation optimization, struct scanning a large row
	// should take < 2.0 allocations per row. We expect ~1 allocation per loop iteration.
	if perRow > 6.0 {
		t.Errorf("Rows.StructScan too many allocations: %v per row. Zero-Allocation optimization may have regressed.", perRow)
	}
}
