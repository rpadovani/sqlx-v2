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

package reflectx

import (
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestMapper_BasicTags(t *testing.T) {
	type BasicStruct struct {
		ID    int    `db:"id"`
		Name  string `db:"name"`
		Email string `db:"email"`
	}

	m := NewMapperFunc("db", strings.ToLower)
	tm := m.TypeMap(reflect.TypeFor[BasicStruct]())

	if _, ok := tm.Names["id"]; !ok {
		t.Error("expected 'id' field")
	}
	if _, ok := tm.Names["name"]; !ok {
		t.Error("expected 'name' field")
	}
	if _, ok := tm.Names["email"]; !ok {
		t.Error("expected 'email' field")
	}
	if len(tm.Names) != 3 {
		t.Errorf("expected 3 fields, got %d", len(tm.Names))
	}
}

func TestMapper_DashTag(t *testing.T) {
	type DashStruct struct {
		ID     int    `db:"id"`
		Ignore string `db:"-"`
		Name   string `db:"name"`
	}

	m := NewMapperFunc("db", strings.ToLower)
	tm := m.TypeMap(reflect.TypeFor[DashStruct]())

	if _, ok := tm.Names["id"]; !ok {
		t.Error("expected 'id' field")
	}
	if _, ok := tm.Names["name"]; !ok {
		t.Error("expected 'name' field")
	}
	if _, ok := tm.Names["-"]; ok {
		t.Error("should not have '-' field")
	}
	if len(tm.Names) != 2 {
		t.Errorf("expected 2 fields, got %d", len(tm.Names))
	}
}

func TestMapper_NoTag(t *testing.T) {
	type NoTagStruct struct {
		ID   int
		Name string
	}

	m := NewMapperFunc("db", strings.ToLower)
	tm := m.TypeMap(reflect.TypeFor[NoTagStruct]())

	if _, ok := tm.Names["id"]; !ok {
		t.Error("expected 'id' field (lowercased)")
	}
	if _, ok := tm.Names["name"]; !ok {
		t.Error("expected 'name' field (lowercased)")
	}
}

func TestMapper_EmbeddedStruct(t *testing.T) {
	type Person struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
	}

	type User struct {
		Person
		Email string `db:"email"`
	}

	m := NewMapperFunc("db", strings.ToLower)
	tm := m.TypeMap(reflect.TypeFor[User]())

	if _, ok := tm.Names["first_name"]; !ok {
		t.Error("expected 'first_name' field from embedded struct")
	}
	if _, ok := tm.Names["last_name"]; !ok {
		t.Error("expected 'last_name' field from embedded struct")
	}
	if _, ok := tm.Names["email"]; !ok {
		t.Error("expected 'email' field")
	}
}

func TestMapper_DeepEmbeddedStruct(t *testing.T) {
	type Address struct {
		Street string `db:"street"`
		City   string `db:"city"`
	}

	type Person struct {
		Address
		Name string `db:"name"`
	}

	type Employee struct {
		Person
		EmployeeID int `db:"employee_id"`
	}

	m := NewMapperFunc("db", strings.ToLower)
	tm := m.TypeMap(reflect.TypeFor[Employee]())

	if _, ok := tm.Names["street"]; !ok {
		t.Error("expected 'street' field from deeply embedded struct")
	}
	if _, ok := tm.Names["city"]; !ok {
		t.Error("expected 'city' field from deeply embedded struct")
	}
	if _, ok := tm.Names["name"]; !ok {
		t.Error("expected 'name' field from embedded struct")
	}
	if _, ok := tm.Names["employee_id"]; !ok {
		t.Error("expected 'employee_id' field")
	}
}

func TestMapper_FieldByIndexes(t *testing.T) {
	type Inner struct {
		Value string `db:"value"`
	}

	type Outer struct {
		Inner
		ID int `db:"id"`
	}

	m := NewMapperFunc("db", strings.ToLower)
	tm := m.TypeMap(reflect.TypeFor[Outer]())

	o := Outer{
		Inner: Inner{Value: "hello"},
		ID:    42,
	}

	v := reflect.ValueOf(&o).Elem()

	fi := tm.Names["value"]
	result := FieldByIndexes(v, fi.Index)
	if result.String() != "hello" {
		t.Errorf("expected 'hello', got '%s'", result.String())
	}

	fi = tm.Names["id"]
	result = FieldByIndexes(v, fi.Index)
	if result.Int() != 42 {
		t.Errorf("expected 42, got %d", result.Int())
	}
}

func TestMapper_TagWithOptions(t *testing.T) {
	type OptionsStruct struct {
		ID   int    `db:"id,pk"`
		Name string `db:"name,omitempty"`
	}

	m := NewMapperFunc("db", strings.ToLower)
	tm := m.TypeMap(reflect.TypeFor[OptionsStruct]())

	if fi, ok := tm.Names["id"]; !ok {
		t.Error("expected 'id' field")
	} else if _, hasPK := fi.Options["pk"]; !hasPK {
		t.Error("expected 'pk' option on 'id' field")
	}

	if fi, ok := tm.Names["name"]; !ok {
		t.Error("expected 'name' field")
	} else if _, hasOmit := fi.Options["omitempty"]; !hasOmit {
		t.Error("expected 'omitempty' option on 'name' field")
	}
}

func TestMapper_Caching(t *testing.T) {
	type CacheStruct struct {
		ID int `db:"id"`
	}

	m := NewMapperFunc("db", strings.ToLower)

	tm1 := m.TypeMap(reflect.TypeFor[CacheStruct]())
	tm2 := m.TypeMap(reflect.TypeFor[CacheStruct]())

	if tm1 != tm2 {
		t.Error("expected TypeMap to return cached result")
	}
}

func TestMapper_UnexportedFields(t *testing.T) {
	type UnexportedStruct struct {
		ID       int    `db:"id"`
		name     string `db:"name"`     //nolint:unused
		internal int    `db:"internal"` //nolint:unused
	}

	m := NewMapperFunc("db", strings.ToLower)
	tm := m.TypeMap(reflect.TypeFor[UnexportedStruct]())

	if _, ok := tm.Names["id"]; !ok {
		t.Error("expected 'id' field")
	}
	if _, ok := tm.Names["name"]; ok {
		t.Error("should not have unexported 'name' field")
	}
	if _, ok := tm.Names["internal"]; ok {
		t.Error("should not have unexported 'internal' field")
	}
}

func TestMapper_PointerType(t *testing.T) {
	type PtrStruct struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	m := NewMapperFunc("db", strings.ToLower)
	tm := m.TypeMap(reflect.TypeFor[*PtrStruct]())

	if _, ok := tm.Names["id"]; !ok {
		t.Error("expected 'id' field from pointer type")
	}
	if _, ok := tm.Names["name"]; !ok {
		t.Error("expected 'name' field from pointer type")
	}
}

func TestMapper_TraversalsByName(t *testing.T) {
	type SimpleStruct struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	m := NewMapperFunc("db", strings.ToLower)
	traversals := m.TraversalsByName(reflect.TypeFor[SimpleStruct](), []string{"id", "name", "nonexistent"})

	if traversals[0] == nil {
		t.Error("expected traversal for 'id'")
	}
	if traversals[1] == nil {
		t.Error("expected traversal for 'name'")
	}
	if traversals[2] != nil {
		t.Error("expected nil traversal for 'nonexistent'")
	}
}

func TestTypeMapStampede(t *testing.T) {
	type Level4 struct {
		F40, F41, F42, F43, F44, F45, F46, F47, F48, F49 string
	}
	type Level3 struct {
		Level4
		F30, F31, F32, F33, F34, F35, F36, F37, F38, F39 string
	}
	type Level2 struct {
		Level3
		F20, F21, F22, F23, F24, F25, F26, F27, F28, F29 string
	}
	type Level1 struct {
		Level2
		F10, F11, F12, F13, F14, F15, F16, F17, F18, F19 string
	}
	type HeavyStampedeStruct struct {
		Level1
		F00, F01, F02, F03, F04, F05, F06, F07, F08, F09 string
	}

	var mapFuncCalls int
	var mapFuncMu sync.Mutex

	slowMapFunc := func(s string) string {
		mapFuncMu.Lock()
		mapFuncCalls++
		mapFuncMu.Unlock()
		return strings.ToLower(s)
	}

	m := NewMapperFunc("db", slowMapFunc)

	const numGoroutines = 1000
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	start := make(chan struct{})
	results := make([]*StructMap, numGoroutines)

	for i := range numGoroutines {
		go func(idx int) {
			defer wg.Done()
			<-start
			results[idx] = m.TypeMap(reflect.TypeFor[HeavyStampedeStruct]())
		}(i)
	}

	close(start)
	wg.Wait()

	first := results[0]
	if first == nil {
		t.Fatal("expected non-nil StructMap")
	}
	for i := 1; i < numGoroutines; i++ {
		if results[i] != first {
			t.Errorf("goroutine %d returned a different StructMap instance", i)
		}
	}

	expectedFields := 50 // 5 levels * 10 fields
	if mapFuncCalls != expectedFields {
		t.Errorf("expected mapFunc to be called exactly %d times, heavily mapped %d times. Singleflight failed.", expectedFields, mapFuncCalls)
	}
}

func TestNewMapper(t *testing.T) {
	m := NewMapper("db")
	if m.tagName != "db" {
		t.Error("expected tag name 'db'")
	}
}

func TestNewMapperTagFunc(t *testing.T) {
	m := NewMapperTagFunc("json", strings.ToUpper, strings.ToLower)
	if m.tagName != "json" {
		t.Error("expected tag name 'json'")
	}
	if m.mapFunc("id") != "ID" {
		t.Error("expected 'ID'")
	}
}

func TestMapper_FieldMap(t *testing.T) {
	type TestStruct struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}
	m := NewMapperFunc("db", strings.ToLower)
	fm := m.FieldMap(reflect.TypeFor[TestStruct]())
	if len(fm) != 2 {
		t.Errorf("expected 2 fields, got %d", len(fm))
	}
	if _, ok := fm["id"]; !ok {
		t.Error("missing 'id' field")
	}
}

func TestAddrByTraversal(t *testing.T) {
	type Inner struct {
		Value int
	}
	type Outer struct {
		*Inner
		ID int
	}
	m := NewMapperFunc("db", strings.ToLower)
	tm := m.TypeMap(reflect.TypeFor[Outer]())

	var o Outer
	v := reflect.ValueOf(&o).Elem()
	fi := tm.Names["value"]

	ptr := AddrByTraversal(v.Addr().UnsafePointer(), fi.Traversal)
	if ptr == nil {
		t.Fatal("expected non-nil pointer from AddrByTraversal")
	}

	// Value should be allocated since it was a nil pointer originally
	if o.Inner == nil {
		t.Fatal("expected Inner to be allocated")
	}
	o.Value = 42

	// Check that the returned pointer points to the same underlying data
	intPtr := (*int)(ptr)
	if *intPtr != 42 {
		t.Errorf("expected 42, got %d", *intPtr)
	}
}

func TestDeref(t *testing.T) {
	typ := reflect.TypeFor[*int]()
	derefTyp := Deref(typ)
	if derefTyp.Kind() != reflect.Int {
		t.Errorf("expected Int kind, got %v", derefTyp.Kind())
	}

	valTyp := reflect.TypeFor[int]()
	derefValTyp := Deref(valTyp)
	if derefValTyp.Kind() != reflect.Int {
		t.Errorf("expected Int kind, got %v", derefValTyp.Kind())
	}
}

func TestMapper_TraversalsByNameFunc(t *testing.T) {
	type TestStruct struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}
	m := NewMapperFunc("db", strings.ToLower)

	names := []string{"id", "nonexistent"}
	called := 0

	err := m.TraversalsByNameFunc(reflect.TypeFor[TestStruct](), names, func(i int, index []int) error {
		called++
		if i == 0 && len(index) != 1 {
			t.Errorf("expected traversal length 1 for 'id', got %d", len(index))
		}
		if i == 1 && len(index) != 0 {
			t.Errorf("expected traversal length 0 for 'nonexistent', got %d", len(index))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if called != 2 {
		t.Errorf("expected 2 calls, got %d", called)
	}
}

func TestFieldByIndexesReadOnly(t *testing.T) {
	type Inner struct {
		Value string `db:"value"`
	}
	type Outer struct {
		*Inner
		ID int `db:"id"`
	}

	m := NewMapperFunc("db", strings.ToLower)
	tm := m.TypeMap(reflect.TypeFor[Outer]())

	o := Outer{
		ID: 42,
	}

	v := reflect.ValueOf(&o).Elem()
	fi := tm.Names["value"]

	// Inner is nil, so it should return the zero value of string
	res := FieldByIndexesReadOnly(v, fi.Index)
	if res.String() != "" {
		t.Errorf("expected empty string, got '%s'", res.String())
	}

	o.Inner = &Inner{Value: "hello"}
	res = FieldByIndexesReadOnly(v, fi.Index)
	if res.String() != "hello" {
		t.Errorf("expected 'hello', got '%s'", res.String())
	}
}

func TestMapper_TraversalsByName_CachePoisoning(t *testing.T) {
	type VictimStruct struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	m := NewMapperFunc("db", strings.ToLower)
	
	// Query 1 gets traversals
	traversals1 := m.TraversalsByName(reflect.TypeFor[VictimStruct](), []string{"name"})
	if traversals1[0][0] != 1 {
		t.Fatalf("expected initial 'name' index to be 1, got %d", traversals1[0][0])
	}

	// Malicious actor modifies the returned traversal
	traversals1[0][0] = 99

	// Query 2 gets traversals, expecting the original uncorrupted value
	traversals2 := m.TraversalsByName(reflect.TypeFor[VictimStruct](), []string{"name"})
	if traversals2[0][0] == 99 {
		t.Fatalf("CACHE POISONED! TraversalsByName returned a mutated slice. Expected 1, got 99")
	}
}

func TestMapper_FieldMap_CachePoisoning(t *testing.T) {
	type VictimStruct struct {
		ID   int    `db:"id"`
		Name string `db:"name"`
	}

	m := NewMapperFunc("db", strings.ToLower)
	
	// Query 1 gets the FieldMap
	fieldMap1 := m.FieldMap(reflect.TypeFor[VictimStruct]())
	if _, ok := fieldMap1["name"]; !ok {
		t.Fatalf("expected 'name' in field map")
	}

	// Malicious actor modifies the returned map
	delete(fieldMap1, "name")

	// Query 2 gets the FieldMap
	fieldMap2 := m.FieldMap(reflect.TypeFor[VictimStruct]())
	if _, ok := fieldMap2["name"]; !ok {
		t.Fatalf("CACHE POISONED! FieldMap returned a mutated map where 'name' was deleted")
	}
}

func TestTypeMapCacheCollision(t *testing.T) {
	// Different packages/local scopes can have structurally distinct types with the same short name.
	// Since singleflight previously keyed on `t.String()`, this test ensures they don't collide.
	getA := func() reflect.Type {
		type User struct { ID int `db:"id"` }
		return reflect.TypeFor[User]()
	}
	getB := func() reflect.Type {
		type User struct { Count int `db:"count"` }
		return reflect.TypeFor[User]()
	}

	typeA, typeB := getA(), getB()
	if typeA.String() != typeB.String() {
		t.Fatalf("Test assumption failed: types should have the same String() representation, got %s and %s", typeA.String(), typeB.String())
	}

	mapper := NewMapper("db")
	startCh := make(chan struct{})
	var wg sync.WaitGroup
	workers := 1000
	wg.Add(workers * 2)

	for range workers {
		go func() {
			defer wg.Done()
			<-startCh
			res := mapper.TypeMap(typeA)
			if _, ok := res.Names["id"]; !ok {
				panic("mapA missing id (got mapB's fields on typeA!)")
			}
		}()

		go func() {
			defer wg.Done()
			<-startCh
			res := mapper.TypeMap(typeB)
			if _, ok := res.Names["count"]; !ok {
				panic("mapB missing count (got mapA's fields on typeB!)")
			}
		}()
	}

	close(startCh) // release all goroutines simultaneously
	wg.Wait()
}
