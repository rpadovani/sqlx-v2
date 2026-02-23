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
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/rpadovani/sqlx-v2/internal/shadow"
)

func newSQLiteDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	return db
}

func TestShadow_TagParsing(t *testing.T) {
	db := newSQLiteDB(t)
	defer func() { _ = db.Close() }()
	h := shadow.NewHarness(t, db, "sqlite", false)
	defer h.Close()
	h.TestTagParsing(t)
}

func TestShadow_TagParsing_Strict(t *testing.T) {
	db := newSQLiteDB(t)
	defer func() { _ = db.Close() }()
	h := shadow.NewHarness(t, db, "sqlite", true)
	defer h.Close()
	h.TestTagParsing(t)
}

func TestShadow_NullHandling(t *testing.T) {
	db := newSQLiteDB(t)
	defer func() { _ = db.Close() }()
	h := shadow.NewHarness(t, db, "sqlite", false)
	defer h.Close()
	h.TestNullHandling(t)
}

func TestShadow_ScanBehavior(t *testing.T) {
	db := newSQLiteDB(t)
	defer func() { _ = db.Close() }()
	h := shadow.NewHarness(t, db, "sqlite", false)
	defer h.Close()
	h.TestScanBehavior(t)
}

func TestShadow_ConnectionWrapping(t *testing.T) {
	db := newSQLiteDB(t)
	defer func() { _ = db.Close() }()
	h := shadow.NewHarness(t, db, "sqlite", false)
	defer h.Close()
	h.TestConnectionWrapping(t)
}
