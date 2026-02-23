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
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/rpadovani/sqlx-v2/internal/reflectx"
)

// ErrBindMismatch is returned when there is a mismatch between binding variables and parameters.
var ErrBindMismatch = errors.New("sqlx: bind count mismatch")

// ErrNamedPropertyNotFound is returned when a named parameter cannot be found in the provided struct or map.
var ErrNamedPropertyNotFound = errors.New("sqlx: named property not found")

// ErrUnsupportedType is returned when an unsupported type is provided for binding.
var ErrUnsupportedType = errors.New("sqlx: unsupported type")

// ErrSyntax is returned when there is a syntax error in the query parsing.
var ErrSyntax = errors.New("sqlx: query syntax error")

// ErrColumnNotFound is returned when a column from the database cannot be mapped to a destination.
var ErrColumnNotFound = errors.New("sqlx: column not found")

// Bindvar types supported by Rebind, BindMap and BindStruct.
const (
	UNKNOWN = iota
	QUESTION
	DOLLAR
	NAMED
	AT
)

var defaultBinds = map[int][]string{
	DOLLAR:   {"postgres", "pgx", "pq-timeouts", "cloudsqlpostgres", "ql", "nrpostgres", "cockroach"},
	QUESTION: {"mysql", "sqlite3", "nrmysql", "nrsqlite3"},
	NAMED:    {"oci8", "ora", "goracle", "godror"},
	AT:       {"sqlserver", "azuresql"},
}

var binds sync.Map

func init() {
	for bind, drivers := range defaultBinds {
		for _, driver := range drivers {
			BindDriver(driver, bind)
		}
	}

}

// BindType returns the bindtype for a given database given a drivername.
func BindType(driverName string) int {
	itype, ok := binds.Load(driverName)
	if !ok {
		return UNKNOWN
	}
	return itype.(int)
}

// BindDriver sets the BindType for driverName to bindType.
func BindDriver(driverName string, bindType int) {
	binds.Store(driverName, bindType)
}

// FIXME: this should be able to be tolerant of escaped ?'s in queries without
// losing much speed, and should be to avoid confusion.

// Rebind a query from the default bindtype (QUESTION) to the target bindtype.
func Rebind(bindType int, query string) string {
	switch bindType {
	case QUESTION, UNKNOWN:
		return query
	}

	// Add space enough for 10 params before we have to allocate
	rqb := make([]byte, 0, len(query)+10)

	var i, j int

	for i = strings.Index(query, "?"); i != -1; i = strings.Index(query, "?") {
		rqb = append(rqb, query[:i]...)

		switch bindType {
		case DOLLAR:
			rqb = append(rqb, '$')
		case NAMED:
			rqb = append(rqb, ':', 'a', 'r', 'g')
		case AT:
			rqb = append(rqb, '@', 'p')
		}

		j++
		rqb = strconv.AppendInt(rqb, int64(j), 10)

		query = query[i+1:]
	}

	return string(append(rqb, query...))
}

func asSliceForIn(i any) (v reflect.Value, ok bool) {
	if i == nil {
		return reflect.Value{}, false
	}

	v = reflect.ValueOf(i)
	t := reflectx.Deref(v.Type())

	// Only expand slices
	if t.Kind() != reflect.Slice {
		return reflect.Value{}, false
	}

	// []byte is a driver.Value type so it should not be expanded
	if t == reflect.TypeFor[[]byte]() {
		return reflect.Value{}, false

	}

	return v, true
}

// In expands slice values in args, returning the modified query string
// and a new arg list that can be executed by a database. The `query` should
// use the `?` bindVar.  The return value uses the `?` bindVar.
func In(query string, args ...any) (string, []any, error) {
	// argMeta stores reflect.Value and length for slices and
	// the value itself for non-slice arguments
	type argMeta struct {
		v      reflect.Value
		i      any
		length int
	}

	var flatArgsCount int
	var anySlices bool

	var stackMeta [32]argMeta

	var meta []argMeta
	if len(args) <= len(stackMeta) {
		meta = stackMeta[:len(args)]
	} else {
		meta = make([]argMeta, len(args))
	}

	for i, arg := range args {
		if a, ok := arg.(driver.Valuer); ok {
			var err error
			arg, err = a.Value()
			if err != nil {
				return "", nil, fmt.Errorf("sqlx: error getting value from driver.Valuer: %w", err)
			}
		}

		if v, ok := asSliceForIn(arg); ok {
			meta[i].length = v.Len()
			meta[i].v = v

			anySlices = true
			flatArgsCount += meta[i].length

			if meta[i].length == 0 {
				return "", nil, fmt.Errorf("sqlx: empty slice passed to 'In' query: %w", sql.ErrNoRows) // Using sql.ErrNoRows as a semi-appropriate wrapper or just fmt.Errorf
			}
		} else {
			meta[i].i = arg
			flatArgsCount++
		}
	}

	// don't do any parsing if there aren't any slices;  note that this means
	// some errors that we might have caught below will not be returned.
	if !anySlices {
		return query, args, nil
	}

	newArgs := make([]any, 0, flatArgsCount)

	var buf strings.Builder
	buf.Grow(len(query) + len(", ?")*flatArgsCount)

	var arg, offset int

	for i := strings.IndexByte(query[offset:], '?'); i != -1; i = strings.IndexByte(query[offset:], '?') {
		if arg >= len(meta) {
			// if an argument wasn't passed, lets return an error;  this is
			// not actually how database/sql Exec/Query works, but since we are
			// creating an argument list programmatically, we want to be able
			// to catch these programmer errors earlier.
			return "", nil, fmt.Errorf("sqlx: number of bindVars exceeds arguments: %w", ErrBindMismatch)
		}

		argMeta := meta[arg]
		arg++

		// not a slice, continue.
		// our questionmark will either be written before the next expansion
		// of a slice or after the loop when writing the rest of the query
		if argMeta.length == 0 {
			offset = offset + i + 1
			newArgs = append(newArgs, argMeta.i)
			continue
		}

		// write everything up to and including our ? character
		buf.WriteString(query[:offset+i+1])

		for si := 1; si < argMeta.length; si++ {
			buf.WriteString(", ?")
		}

		newArgs = appendReflectSlice(newArgs, argMeta.v, argMeta.length)

		// slice the query and reset the offset. this avoids some bookkeeping for
		// the write after the loop
		query = query[offset+i+1:]
		offset = 0
	}

	buf.WriteString(query)

	if arg < len(meta) {
		return "", nil, fmt.Errorf("sqlx: number of bindVars less than number arguments: %w", ErrBindMismatch)
	}

	return buf.String(), newArgs, nil
}

func appendReflectSlice(args []any, v reflect.Value, vlen int) []any {
	switch val := v.Interface().(type) {
	case []any:
		args = append(args, val...)
	case []int:
		for i := range val {
			args = append(args, val[i])
		}
	case []string:
		for i := range val {
			args = append(args, val[i])
		}
	default:
		for si := range vlen {
			args = append(args, v.Index(si).Interface())
		}
	}

	return args
}

// convertMapStringInterface attempts to convert v to map[string]any.
func convertMapStringInterface(v any) (map[string]any, bool) {
	mtype := reflect.TypeFor[map[string]any]()
	t := reflect.TypeOf(v)
	if !t.ConvertibleTo(mtype) {
		return nil, false
	}
	return reflect.ValueOf(v).Convert(mtype).Interface().(map[string]any), true
}

func bindAnyArgs(names []string, arg any, m *reflectx.Mapper) ([]any, error) {
	if maparg, ok := convertMapStringInterface(arg); ok {
		return bindMapArgs(names, maparg)
	}
	return bindArgs(names, arg, m)
}

func bindArgs(names []string, arg any, m *reflectx.Mapper) ([]any, error) {
	arglist := make([]any, 0, len(names))
	v := reflect.ValueOf(arg)
	for v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	err := m.TraversalsByNameFunc(v.Type(), names, func(i int, t []int) error {
		if len(t) == 0 {
			return fmt.Errorf("sqlx: could not find name %q in %T: %w", names[i], arg, ErrNamedPropertyNotFound)
		}
		val := reflectx.FieldByIndexesReadOnly(v, t)
		arglist = append(arglist, val.Interface())
		return nil
	})
	return arglist, err
}

func bindMapArgs(names []string, arg map[string]any) ([]any, error) {
	arglist := make([]any, 0, len(names))
	for _, name := range names {
		val, ok := arg[name]
		if !ok {
			return arglist, fmt.Errorf("sqlx: could not find name %q in map: %w", name, ErrNamedPropertyNotFound)
		}
		arglist = append(arglist, val)
	}
	return arglist, nil
}

// bindStructDirect binds struct field values to the named parameter positions
// using direct StructMap lookups. It writes into the provided arglist slice
// which must have len >= len(names). This eliminates map allocations and
// closure overhead compared to the TraversalsByNameFunc path.
func bindStructDirect(names []string, arg any, m *reflectx.Mapper, arglist []any) error {
	v := reflect.ValueOf(arg)
	for v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	tm := m.TypeMap(v.Type())
	for i, name := range names {
		fi, ok := tm.Names[name]
		if !ok {
			return fmt.Errorf("sqlx: could not find name %q in %T: %w", name, arg, ErrNamedPropertyNotFound)
		}
		val := reflectx.FieldByIndexesReadOnly(v, fi.Index)
		arglist[i] = val.Interface()
	}
	return nil
}

func bindStruct(bindType int, query string, arg any, m *reflectx.Mapper) (string, []any, error) {
	bound, names, err := compileNamedQuery([]byte(query), bindType)
	if err != nil {
		return "", []any{}, fmt.Errorf("sqlx: error compiling named query: %w", err)
	}
	arglist := make([]any, len(names))
	if err := bindStructDirect(names, arg, m, arglist); err != nil {
		return "", []any{}, fmt.Errorf("sqlx: error binding args: %w", err)
	}
	return bound, arglist, nil
}

type compileKey struct {
	query    string
	bindType int
}

func bindArray(bindType int, query string, arg any, m *reflectx.Mapper) (string, []any, error) {
	v := reflect.ValueOf(arg)
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		return "", nil, fmt.Errorf("sqlx: bindArray expects an array or slice, got %T: %w", arg, ErrBindMismatch)
	}

	if v.Len() == 0 {
		return "", nil, fmt.Errorf("sqlx: bindArray cannot bind an empty slice: %w", sql.ErrNoRows)
	}

	// 1. Compile the base named query to get parameter names.
	// Since we are expanding an array into positional markers, we first compile to QUESTION (?) format.
	unboundQuery, names, err := compileNamedQuery([]byte(query), QUESTION)
	if err != nil {
		return "", []any{}, err
	}

	if len(names) == 0 {
		return "", nil, fmt.Errorf("sqlx: no named parameters found in slice binding: %w", ErrBindMismatch)
	}

	// 2. Extract values for all rows into arglist.
	arglist := make([]any, 0, len(names)*v.Len())
	for i := 0; i < v.Len(); i++ {
		val := v.Index(i)
		for val.Kind() == reflect.Pointer {
			val = val.Elem()
		}

		err := m.TraversalsByNameFunc(val.Type(), names, func(idx int, t []int) error {
			if len(t) == 0 {
				return fmt.Errorf("sqlx: could not find name %q in slice element %T: %w", names[idx], val.Interface(), ErrNamedPropertyNotFound)
			}
			fval := reflectx.FieldByIndexesReadOnly(val, t)
			arglist = append(arglist, fval.Interface())
			return nil
		})
		if err != nil {
			return "", nil, fmt.Errorf("sqlx: error in traversal during bindArray: %w", err)
		}
	}

	// 3. Reconstruct query for batch inserts.
	// Find the FIRST `?` and LAST `?` in the unbound query.
	var newQuery string
	//nolint:modernize // strings.Cut degrades absolute index computation readability
	firstQ := strings.IndexByte(unboundQuery, '?')
	lastQ := strings.LastIndexByte(unboundQuery, '?')

	if firstQ != -1 && lastQ != -1 {
		// Find the closest `(` before the first `?` and the closest `)` after the last `?`.
		openP := strings.LastIndex(unboundQuery[:firstQ], "(")
		closeP := strings.Index(unboundQuery[lastQ:], ")")

		if openP != -1 && closeP != -1 {
			closeP += lastQ // absolute index

			// Extract single (?, ?, ...) value group
			singleGroup := unboundQuery[openP : closeP+1]

			// Duplicate the single group for each item in the slice.
			groups := make([]string, v.Len())
			for i := 0; i < v.Len(); i++ {
				groups[i] = singleGroup
			}
			replacement := strings.Join(groups, ", ")

			newQuery = unboundQuery[:openP] + replacement + unboundQuery[closeP+1:]
		} else {
			return "", nil, fmt.Errorf("sqlx: slice binding is only valid for queries where parameters are enclosed in parentheses, like INSERT: %w", ErrBindMismatch)
		}
	} else {
		newQuery = unboundQuery
	}

	// 4. Finally, rebind the duplicated `?` values back to target bindType (e.g. DOLLAR adds $1, $2, ... )
	finalQuery := Rebind(bindType, newQuery)

	return finalQuery, arglist, nil
}

type compileResult struct {
	query string
	names []string
}

var compileCache sync.Map // map[compileKey]compileResult

// compileNamedQuery compiles a NamedQuery into an unbound query (using the '?' bindvar) and a list of names.
// It caches the results based on the query string and bindType.
func compileNamedQuery(qs []byte, bindType int) (query string, names []string, err error) {
	key := compileKey{query: string(qs), bindType: bindType}
	if v, ok := compileCache.Load(key); ok {
		res := v.(compileResult)
		// Internal callers are read-only; the public CompileNamedQuery wrapper
		// handles the defensive copy to prevent cache poisoning.
		return res.query, res.names, nil
	}

	names = nil
	rebound := make([]byte, 0, len(qs))
	inName := false
	last := len(qs) - 1
	currentVar := 1
	var name []byte

	for i, b := range qs {
		if b == ':' {
			if inName && i > 0 && qs[i-1] == ':' {
				rebound = append(rebound, ':')
				inName = false
				continue
			} else if inName {
				err = fmt.Errorf("sqlx: unexpected `:` while reading named param at %d: %w", i, ErrSyntax)
				return query, names, err
			}
			inName = true
			name = []byte{}
		} else if inName && i > 0 && b == '=' && len(name) == 0 {
			rebound = append(rebound, ':', '=')
			inName = false
			continue
		} else if inName && ((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_' || b == '.') && i != last {
			name = append(name, b)
		} else if inName {
			inName = false
			if i == last && ((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')) {
				name = append(name, b)
			}
			names = append(names, string(name))
			switch bindType {
			case NAMED:
				rebound = append(rebound, ':')
				rebound = append(rebound, name...)
			case QUESTION, UNKNOWN:
				rebound = append(rebound, '?')
			case DOLLAR:
				rebound = append(rebound, '$')
				for _, b := range strconv.Itoa(currentVar) {
					rebound = append(rebound, byte(b))
				}
				currentVar++
			case AT:
				rebound = append(rebound, '@', 'p')
				for _, b := range strconv.Itoa(currentVar) {
					rebound = append(rebound, byte(b))
				}
				currentVar++
			}
			if i != last {
				rebound = append(rebound, b)
			} else if (b < 'a' || b > 'z') && (b < 'A' || b > 'Z') && (b < '0' || b > '9') {
				rebound = append(rebound, b)
			}
		} else {
			rebound = append(rebound, b)
		}
	}
	resQuery := string(rebound)
	compileCache.Store(key, compileResult{query: resQuery, names: names})
	return resQuery, names, nil
}

func bindNamedMapper(bindType int, query string, arg any, m *reflectx.Mapper) (string, []any, error) {
	t := reflect.TypeOf(arg)
	k := t.Kind()
	// Dereference pointers to get the underlying kind.
	for k == reflect.Pointer {
		t = t.Elem()
		k = t.Kind()
	}
	switch {
	case k == reflect.Struct:
		// Fast path: struct binding (most common case for NamedExec)
		return bindStruct(bindType, query, arg, m)
	case k == reflect.Map && t.Key().Kind() == reflect.String:
		m, ok := convertMapStringInterface(arg)
		if !ok {
			return "", nil, fmt.Errorf("sqlx: unsupported map type %T: %w", arg, ErrUnsupportedType)
		}
		// bindMap inline
		bound, names, err := compileNamedQuery([]byte(query), bindType)
		if err != nil {
			return "", []any{}, fmt.Errorf("sqlx: error in compileNamedQuery (map): %w", err)
		}
		arglist, err := bindMapArgs(names, m)
		if err != nil {
			return bound, arglist, fmt.Errorf("sqlx: error in bindMapArgs: %w", err)
		}
		return bound, arglist, nil
	case k == reflect.Array || k == reflect.Slice:
		return bindArray(bindType, query, arg, m)
	default:
		// Fallback for unexpected types — try struct binding
		return bindStruct(bindType, query, arg, m)
	}
}

var (
	bindDefaultMapper     *reflectx.Mapper
	bindDefaultMapperOnce sync.Once
)

func getBindDefaultMapper() *reflectx.Mapper {
	bindDefaultMapperOnce.Do(func() {
		bindDefaultMapper = reflectx.NewMapperFunc("db", strings.ToLower)
	})
	return bindDefaultMapper
}

// BindNamed binds a struct or a map to a query with named parameters.
func BindNamed(bindType int, query string, arg any) (string, []any, error) {
	return bindNamedMapper(bindType, query, arg, getBindDefaultMapper())
}

// Named takes a query using named parameters and an argument and returns a new query.
func Named(query string, arg any) (string, []any, error) {
	return bindNamedMapper(QUESTION, query, arg, getBindDefaultMapper())
}

// CompileNamedQuery is public mapping logic used by the prepared named queries internally.
// It returns a defensive copy of the cached names slice to prevent external callers
// from poisoning the compile cache (e.g. via NamedStmt.Params mutation).
func CompileNamedQuery(qs []byte, bindType int) (query string, names []string, err error) {
	query, names, err = compileNamedQuery(qs, bindType)
	if err != nil || len(names) == 0 {
		return query, names, err
	}
	namesCopy := make([]string, len(names))
	copy(namesCopy, names)
	return query, namesCopy, err
}

// BindAnyArgs is public mapping logic used by named queries to bind struct args dynamically.
func BindAnyArgs(names []string, arg any, m *reflectx.Mapper) ([]any, error) {
	return bindAnyArgs(names, arg, m)
}
