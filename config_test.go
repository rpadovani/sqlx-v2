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
	"strings"
	"testing"

	sqlx "github.com/rpadovani/sqlx-v2"
	_ "github.com/rpadovani/sqlx-v2/internal/mockdb"
)

// TestStrictTagParsing_RejectsMalformedTag verifies that StrictTagParsing
// causes an error when a struct has a tag with invalid characters.
func TestStrictTagParsing_RejectsMalformedTag(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mock")
	db.StrictTagParsing = true

	type BadStruct struct {
		Name string `db:"invalid name!"`
	}

	var res []BadStruct
	err = db.Select(&res, "SELECT 'foo' as `invalid name!`")
	if err == nil {
		t.Fatal("expected strict tag validation error, got none")
	}
	if !strings.Contains(err.Error(), "invalid character") {
		t.Fatalf("expected 'invalid character' in error, got: %s", err)
	}
}

// TestStrictTagParsing_RejectsEmptyTagName verifies that StrictTagParsing
// rejects tags like `db:","` where the name portion is empty.
func TestStrictTagParsing_RejectsEmptyTagName(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mock")
	db.StrictTagParsing = true

	type BadStruct struct {
		Name string `db:",omitempty"`
	}

	var res []BadStruct
	err = db.Select(&res, "SELECT 'foo' as `Name`")
	if err == nil {
		t.Fatal("expected strict tag validation error for empty tag name, got none")
	}
	if !strings.Contains(err.Error(), "empty tag name") {
		t.Fatalf("expected 'empty tag name' in error, got: %s", err)
	}
}

// TestStrictTagParsing_AcceptsValidTag verifies that valid tags pass strict validation.
func TestStrictTagParsing_AcceptsValidTag(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mock")
	db.StrictTagParsing = true
	db.Unsafe = true

	type GoodStruct struct {
		Name  string `db:"name"`
		Email string `db:"email"`
		ID    int    `db:"id"`
	}

	var res []GoodStruct
	err = db.Select(&res, "SELECT small:1")
	if err != nil {
		t.Fatalf("expected no errors for valid struct, got: %v", err)
	}
}

// TestStrictTagParsing_DefaultConfigIgnoresMalformed verifies that the default
// config (legacy mode) does NOT produce errors for malformed tags.
func TestStrictTagParsing_DefaultConfigIgnoresMalformed(t *testing.T) {
	rawDB, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rawDB.Close() }()

	db := sqlx.NewDb(rawDB, "mock")
	db.Unsafe = true

	type BadStruct struct {
		Name string `db:"invalid name!"`
	}

	var res []BadStruct
	err = db.Select(&res, "SELECT 'foo' as `invalid name!`")
	if err != nil && strings.Contains(err.Error(), "invalid character") {
		t.Fatalf("expected no tag errors in legacy mode, got: %v", err)
	}
}
