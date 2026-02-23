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

// Package mockdb provides a mock database/sql driver for benchmarking purposes.
// It eliminates network/IO noise so benchmarks measure only the CPU and memory
// overhead of the mapping engine.
package mockdb

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"math/rand/v2"
	"strings"
)

func init() {
	sql.Register("mockdb", &MockDriver{})
}

// MockDriver implements driver.Driver.
type MockDriver struct{}

func (d *MockDriver) Open(name string) (driver.Conn, error) {
	return &MockConn{}, nil
}

// MockConn implements driver.Conn.
type MockConn struct{}

func (c *MockConn) Prepare(query string) (driver.Stmt, error) {
	return &MockStmt{query: query}, nil
}

func (c *MockConn) Close() error {
	return nil
}

func (c *MockConn) Begin() (driver.Tx, error) {
	return &MockTx{}, nil
}

// MockTx implements driver.Tx.
type MockTx struct{}

func (t *MockTx) Commit() error   { return nil }
func (t *MockTx) Rollback() error { return nil }

// MockStmt implements driver.Stmt.
type MockStmt struct {
	query string
}

func (s *MockStmt) Close() error {
	return nil
}

func (s *MockStmt) NumInput() int {
	return -1
}

func (s *MockStmt) Exec(args []driver.Value) (driver.Result, error) {
	return &MockResult{}, nil
}

func (s *MockStmt) Query(args []driver.Value) (driver.Rows, error) {
	// Parse the query to determine what kind of mock rows to return.
	// The query format is: "SELECT <type>:<numRows>"
	// where type is "small", "medium", or "large"
	query := strings.TrimSpace(s.query)

	if !strings.HasPrefix(query, "SELECT ") {
		return NewMockRows(SmallColumns(), SmallRow(), 1, -1, 0, "error"), nil
	}

	parts := strings.Split(strings.TrimPrefix(query, "SELECT "), ":")

	var structType string
	numRows := 100
	failAt := -1
	failProb := 0.0
	failType := "error"

	for _, part := range parts {
		if strings.Contains(part, "=") {
			for p := range strings.SplitSeq(part, ",") {
				kv := strings.SplitN(p, "=", 2)
				if len(kv) != 2 {
					continue
				}
				switch kv[0] {
				case "at":
					_, _ = fmt.Sscanf(kv[1], "%d", &failAt)
				case "p":
					_, _ = fmt.Sscanf(kv[1], "%f", &failProb)
				case "type":
					failType = kv[1]
				}
			}
			continue
		}

		if part == "fail" || part == "failure" {
			continue
		}

		val := 0
		if n, _ := fmt.Sscanf(part, "%d", &val); n == 1 {
			numRows = val
			continue
		}

		if structType == "" {
			structType = part
		}
	}

	if structType == "" {
		structType = "small"
	}

	cols, row := getRowsForType(structType)
	return NewMockRows(cols, row, numRows, failAt, failProb, failType), nil
}

func getRowsForType(t string) ([]string, []driver.Value) {
	switch t {
	case "small":
		return SmallColumns(), SmallRow()
	case "medium":
		return MediumColumns(), MediumRow()
	case "large":
		return LargeColumns(), LargeRow()
	case "embedded":
		return EmbeddedColumns(), EmbeddedRow()
	default:
		return SmallColumns(), SmallRow()
	}
}

// MockResult implements driver.Result.
type MockResult struct{}

func (r *MockResult) LastInsertId() (int64, error) { return 0, nil }
func (r *MockResult) RowsAffected() (int64, error) { return 0, nil }

// MockRows implements driver.Rows.
type MockRows struct {
	columns  []string
	row      []driver.Value
	numRows  int
	cursor   int
	failAt   int
	failProb float64
	failType string
}

// NewMockRows creates a new MockRows that returns the same row numRows times.
func NewMockRows(columns []string, row []driver.Value, numRows, failAt int, failProb float64, failType string) *MockRows {
	return &MockRows{
		columns:  columns,
		row:      row,
		numRows:  numRows,
		cursor:   0,
		failAt:   failAt,
		failProb: failProb,
		failType: failType,
	}
}

func (r *MockRows) Columns() []string {
	return r.columns
}

func (r *MockRows) Close() error {
	return nil
}

func (r *MockRows) Next(dest []driver.Value) error {
	if r.failAt != -1 && r.cursor == r.failAt {
		return r.generateError()
	}
	if r.failProb > 0 && rand.Float64() < r.failProb {
		return r.generateError()
	}

	if r.cursor >= r.numRows {
		return io.EOF
	}
	r.cursor++
	copy(dest, r.row)
	return nil
}

func (r *MockRows) generateError() error {
	switch r.failType {
	case "eof":
		return io.ErrUnexpectedEOF
	case "timeout":
		return fmt.Errorf("mockdb: network timeout")
	case "reset":
		return fmt.Errorf("mockdb: connection reset by peer")
	default:
		return fmt.Errorf("mockdb: failure at row %d", r.cursor)
	}
}

// --- Struct definitions for small, medium, and large structs ---

// SmallColumns returns columns for a 3-field struct.
func SmallColumns() []string {
	return []string{"id", "name", "email"}
}

// SmallRow returns a sample row for the small struct.
func SmallRow() []driver.Value {
	return []driver.Value{int64(1), "John Doe", "john@example.com"}
}

// MediumColumns returns columns for a 15-field struct.
func MediumColumns() []string {
	return []string{
		"id", "first_name", "last_name", "email", "phone",
		"address", "city", "state", "zip", "country",
		"age", "score", "active", "created_at", "updated_at",
	}
}

// MediumRow returns a sample row for the medium struct.
func MediumRow() []driver.Value {
	return []driver.Value{
		int64(1), "John", "Doe", "john@example.com", "+1234567890",
		"123 Main St", "Anytown", "CA", "90210", "US",
		int64(30), float64(95.5), true, "2024-01-01T00:00:00Z", "2024-06-01T00:00:00Z",
	}
}

// LargeColumns returns columns for a 50-field struct.
func LargeColumns() []string {
	cols := make([]string, 50)
	for i := range 50 {
		cols[i] = fmt.Sprintf("field_%02d", i)
	}
	return cols
}

// EmbeddedColumns returns columns for a struct with embedded fields.
func EmbeddedColumns() []string {
	return []string{"name", "street", "city"}
}

// EmbeddedRow returns a sample row for the embedded struct.
func EmbeddedRow() []driver.Value {
	return []driver.Value{"Alice", "123 Main St", "Springfield"}
}

// LargeRow returns a sample row for the large struct.
func LargeRow() []driver.Value {
	row := make([]driver.Value, 50)
	for i := range 50 {
		switch i % 3 {
		case 0:
			row[i] = int64(i)
		case 1:
			row[i] = fmt.Sprintf("value_%d", i)
		case 2:
			row[i] = float64(i) * 1.1
		}
	}
	return row
}
