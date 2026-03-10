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
	"strconv"
	"strings"
	"testing"

	v1sqlx "github.com/jmoiron/sqlx"
	sqlx "github.com/rpadovani/sqlx-v2"
	_ "github.com/rpadovani/sqlx-v2/internal/mockdb"
)

var (
	GlobalSmallSlice  []SmallStruct
	GlobalMediumSlice []MediumStruct
	GlobalLargeSlice  []LargeStruct
	GlobalSmall       SmallStruct
	GlobalMedium      MediumStruct
	GlobalLarge       LargeStruct
	GlobalMap         map[string]any
	GlobalSlice       []any
	GlobalError       error
	GlobalResult      any
)

type SmallStruct struct {
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

type MediumStruct struct {
	ID        int64   `db:"id"`
	FirstName string  `db:"first_name"`
	LastName  string  `db:"last_name"`
	Email     string  `db:"email"`
	Phone     string  `db:"phone"`
	Address   string  `db:"address"`
	City      string  `db:"city"`
	State     string  `db:"state"`
	Zip       string  `db:"zip"`
	Country   string  `db:"country"`
	Age       int64   `db:"age"`
	Score     float64 `db:"score"`
	Active    bool    `db:"active"`
	CreatedAt string  `db:"created_at"`
	UpdatedAt string  `db:"updated_at"`
}

type NestedStruct1 struct {
	Field10 string  `db:"field_10"`
	Field11 float64 `db:"field_11"`
	Field12 int64   `db:"field_12"`
	Field13 string  `db:"field_13"`
	Field14 float64 `db:"field_14"`
}

type NestedStruct2 struct {
	Field20 float64 `db:"field_20"`
	Field21 int64   `db:"field_21"`
	Field22 string  `db:"field_22"`
	Field23 float64 `db:"field_23"`
	Field24 int64   `db:"field_24"`
}

type EmbeddedStruct struct {
	Field30 int64   `db:"field_30"`
	Field31 string  `db:"field_31"`
	Field32 float64 `db:"field_32"`
	Field33 int64   `db:"field_33"`
	Field34 string  `db:"field_34"`
}

type LargeStruct struct {
	Field00 int64   `db:"field_00"`
	Field01 string  `db:"field_01"`
	Field02 float64 `db:"field_02"`
	Field03 int64   `db:"field_03"`
	Field04 string  `db:"field_04"`
	Field05 float64 `db:"field_05"`
	Field06 int64   `db:"field_06"`
	Field07 string  `db:"field_07"`
	Field08 float64 `db:"field_08"`
	Field09 int64   `db:"field_09"`

	Nested1 NestedStruct1 `db:""`

	Field15 int64   `db:"field_15"`
	Field16 string  `db:"field_16"`
	Field17 float64 `db:"field_17"`
	Field18 int64   `db:"field_18"`
	Field19 string  `db:"field_19"`

	NestedPtr *NestedStruct2 `db:""`

	Field25 string  `db:"field_25"`
	Field26 float64 `db:"field_26"`
	Field27 int64   `db:"field_27"`
	Field28 string  `db:"field_28"`
	Field29 float64 `db:"field_29"`

	*EmbeddedStruct

	Field35 float64 `db:"field_35"`
	Field36 int64   `db:"field_36"`
	Field37 string  `db:"field_37"`
	Field38 float64 `db:"field_38"`
	Field39 int64   `db:"field_39"`

	Field40 string  `db:"field_40"`
	Field41 float64 `db:"field_41"`
	Field42 int64   `db:"field_42"`
	Field43 string  `db:"field_43"`
	Field44 float64 `db:"field_44"`
	Field45 int64   `db:"field_45"`
	Field46 string  `db:"field_46"`
	Field47 float64 `db:"field_47"`
	Field48 int64   `db:"field_48"`
	Field49 string  `db:"field_49"`
}

func getDB_V1(b *testing.B) *v1sqlx.DB {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		b.Fatal(err)
	}
	return v1sqlx.NewDb(rawDB, "mockdb")
}

func getDB_V2(b *testing.B) *sqlx.DB {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		b.Fatal(err)
	}
	return sqlx.NewDb(rawDB, "mockdb")
}

func BenchmarkV1_Select_Small_1(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var results []SmallStruct
	_ = db.Select(&results, "SELECT small:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var results []SmallStruct
		err := db.Select(&results, "SELECT small:1")
		GlobalResult = results
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2_Select_Small_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var results []SmallStruct
	_ = db.Select(&results, "SELECT small:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var results []SmallStruct
		err := db.Select(&results, "SELECT small:1")
		GlobalResult = results
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2G_Select_Small_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	// Warmup
	_, _ = sqlx.SelectG[SmallStruct](ctx, db, "SELECT small:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res, err := sqlx.SelectG[SmallStruct](ctx, db, "SELECT small:1")
		GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2G_SelectIter_Small_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	// Warmup
	for _, err := range sqlx.SelectIter[SmallStruct](ctx, db, "SELECT small:1") {
		_ = err
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for val, err := range sqlx.SelectIter[SmallStruct](ctx, db, "SELECT small:1") {
			GlobalSmall = val
			GlobalError = err
		}
	}
}

func BenchmarkV1_Select_Medium_1(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var results []MediumStruct
	_ = db.Select(&results, "SELECT medium:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var results []MediumStruct
		err := db.Select(&results, "SELECT medium:1")
		GlobalResult = results
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2_Select_Medium_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var results []MediumStruct
	_ = db.Select(&results, "SELECT medium:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var results []MediumStruct
		err := db.Select(&results, "SELECT medium:1")
		GlobalResult = results
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2G_Select_Medium_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	// Warmup
	_, _ = sqlx.SelectG[MediumStruct](ctx, db, "SELECT medium:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res, err := sqlx.SelectG[MediumStruct](ctx, db, "SELECT medium:1")
		GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2G_SelectIter_Medium_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	// Warmup
	for _, err := range sqlx.SelectIter[MediumStruct](ctx, db, "SELECT medium:1") {
		_ = err
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for val, err := range sqlx.SelectIter[MediumStruct](ctx, db, "SELECT medium:1") {
			GlobalMedium = val
			GlobalError = err
		}
	}
}

func BenchmarkV1_Select_Large_1(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var results []LargeStruct
	_ = db.Select(&results, "SELECT large:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var results []LargeStruct
		err := db.Select(&results, "SELECT large:1")
		GlobalResult = results
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2_Select_Large_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var results []LargeStruct
	_ = db.Select(&results, "SELECT large:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var results []LargeStruct
		err := db.Select(&results, "SELECT large:1")
		GlobalResult = results
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2G_Select_Large_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	// Warmup
	_, _ = sqlx.SelectG[LargeStruct](ctx, db, "SELECT large:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res, err := sqlx.SelectG[LargeStruct](ctx, db, "SELECT large:1")
		GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2G_SelectIter_Large_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	// Warmup
	for _, err := range sqlx.SelectIter[LargeStruct](ctx, db, "SELECT large:1") {
		_ = err
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for val, err := range sqlx.SelectIter[LargeStruct](ctx, db, "SELECT large:1") {
			GlobalLarge = val
			GlobalError = err
		}
	}
}

func BenchmarkV1_Select_Small_1000(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var results []SmallStruct
	_ = db.Select(&results, "SELECT small:1000")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var results []SmallStruct
		err := db.Select(&results, "SELECT small:1000")
		GlobalResult = results
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2_Select_Small_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var results []SmallStruct
	_ = db.Select(&results, "SELECT small:1000")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var results []SmallStruct
		err := db.Select(&results, "SELECT small:1000")
		GlobalResult = results
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2G_Select_Small_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	// Warmup
	_, _ = sqlx.SelectG[SmallStruct](ctx, db, "SELECT small:1000")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res, err := sqlx.SelectG[SmallStruct](ctx, db, "SELECT small:1000")
		GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2G_SelectIter_Small_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	// Warmup
	for _, err := range sqlx.SelectIter[SmallStruct](ctx, db, "SELECT small:1000") {
		_ = err
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for val, err := range sqlx.SelectIter[SmallStruct](ctx, db, "SELECT small:1000") {
			GlobalSmall = val
			GlobalError = err
		}
	}
}

func BenchmarkV1_Select_Medium_1000(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var results []MediumStruct
	_ = db.Select(&results, "SELECT medium:1000")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var results []MediumStruct
		err := db.Select(&results, "SELECT medium:1000")
		GlobalResult = results
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2_Select_Medium_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var results []MediumStruct
	_ = db.Select(&results, "SELECT medium:1000")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var results []MediumStruct
		err := db.Select(&results, "SELECT medium:1000")
		GlobalResult = results
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2G_Select_Medium_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	// Warmup
	_, _ = sqlx.SelectG[MediumStruct](ctx, db, "SELECT medium:1000")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res, err := sqlx.SelectG[MediumStruct](ctx, db, "SELECT medium:1000")
		GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2G_SelectIter_Medium_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	// Warmup
	for _, err := range sqlx.SelectIter[MediumStruct](ctx, db, "SELECT medium:1000") {
		_ = err
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for val, err := range sqlx.SelectIter[MediumStruct](ctx, db, "SELECT medium:1000") {
			GlobalMedium = val
			GlobalError = err
		}
	}
}

func BenchmarkV1_Select_Large_1000(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var results []LargeStruct
	_ = db.Select(&results, "SELECT large:1000")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var results []LargeStruct
		err := db.Select(&results, "SELECT large:1000")
		GlobalResult = results
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2_Select_Large_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var results []LargeStruct
	_ = db.Select(&results, "SELECT large:1000")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var results []LargeStruct
		err := db.Select(&results, "SELECT large:1000")
		GlobalResult = results
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2G_Select_Large_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	// Warmup
	_, _ = sqlx.SelectG[LargeStruct](ctx, db, "SELECT large:1000")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res, err := sqlx.SelectG[LargeStruct](ctx, db, "SELECT large:1000")
		GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2G_SelectIter_Large_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	// Warmup
	for _, err := range sqlx.SelectIter[LargeStruct](ctx, db, "SELECT large:1000") {
		_ = err
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for val, err := range sqlx.SelectIter[LargeStruct](ctx, db, "SELECT large:1000") {
			GlobalLarge = val
			GlobalError = err
		}
	}
}

func BenchmarkV1_Get_Small(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var result SmallStruct
	_ = db.Get(&result, "SELECT small:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result SmallStruct
		err := db.Get(&result, "SELECT small:1")
		GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}
func BenchmarkV2_Get_Small(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var result SmallStruct
	_ = db.Get(&result, "SELECT small:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result SmallStruct
		err := db.Get(&result, "SELECT small:1")
		GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}
func BenchmarkV2G_Get_Small(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	// Warmup
	_, _ = sqlx.GetG[SmallStruct](ctx, db, "SELECT small:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res, err := sqlx.GetG[SmallStruct](ctx, db, "SELECT small:1")
		GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV1_Get_Medium(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var result MediumStruct
	_ = db.Get(&result, "SELECT medium:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result MediumStruct
		err := db.Get(&result, "SELECT medium:1")
		GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}
func BenchmarkV2_Get_Medium(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var result MediumStruct
	_ = db.Get(&result, "SELECT medium:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result MediumStruct
		err := db.Get(&result, "SELECT medium:1")
		GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}
func BenchmarkV2G_Get_Medium(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	// Warmup
	_, _ = sqlx.GetG[MediumStruct](ctx, db, "SELECT medium:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res, err := sqlx.GetG[MediumStruct](ctx, db, "SELECT medium:1")
		GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV1_Get_Large(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var result LargeStruct
	_ = db.Get(&result, "SELECT large:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result LargeStruct
		err := db.Get(&result, "SELECT large:1")
		GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}
func BenchmarkV2_Get_Large(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	var result LargeStruct
	_ = db.Get(&result, "SELECT large:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var result LargeStruct
		err := db.Get(&result, "SELECT large:1")
		GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}
func BenchmarkV2G_Get_Large(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()
	// Warmup
	_, _ = sqlx.GetG[LargeStruct](ctx, db, "SELECT large:1")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		res, err := sqlx.GetG[LargeStruct](ctx, db, "SELECT large:1")
		GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV1_StructScan_Small_1(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT small:1")
	var result SmallStruct
	_ = rows.StructScan(&result)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT small:1")
		for rows.Next() {
			var result SmallStruct
			err := rows.StructScan(&result)
			GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_StructScan_Small_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT small:1")
	var result SmallStruct
	_ = rows.StructScan(&result)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT small:1")
		for rows.Next() {
			var result SmallStruct
			err := rows.StructScan(&result)
			GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_MapScan_Small_1(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT small:1")
	dest := make(map[string]any)
	_ = rows.MapScan(dest)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT small:1")
		for rows.Next() {
			dest := make(map[string]any)
			err := rows.MapScan(dest)
			GlobalMap = dest
			GlobalError = err
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_MapScan_Small_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT small:1")
	dest := make(map[string]any)
	_ = rows.MapScan(dest)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT small:1")
		for rows.Next() {
			dest := make(map[string]any)
			err := rows.MapScan(dest)
			GlobalMap = dest
			GlobalError = err
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_SliceScan_Small_1(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT small:1")
	_, _ = rows.SliceScan()
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT small:1")
		for rows.Next() {
			res, err := rows.SliceScan()
			GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_SliceScan_Small_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT small:1")
	_, _ = rows.SliceScan()
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT small:1")
		for rows.Next() {
			res, err := rows.SliceScan()
			GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_StructScan_Medium_1(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT medium:1")
	var result MediumStruct
	_ = rows.StructScan(&result)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT medium:1")
		for rows.Next() {
			var result MediumStruct
			err := rows.StructScan(&result)
			GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_StructScan_Medium_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT medium:1")
	var result MediumStruct
	_ = rows.StructScan(&result)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT medium:1")
		for rows.Next() {
			var result MediumStruct
			err := rows.StructScan(&result)
			GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_MapScan_Medium_1(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT medium:1")
	dest := make(map[string]any)
	_ = rows.MapScan(dest)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT medium:1")
		for rows.Next() {
			dest := make(map[string]any)
			err := rows.MapScan(dest)
			GlobalMap = dest
			GlobalError = err
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_MapScan_Medium_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT medium:1")
	dest := make(map[string]any)
	_ = rows.MapScan(dest)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT medium:1")
		for rows.Next() {
			dest := make(map[string]any)
			err := rows.MapScan(dest)
			GlobalMap = dest
			GlobalError = err
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_SliceScan_Medium_1(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT medium:1")
	_, _ = rows.SliceScan()
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT medium:1")
		for rows.Next() {
			res, err := rows.SliceScan()
			GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_SliceScan_Medium_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT medium:1")
	_, _ = rows.SliceScan()
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT medium:1")
		for rows.Next() {
			res, err := rows.SliceScan()
			GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_StructScan_Large_1(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT large:1")
	var result LargeStruct
	_ = rows.StructScan(&result)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT large:1")
		for rows.Next() {
			var result LargeStruct
			err := rows.StructScan(&result)
			GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_StructScan_Large_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT large:1")
	var result LargeStruct
	_ = rows.StructScan(&result)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT large:1")
		for rows.Next() {
			var result LargeStruct
			err := rows.StructScan(&result)
			GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_MapScan_Large_1(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT large:1")
	dest := make(map[string]any)
	_ = rows.MapScan(dest)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT large:1")
		for rows.Next() {
			dest := make(map[string]any)
			err := rows.MapScan(dest)
			GlobalMap = dest
			GlobalError = err
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_MapScan_Large_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT large:1")
	dest := make(map[string]any)
	_ = rows.MapScan(dest)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT large:1")
		for rows.Next() {
			dest := make(map[string]any)
			err := rows.MapScan(dest)
			GlobalMap = dest
			GlobalError = err
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_SliceScan_Large_1(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT large:1")
	_, _ = rows.SliceScan()
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT large:1")
		for rows.Next() {
			res, err := rows.SliceScan()
			GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_SliceScan_Large_1(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT large:1")
	_, _ = rows.SliceScan()
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT large:1")
		for rows.Next() {
			res, err := rows.SliceScan()
			GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_StructScan_Small_1000(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT small:1000")
	var result SmallStruct
	_ = rows.StructScan(&result)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT small:1000")
		for rows.Next() {
			var result SmallStruct
			err := rows.StructScan(&result)
			GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_StructScan_Small_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT small:1000")
	var result SmallStruct
	_ = rows.StructScan(&result)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT small:1000")
		for rows.Next() {
			var result SmallStruct
			err := rows.StructScan(&result)
			GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_MapScan_Small_1000(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT small:1000")
	dest := make(map[string]any)
	_ = rows.MapScan(dest)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT small:1000")
		for rows.Next() {
			dest := make(map[string]any)
			err := rows.MapScan(dest)
			GlobalMap = dest
			GlobalError = err
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_MapScan_Small_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT small:1000")
	dest := make(map[string]any)
	_ = rows.MapScan(dest)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT small:1000")
		for rows.Next() {
			dest := make(map[string]any)
			err := rows.MapScan(dest)
			GlobalMap = dest
			GlobalError = err
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_SliceScan_Small_1000(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT small:1000")
	_, _ = rows.SliceScan()
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT small:1000")
		for rows.Next() {
			res, err := rows.SliceScan()
			GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_SliceScan_Small_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT small:1000")
	_, _ = rows.SliceScan()
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT small:1000")
		for rows.Next() {
			res, err := rows.SliceScan()
			GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_StructScan_Medium_1000(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT medium:1000")
	var result MediumStruct
	_ = rows.StructScan(&result)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT medium:1000")
		for rows.Next() {
			var result MediumStruct
			err := rows.StructScan(&result)
			GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_StructScan_Medium_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT medium:1000")
	var result MediumStruct
	_ = rows.StructScan(&result)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT medium:1000")
		for rows.Next() {
			var result MediumStruct
			err := rows.StructScan(&result)
			GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_MapScan_Medium_1000(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT medium:1000")
	dest := make(map[string]any)
	_ = rows.MapScan(dest)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT medium:1000")
		for rows.Next() {
			dest := make(map[string]any)
			err := rows.MapScan(dest)
			GlobalMap = dest
			GlobalError = err
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_MapScan_Medium_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT medium:1000")
	dest := make(map[string]any)
	_ = rows.MapScan(dest)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT medium:1000")
		for rows.Next() {
			dest := make(map[string]any)
			err := rows.MapScan(dest)
			GlobalMap = dest
			GlobalError = err
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_SliceScan_Medium_1000(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT medium:1000")
	_, _ = rows.SliceScan()
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT medium:1000")
		for rows.Next() {
			res, err := rows.SliceScan()
			GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_SliceScan_Medium_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT medium:1000")
	_, _ = rows.SliceScan()
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT medium:1000")
		for rows.Next() {
			res, err := rows.SliceScan()
			GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_StructScan_Large_1000(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT large:1000")
	var result LargeStruct
	_ = rows.StructScan(&result)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT large:1000")
		for rows.Next() {
			var result LargeStruct
			err := rows.StructScan(&result)
			GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_StructScan_Large_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT large:1000")
	var result LargeStruct
	_ = rows.StructScan(&result)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT large:1000")
		for rows.Next() {
			var result LargeStruct
			err := rows.StructScan(&result)
			GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_MapScan_Large_1000(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT large:1000")
	dest := make(map[string]any)
	_ = rows.MapScan(dest)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT large:1000")
		for rows.Next() {
			dest := make(map[string]any)
			err := rows.MapScan(dest)
			GlobalMap = dest
			GlobalError = err
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_MapScan_Large_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT large:1000")
	dest := make(map[string]any)
	_ = rows.MapScan(dest)
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT large:1000")
		for rows.Next() {
			dest := make(map[string]any)
			err := rows.MapScan(dest)
			GlobalMap = dest
			GlobalError = err
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_SliceScan_Large_1000(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT large:1000")
	_, _ = rows.SliceScan()
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT large:1000")
		for rows.Next() {
			res, err := rows.SliceScan()
			GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}
func BenchmarkV2_SliceScan_Large_1000(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	rows, _ := db.Queryx("SELECT large:1000")
	_, _ = rows.SliceScan()
	_ = rows.Close()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rows, _ := db.Queryx("SELECT large:1000")
		for rows.Next() {
			res, err := rows.SliceScan()
			GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
		}
		_ = rows.Close()
	}
}

func BenchmarkV1_Exec(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	_, _ = db.Exec("INSERT INTO foo VALUES (1)")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := db.Exec("INSERT INTO foo VALUES (1)")
		GlobalError = err
	}
}
func BenchmarkV2_Exec(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	_, _ = db.Exec("INSERT INTO foo VALUES (1)")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := db.Exec("INSERT INTO foo VALUES (1)")
		GlobalError = err
	}
}

func BenchmarkV1_MustExec(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		db.MustExec("INSERT INTO foo VALUES (1)")
	}
}
func BenchmarkV2_MustExec(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		db.MustExec("INSERT INTO foo VALUES (1)")
	}
}

func BenchmarkV1_NamedExec(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	_, _ = db.NamedExec("INSERT INTO foo VALUES (:id)", &SmallStruct{})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := db.NamedExec("INSERT INTO foo VALUES (:id)", &SmallStruct{})
		GlobalError = err
	}
}
func BenchmarkV2_NamedExec(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	_, _ = db.NamedExec("INSERT INTO foo VALUES (:id)", &SmallStruct{})

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := db.NamedExec("INSERT INTO foo VALUES (:id)", &SmallStruct{})
		GlobalError = err
	}
}

func BenchmarkV2_NamedExec_Cold(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// Append unique comment to force compiled query cache miss
		q := "INSERT INTO foo VALUES (:id) /* " + strconv.Itoa(i) + " */"
		_, err := db.NamedExec(q, &SmallStruct{})
		GlobalError = err
	}
}

func BenchmarkV1_Prepare(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	_, _ = db.Prepare("INSERT INTO foo VALUES (1)")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		stmt, err := db.Prepare("INSERT INTO foo VALUES (1)")
		if stmt != nil {
			_ = stmt.Close()
		}
		GlobalError = err
	}
}
func BenchmarkV2_Prepare(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	_, _ = db.Prepare("INSERT INTO foo VALUES (1)")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		stmt, err := db.Prepare("INSERT INTO foo VALUES (1)")
		if stmt != nil {
			_ = stmt.Close()
		}
		GlobalError = err
	}
}

func BenchmarkV1_PrepareNamed(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	_, _ = db.PrepareNamed("INSERT INTO foo VALUES (:id)")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		stmt, err := db.PrepareNamed("INSERT INTO foo VALUES (:id)")
		if stmt != nil {
			_ = stmt.Close()
		}
		GlobalError = err
	}
}
func BenchmarkV2_PrepareNamed(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	_, _ = db.PrepareNamed("INSERT INTO foo VALUES (:id)")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		stmt, err := db.PrepareNamed("INSERT INTO foo VALUES (:id)")
		if stmt != nil {
			_ = stmt.Close()
		}
		GlobalError = err
	}
}

func BenchmarkV1_Preparex(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()
	// Warmup
	_, _ = db.Preparex("INSERT INTO foo VALUES (1)")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		stmt, err := db.Preparex("INSERT INTO foo VALUES (1)")
		if stmt != nil {
			_ = stmt.Close()
		}
		GlobalError = err
	}
}
func BenchmarkV2_Preparex(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	// Warmup
	_, _ = db.Preparex("INSERT INTO foo VALUES (1)")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		stmt, err := db.Preparex("INSERT INTO foo VALUES (1)")
		if stmt != nil {
			_ = stmt.Close()
		}
		GlobalError = err
	}
}

func BenchmarkV1_Get_Small_Cold(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		db.MapperFunc(strings.ToLower) // clear cache
		var result SmallStruct
		err := db.Get(&result, "SELECT small:1")
		GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}
func BenchmarkV2_Get_Small_Cold(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		db.MapperFunc(strings.ToLower) // clear cache
		var result SmallStruct
		err := db.Get(&result, "SELECT small:1")
		GlobalResult = result
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}
func BenchmarkV2G_Get_Small_Cold(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		db.MapperFunc(strings.ToLower) // clear cache
		res, err := sqlx.GetG[SmallStruct](ctx, db, "SELECT small:1")
		GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV1_Select_Small_1000_Cold(b *testing.B) {
	db := getDB_V1(b)
	defer func() { _ = db.Close() }()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		db.MapperFunc(strings.ToLower) // clear cache
		var results []SmallStruct
		err := db.Select(&results, "SELECT small:1000")
		GlobalResult = results
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2_Select_Small_1000_Cold(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		db.MapperFunc(strings.ToLower) // clear cache
		var results []SmallStruct
		err := db.Select(&results, "SELECT small:1000")
		GlobalResult = results
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}

func BenchmarkV2G_Select_Small_1000_Cold(b *testing.B) {
	db := getDB_V2(b)
	defer func() { _ = db.Close() }()
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		db.MapperFunc(strings.ToLower) // clear cache
		res, err := sqlx.SelectG[SmallStruct](ctx, db, "SELECT small:1000")
		GlobalResult = res
		if err != nil {
			b.Fatalf("benchmark failed: %v", err)
		}
	}
}
