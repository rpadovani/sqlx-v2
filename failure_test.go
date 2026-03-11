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
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/rpadovani/sqlx-v2"
	_ "github.com/rpadovani/sqlx-v2/internal/mockdb"
)

func TestFailure_DeterministicCleanup(t *testing.T) {
	db, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	sdb := sqlx.NewDb(db, "mockdb")
	ctx := context.Background()

	testCases := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "fail_at_row_0",
			query:   "SELECT fail:at=0:small:10",
			wantErr: true,
		},
		{
			name:    "fail_at_row_5",
			query:   "SELECT fail:at=5:small:10",
			wantErr: true,
		},
		{
			name:    "fail_at_row_9",
			query:   "SELECT fail:at=9:small:10",
			wantErr: true,
		},
		{
			name:    "no_failure",
			query:   "SELECT small:10",
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			count := 0
			for s, err := range sqlx.SelectIter[SmallStruct](ctx, sdb, tc.query) {
				if err != nil {
					if !tc.wantErr {
						t.Errorf("unexpected error: %v", err)
					}
					return
				}
				_ = s
				count++
			}
			if tc.wantErr {
				t.Error("expected error but got none")
			}
		})
	}
}

func TestFailure_ProbabilisticStress(t *testing.T) {
	db, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	sdb := sqlx.NewDb(db, "mockdb")
	ctx := context.Background()

	// Stress test with multiple parallel iterators failing randomly
	const parallel = 10
	const rows = 1000
	errChan := make(chan error, parallel)

	for range parallel {
		go func() {
			var lastErr error
			for _, err := range sqlx.SelectIter[LargeStruct](ctx, sdb, fmt.Sprintf("SELECT fail:p=0.01:large:%d", rows)) {
				if err != nil {
					lastErr = err
					break
				}
			}
			errChan <- lastErr
		}()
	}

	for i := range parallel {
		err := <-errChan
		if err != nil {
			t.Logf("Iterator %d failed as expected: %v", i, err)
		}
	}
}

func TestFailure_ErrorTypes(t *testing.T) {
	db, err := sql.Open("mockdb", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()

	sdb := sqlx.NewDb(db, "mockdb")
	ctx := context.Background()

	testCases := []struct {
		failType string
		check    func(error) bool
	}{
		{
			failType: "eof",
			check:    func(err error) bool { return errors.Is(err, io.ErrUnexpectedEOF) },
		},
		{
			failType: "timeout",
			check:    func(err error) bool { return strings.Contains(err.Error(), "mockdb: network timeout") },
		},
		{
			failType: "reset",
			check:    func(err error) bool { return strings.Contains(err.Error(), "mockdb: connection reset by peer") },
		},
	}

	for _, tc := range testCases {
		t.Run(tc.failType, func(t *testing.T) {
			query := fmt.Sprintf("SELECT fail:at=0,type=%s:small:1", tc.failType)
			for _, err := range sqlx.SelectIter[SmallStruct](ctx, sdb, query) {
				if err == nil {
					t.Fatal("expected error")
				}
				if !tc.check(err) {
					t.Errorf("wrong error type: %v", err)
				}
			}
		})
	}
}
