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

package bind

import (
	"math/rand"
	"testing"
)

func oldBindType(driverName string) int {
	switch driverName {
	case "postgres", "pgx", "pq-timeouts", "cloudsqlpostgres", "ql":
		return DOLLAR
	case "mysql":
		return QUESTION
	case "sqlite3":
		return QUESTION
	case "oci8", "ora", "goracle", "godror":
		return NAMED
	case "sqlserver":
		return AT
	}
	return UNKNOWN
}

/*
sync.Map implementation:

goos: linux
goarch: amd64
pkg: github.com/rpadovani/sqlx-v2
BenchmarkBindSpeed/old-4         	100000000	        11.0 ns/op
BenchmarkBindSpeed/new-4         	24575726	        50.8 ns/op


async.Value map implementation:

goos: linux
goarch: amd64
pkg: github.com/rpadovani/sqlx-v2
BenchmarkBindSpeed/old-4         	100000000	        11.0 ns/op
BenchmarkBindSpeed/new-4         	42535839	        27.5 ns/op
*/

func BenchmarkBindSpeed(b *testing.B) {
	testDrivers := []string{
		"postgres", "pgx", "mysql", "sqlite3", "ora", "sqlserver",
	}

	b.Run("old", func(b *testing.B) {
		b.StopTimer()
		var seq []int
		for i := 0; i < b.N; i++ {
			seq = append(seq, rand.Intn(len(testDrivers)))
		}
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			s := oldBindType(testDrivers[seq[i]])
			if s == UNKNOWN {
				b.Error("unknown driver")
			}
		}
	})

	b.Run("new", func(b *testing.B) {
		b.StopTimer()
		var seq []int
		for i := 0; i < b.N; i++ {
			seq = append(seq, rand.Intn(len(testDrivers)))
		}
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			s := BindType(testDrivers[seq[i]])
			if s == UNKNOWN {
				b.Error("unknown driver")
			}
		}
	})
}

func TestCompileQuery(t *testing.T) {
	table := []struct {
		Q, R, D, T, N string
		V             []string
	}{
		// basic test for named parameters, invalid char ',' terminating
		{
			Q: `INSERT INTO foo (a,b,c,d) VALUES (:name, :age, :first, :last)`,
			R: `INSERT INTO foo (a,b,c,d) VALUES (?, ?, ?, ?)`,
			D: `INSERT INTO foo (a,b,c,d) VALUES ($1, $2, $3, $4)`,
			T: `INSERT INTO foo (a,b,c,d) VALUES (@p1, @p2, @p3, @p4)`,
			N: `INSERT INTO foo (a,b,c,d) VALUES (:name, :age, :first, :last)`,
			V: []string{"name", "age", "first", "last"},
		},
		// This query tests a named parameter ending the string as well as numbers
		{
			Q: `SELECT * FROM a WHERE first_name=:name1 AND last_name=:name2`,
			R: `SELECT * FROM a WHERE first_name=? AND last_name=?`,
			D: `SELECT * FROM a WHERE first_name=$1 AND last_name=$2`,
			T: `SELECT * FROM a WHERE first_name=@p1 AND last_name=@p2`,
			N: `SELECT * FROM a WHERE first_name=:name1 AND last_name=:name2`,
			V: []string{"name1", "name2"},
		},
		{
			Q: `SELECT "::foo" FROM a WHERE first_name=:name1 AND last_name=:name2`,
			R: `SELECT ":foo" FROM a WHERE first_name=? AND last_name=?`,
			D: `SELECT ":foo" FROM a WHERE first_name=$1 AND last_name=$2`,
			T: `SELECT ":foo" FROM a WHERE first_name=@p1 AND last_name=@p2`,
			N: `SELECT ":foo" FROM a WHERE first_name=:name1 AND last_name=:name2`,
			V: []string{"name1", "name2"},
		},
		{
			Q: `SELECT 'a::b::c' || first_name, '::::ABC::_::' FROM person WHERE first_name=:first_name AND last_name=:last_name`,
			R: `SELECT 'a:b:c' || first_name, '::ABC:_:' FROM person WHERE first_name=? AND last_name=?`,
			D: `SELECT 'a:b:c' || first_name, '::ABC:_:' FROM person WHERE first_name=$1 AND last_name=$2`,
			T: `SELECT 'a:b:c' || first_name, '::ABC:_:' FROM person WHERE first_name=@p1 AND last_name=@p2`,
			N: `SELECT 'a:b:c' || first_name, '::ABC:_:' FROM person WHERE first_name=:first_name AND last_name=:last_name`,
			V: []string{"first_name", "last_name"},
		},
		{
			Q: `SELECT @name := "name", :age, :first, :last`,
			R: `SELECT @name := "name", ?, ?, ?`,
			D: `SELECT @name := "name", $1, $2, $3`,
			N: `SELECT @name := "name", :age, :first, :last`,
			T: `SELECT @name := "name", @p1, @p2, @p3`,
			V: []string{"age", "first", "last"},
		},
		/* This unicode awareness test sadly fails, because of our byte-wise worldview.
		 * We could certainly iterate by Rune instead, though it's a great deal slower,
		 * it's probably the RightWay(tm)
		{
			Q: `INSERT INTO foo (a,b,c,d) VALUES (:あ, :b, :キコ, :名前)`,
			R: `INSERT INTO foo (a,b,c,d) VALUES (?, ?, ?, ?)`,
			D: `INSERT INTO foo (a,b,c,d) VALUES ($1, $2, $3, $4)`,
			N: []string{"name", "age", "first", "last"},
		},
		*/
	}

	for _, test := range table {
		qr, names, err := compileNamedQuery([]byte(test.Q), QUESTION)
		if err != nil {
			t.Error(err)
		}
		if qr != test.R {
			t.Errorf("expected %s, got %s", test.R, qr)
		}
		if len(names) != len(test.V) {
			t.Errorf("expected %#v, got %#v", test.V, names)
		} else {
			for i, name := range names {
				if name != test.V[i] {
					t.Errorf("expected %dth name to be %s, got %s", i+1, test.V[i], name)
				}
			}
		}
		qd, _, _ := compileNamedQuery([]byte(test.Q), DOLLAR)
		if qd != test.D {
			t.Errorf("\nexpected: `%s`\ngot:      `%s`", test.D, qd)
		}

		qt, _, _ := compileNamedQuery([]byte(test.Q), AT)
		if qt != test.T {
			t.Errorf("\nexpected: `%s`\ngot:      `%s`", test.T, qt)
		}

		qq, _, _ := compileNamedQuery([]byte(test.Q), NAMED)
		if qq != test.N {
			t.Errorf("\nexpected: `%s`\ngot:      `%s`\n(len: %d vs %d)", test.N, qq, len(test.N), len(qq))
		}
	}
}

type Test struct {
	t *testing.T
}

func (t Test) Error(err error, msg ...any) {
	t.t.Helper()
	if err != nil {
		if len(msg) == 0 {
			t.t.Error(err)
		} else {
			t.t.Error(msg...)
		}
	}
}

func (t Test) Errorf(err error, format string, args ...any) {
	t.t.Helper()
	if err != nil {
		t.t.Errorf(format, args...)
	}
}

func TestEscapedColons(t *testing.T) {
	t.Skip("not sure it is possible to support this in general case without an SQL parser")
	qs := `SELECT * FROM testtable WHERE timeposted BETWEEN (now() AT TIME ZONE 'utc') AND
	(now() AT TIME ZONE 'utc') - interval '01:30:00') AND name = '\'this is a test\'' and id = :id`
	_, _, err := compileNamedQuery([]byte(qs), DOLLAR)
	if err != nil {
		t.Error("Didn't handle colons correctly when inside a string")
	}
}

type bindTestStruct struct {
	Name string `db:"name"`
	Age  int    `db:"age"`
}

func TestBindNamed_Struct(t *testing.T) {
	q := "INSERT INTO users (name, age) VALUES (:name, :age)"
	arg := bindTestStruct{Name: "Alice", Age: 30}

	bound, args, err := BindNamed(QUESTION, q, arg)
	if err != nil {
		t.Fatalf("BindNamed struct failed: %v", err)
	}

	if bound != "INSERT INTO users (name, age) VALUES (?, ?)" {
		t.Errorf("Unexpected bound query: %s", bound)
	}
	if len(args) != 2 || args[0] != "Alice" || args[1] != 30 {
		t.Errorf("Unexpected args: %v", args)
	}
}

func TestBindNamed_Map(t *testing.T) {
	q := "SELECT * FROM users WHERE age > :min_age AND status = :status"
	arg := map[string]any{
		"min_age": 18,
		"status":  "active",
	}

	bound, args, err := BindNamed(DOLLAR, q, arg)
	if err != nil {
		t.Fatalf("BindNamed map failed: %v", err)
	}

	if bound != "SELECT * FROM users WHERE age > $1 AND status = $2" {
		t.Errorf("Unexpected bound query: %s", bound)
	}
	if len(args) != 2 || args[0] != 18 || args[1] != "active" {
		t.Errorf("Unexpected args: %v", args)
	}
}

func TestRebind(t *testing.T) {
	q := "SELECT * FROM users WHERE age > ? AND status = ?"

	d := Rebind(DOLLAR, q)
	if d != "SELECT * FROM users WHERE age > $1 AND status = $2" {
		t.Errorf("Rebind DOLLAR failed: %s", d)
	}

	a := Rebind(AT, q)
	if a != "SELECT * FROM users WHERE age > @p1 AND status = @p2" {
		t.Errorf("Rebind AT failed: %s", a)
	}

	n := Rebind(NAMED, q)
	if n != "SELECT * FROM users WHERE age > :arg1 AND status = :arg2" {
		t.Errorf("Rebind NAMED failed: %s", n)
	}

	u := Rebind(UNKNOWN, q)
	if u != q {
		t.Errorf("Rebind UNKNOWN failed: %s", u)
	}
}

func TestBindType(t *testing.T) {
	if BindType("postgres") != DOLLAR {
		t.Error("postgres not DOLLAR")
	}
	if BindType("sqlite3") != QUESTION {
		t.Error("sqlite3 not QUESTION")
	}
	if BindType("mysql") != QUESTION {
		t.Error("mysql not QUESTION")
	}
	if BindType("sqlserver") != AT {
		t.Error("sqlserver not AT")
	}
	if BindType("godror") != NAMED {
		t.Error("godror not NAMED")
	}
	if BindType("fake-driver") != UNKNOWN {
		t.Error("fake-driver not UNKNOWN")
	}
}

func TestInFunctionality(t *testing.T) {
	// Simple slice expansion
	q := "SELECT * FROM users WHERE id IN (?) AND age > ?"
	args := []any{[]int{1, 2, 3}, 18}

	rq, rargs, err := In(q, args...)
	if err != nil {
		t.Fatalf("In func failed: %v", err)
	}

	expectedQ := "SELECT * FROM users WHERE id IN (?, ?, ?) AND age > ?"
	if rq != expectedQ {
		t.Errorf("Unexpected In query: expected %s, got %s", expectedQ, rq)
	}
	if len(rargs) != 4 || rargs[0] != 1 || rargs[1] != 2 || rargs[2] != 3 || rargs[3] != 18 {
		t.Errorf("Unexpected In args: %v", rargs)
	}

	// Double slice expansion
	q2 := "SELECT * FROM users WHERE id IN (?) AND status IN (?)"
	args2 := []any{[]int{1, 2}, []string{"active", "pending"}}

	rq2, rargs2, err2 := In(q2, args2...)
	if err2 != nil {
		t.Fatalf("In func failed: %v", err2)
	}

	expectedQ2 := "SELECT * FROM users WHERE id IN (?, ?) AND status IN (?, ?)"
	if rq2 != expectedQ2 {
		t.Errorf("Unexpected In query: expected %s, got %s", expectedQ2, rq2)
	}
	if len(rargs2) != 4 || rargs2[0] != 1 || rargs2[1] != 2 || rargs2[2] != "active" || rargs2[3] != "pending" {
		t.Errorf("Unexpected In args: %v", rargs2)
	}
}

func TestCompileNamedQueryAndNamed(t *testing.T) {
	q := "SELECT * FROM users WHERE status = :status"
	arg := map[string]any{"status": "active"}

	bound, args, err := Named(q, arg)
	if err != nil {
		t.Fatalf("Named failed: %v", err)
	}
	if bound != "SELECT * FROM users WHERE status = ?" {
		t.Errorf("Unexpected Named query: %s", bound)
	}
	if len(args) != 1 || args[0] != "active" {
		t.Errorf("Unexpected args: %v", args)
	}

	bound2, names, err := CompileNamedQuery([]byte(q), DOLLAR)
	if err != nil {
		t.Fatalf("CompileNamedQuery failed: %v", err)
	}
	if bound2 != "SELECT * FROM users WHERE status = $1" {
		t.Errorf("Unexpected CompileNamed query: %s", bound2)
	}
	if len(names) != 1 || names[0] != "status" {
		t.Errorf("Unexpected names: %v", names)
	}
}

func TestBindArray(t *testing.T) {
	q := "INSERT INTO users (name, age) VALUES (:name, :age)"
	args := []bindTestStruct{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 40},
		{Name: "Charlie", Age: 50},
	}

	bound, arglist, err := BindNamed(QUESTION, q, args)
	if err != nil {
		t.Fatalf("BindNamed array failed: %v", err)
	}

	expectedQ := "INSERT INTO users (name, age) VALUES (?, ?), (?, ?), (?, ?)"
	if bound != expectedQ {
		t.Errorf("Unexpected bound query: expected %s, got %s", expectedQ, bound)
	}
	if len(arglist) != 6 {
		t.Errorf("Unexpected args length: expected 6, got %d", len(arglist))
	}

	boundDollar, arglistDollar, err := BindNamed(DOLLAR, q, args)
	if err != nil {
		t.Fatalf("BindNamed DOLLAR array failed: %v", err)
	}

	expectedQDollar := "INSERT INTO users (name, age) VALUES ($1, $2), ($3, $4), ($5, $6)"
	if boundDollar != expectedQDollar {
		t.Errorf("Unexpected bound DOLLAR query: expected %s, got %s", expectedQDollar, boundDollar)
	}
	if len(arglistDollar) != 6 {
		t.Errorf("Unexpected args DOLLAR length: expected 6, got %d", len(arglistDollar))
	}
}

// TestCompileNamedQuery_CachePoisoning verifies that mutating the names slice
// returned by CompileNamedQuery does not corrupt the internal compile cache.
func TestCompileNamedQuery_CachePoisoning(t *testing.T) {
	q := []byte("INSERT INTO poison_test (a,b) VALUES (:x, :y)")

	// First call: populate cache
	_, names1, err := CompileNamedQuery(q, QUESTION)
	if err != nil {
		t.Fatal(err)
	}
	if len(names1) != 2 || names1[0] != "x" || names1[1] != "y" {
		t.Fatalf("unexpected names: %v", names1)
	}

	// Poison: mutate the returned slice
	names1[0] = "POISONED"
	names1[1] = "EVIL"

	// Second call: must still return clean data from cache
	_, names2, err := CompileNamedQuery(q, QUESTION)
	if err != nil {
		t.Fatal(err)
	}
	if names2[0] != "x" || names2[1] != "y" {
		t.Errorf("cache poisoned! got %v, want [x y]", names2)
	}
}

// TestBindStructDirect verifies the optimized struct binding path.
func TestBindStructDirect(t *testing.T) {
	q := "INSERT INTO users (name, age) VALUES (:name, :age)"
	arg := &bindTestStruct{Name: "Alice", Age: 30}

	bound, args, err := BindNamed(QUESTION, q, arg)
	if err != nil {
		t.Fatalf("BindNamed struct pointer failed: %v", err)
	}

	if bound != "INSERT INTO users (name, age) VALUES (?, ?)" {
		t.Errorf("Unexpected bound query: %s", bound)
	}
	if len(args) != 2 || args[0] != "Alice" || args[1] != 30 {
		t.Errorf("Unexpected args: %v", args)
	}
}
