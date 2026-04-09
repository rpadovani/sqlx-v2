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

package shadow

import (
	"database/sql"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"

	"github.com/rpadovani/sqlx-v2/tests/internal/testutil"
)

func TestRealWorldShadowParity_SQLite(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()

	harness := NewHarness(t, db, "sqlite", false)
	defer harness.Close()

	harness.TestTagParsing(t)
	harness.TestNullHandling(t)
	harness.TestScanBehavior(t)
	harness.TestConnectionWrapping(t)
}

func TestRealWorldShadowParity_Postgres(t *testing.T) {
	dsn := testutil.GlobalPostgres(t)
	if dsn == "skip" {
		t.Skip("Docker not available")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to open postgres: %v", err)
	}
	defer func() { _ = db.Close() }()

	harness := NewHarness(t, db, "pgx", false)
	defer harness.Close()

	harness.TestTagParsing(t)
	harness.TestNullHandling(t)
	harness.TestScanBehavior(t)
	harness.TestConnectionWrapping(t)
}

func TestRealWorldShadowParity_MySQL(t *testing.T) {
	dsn := testutil.GlobalMySQL(t)
	if dsn == "skip" {
		t.Skip("Docker not available")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to open mysql: %v", err)
	}
	defer func() { _ = db.Close() }()

	harness := NewHarness(t, db, "mysql", false)
	defer harness.Close()

	harness.TestTagParsing(t)
	harness.TestNullHandling(t)
	harness.TestScanBehavior(t)
	harness.TestConnectionWrapping(t)
}
