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
	"fmt"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/rpadovani/sqlx-v2"
	"github.com/rpadovani/sqlx-v2/tests/internal/testutil"
)

func TestFailure_SocketReset_SelectIter(t *testing.T) {
	dsn, proxy := testutil.GlobalToxiproxyPostgres(t)
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		t.Fatalf("failed to connect to proxied postgres: %v", err)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()

	// Setup table and data
	db.MustExec(`DROP TABLE IF EXISTS toxiproxy_test`)
	db.MustExec(`CREATE TABLE toxiproxy_test (id INT PRIMARY KEY, val TEXT)`)

	// Insert 50,000 rows to ensure we have a stream to break
	for i := range 50 {
		// Bulk insert for speed
		var values []string
		for j := range 1000 {
			id := i*1000 + j
			values = append(values, fmt.Sprintf("(%d, '%s')", id, fmt.Sprintf("value-%0100d", id)))
		}
		query := fmt.Sprintf("INSERT INTO toxiproxy_test (id, val) VALUES %s", strings.Join(values, ","))
		db.MustExec(query)
	}

	type Row struct {
		ID  int    `db:"id"`
		Val string `db:"val"`
	}

	// Reset proxy state just in case
	_ = proxy.RemoveToxic("reset_downstream")

	count := 0
	var scanErr error

	// Scan through rows and kill the connection halfway
	for _, err := range sqlx.SelectIter[Row](ctx, db, "SELECT * FROM toxiproxy_test ORDER BY id") {
		if err != nil {
			scanErr = err
			break
		}
		count++

		if count == 1000 {
			// PHYSICALLY kill the TCP connection via Toxiproxy
			_, err := proxy.AddToxic("reset_downstream", "reset_peer", "downstream", 1.0, nil)
			if err != nil {
				t.Fatalf("failed to add toxic: %v", err)
			}
		}
	}

	if scanErr == nil {
		t.Errorf("expected scan error due to physical socket reset at row 1000, but scan finished successfully after %d rows", count)
	} else {
		t.Logf("Verified: Scan failed correctly after %d rows with error: %v", count, scanErr)
	}

	// Cleanup for next tests
	_ = proxy.RemoveToxic("reset_downstream")
}
