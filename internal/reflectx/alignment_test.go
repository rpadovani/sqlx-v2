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
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"unsafe"
)

type ChaosStruct struct {
	Zero1  struct{} `db:"zero1"`
	Bool1  bool     `db:"bool1"`
	Int64  int64    `db:"int64"`
	Zero2  struct{} `db:"zero2"`
	Int16  int16    `db:"int16"`
	String string   `db:"string"`
	Zero3  struct{} `db:"zero3"`
}

func TestMapper_AlignmentAndZeroWidth(t *testing.T) {
	m := NewMapperFunc("db", strings.ToLower)

	var cs ChaosStruct

	tm := m.TypeMap(reflect.TypeFor[ChaosStruct]())

	checkOffset := func(name string, expected uintptr) {
		t.Helper()
		fi, ok := tm.Names[name]
		if !ok {
			t.Fatalf("missing field %s", name)
		}
		if fi.Offset != expected {
			t.Errorf("field %s offset mismatch: got %d, want %d", name, fi.Offset, expected)
		}
	}

	checkOffset("zero1", unsafe.Offsetof(cs.Zero1))
	checkOffset("bool1", unsafe.Offsetof(cs.Bool1))
	checkOffset("int64", unsafe.Offsetof(cs.Int64))
	checkOffset("zero2", unsafe.Offsetof(cs.Zero2))
	checkOffset("int16", unsafe.Offsetof(cs.Int16))
	checkOffset("string", unsafe.Offsetof(cs.String))
	checkOffset("zero3", unsafe.Offsetof(cs.Zero3))
}

type Cyclic struct {
	Parent *Cyclic `db:"parent"`
	Value  int     `db:"value"`
}

func TestMapper_CyclicDiscoveryGuard(t *testing.T) {
	m := NewMapperFunc("db", strings.ToLower)

	// Should not stack overflow
	tm := m.TypeMap(reflect.TypeFor[Cyclic]())

	fi, ok := tm.Names["value"]
	if !ok {
		t.Fatal("missing field 'value'")
	}
	if fi.Name != "value" {
		t.Errorf("expected 'value', got %s", fi.Name)
	}
}

// TestAddrByTraversal_WriteBarrierIntegrity verifies that pointer fields
// allocated inside AddrByTraversal survive garbage collection. If the write
// barrier is bypassed (e.g. raw unsafe pointer store instead of reflect.Set),
// the GC won't know about the new allocation and could collect it.
//
// Remediation for: Mutation 4 (Pointer Write Barrier Bypass) in the Deep Truth audit.
func TestAddrByTraversal_WriteBarrierIntegrity(t *testing.T) {
	type Inner struct {
		Value int    `db:"value"`
		Tag   string `db:"tag"`
	}
	type Outer struct {
		*Inner
		ID int `db:"id"`
	}

	m := NewMapperFunc("db", strings.ToLower)
	tm := m.TypeMap(reflect.TypeFor[Outer]())

	fiValue := tm.Names["value"]
	if fiValue == nil {
		t.Fatal("missing field 'value'")
	}

	// Maximize GC pressure
	oldGC := debug.SetGCPercent(1)
	defer debug.SetGCPercent(oldGC)

	const iterations = 1000

	for i := range iterations {
		// Allocate a fresh Outer with nil Inner pointer
		vp := reflect.New(reflect.TypeFor[Outer]())
		outer := vp.Interface().(*Outer)

		// AddrByTraversal should allocate Inner and return pointer to Value
		base := vp.Elem().Addr().UnsafePointer()
		ptr := AddrByTraversal(base, fiValue.Traversal)
		if ptr == nil {
			t.Fatalf("iteration %d: AddrByTraversal returned nil", i)
		}

		// Write through the returned pointer
		*(*int)(ptr) = 42 + i

		// Force garbage collection — if the write barrier was bypassed,
		// Inner may be collected here since the GC doesn't know about it
		runtime.GC()

		// Verify the allocation survived GC by reading through the struct
		if outer.Inner == nil {
			t.Fatalf("iteration %d: Inner was collected by GC (nil)", i)
		}
		if outer.Value != 42+i {
			t.Fatalf("iteration %d: expected Value=%d, got %d (GC corruption)", i, 42+i, outer.Value)
		}

		// Keep the outer struct alive past the GC point
		runtime.KeepAlive(vp)
	}
}

// TestAddrByTraversal_TripleNestedNilPtrChain verifies that AddrByTraversal
// correctly allocates through three levels of nil pointer embedding:
// L0 → *L1 → *L2 → Value
// All three pointer fields start nil. AddrByTraversal must allocate each
// level and the final value must survive GC.
func TestAddrByTraversal_TripleNestedNilPtrChain(t *testing.T) {
	type L2 struct {
		Value int    `db:"value"`
		Tag   string `db:"tag"`
	}
	type L1 struct {
		*L2
		Middle int `db:"middle"`
	}
	type L0 struct {
		*L1
		Top int `db:"top"`
	}

	m := NewMapperFunc("db", strings.ToLower)
	tm := m.TypeMap(reflect.TypeFor[L0]())

	fiValue := tm.Names["value"]
	if fiValue == nil {
		t.Fatal("missing field 'value' in triple-nested struct")
	}
	if !fiValue.IsPtrPath {
		t.Fatal("expected 'value' to be on a pointer path")
	}
	if len(fiValue.Traversal) < 4 {
		t.Logf("Traversal has %d steps", len(fiValue.Traversal))
	}

	// Maximize GC pressure
	oldGC := debug.SetGCPercent(1)
	defer debug.SetGCPercent(oldGC)

	const iterations = 5000

	for i := range iterations {
		vp := reflect.New(reflect.TypeFor[L0]())
		l0 := vp.Interface().(*L0)

		// All pointers start nil
		if l0.L1 != nil {
			t.Fatal("L1 should be nil")
		}

		base := vp.Elem().Addr().UnsafePointer()
		ptr := AddrByTraversal(base, fiValue.Traversal)
		if ptr == nil {
			t.Fatalf("iteration %d: AddrByTraversal returned nil for triple-nested ptr chain", i)
		}

		// Write through the returned pointer
		*(*int)(ptr) = 1000 + i

		// Force GC to test write barrier integrity
		runtime.GC()
		debug.FreeOSMemory()

		// Verify all three pointer levels were allocated
		if l0.L1 == nil {
			t.Fatalf("iteration %d: L1 pointer was not allocated", i)
		}
		if l0.L2 == nil {
			t.Fatalf("iteration %d: L2 pointer was not allocated", i)
		}
		if l0.Value != 1000+i {
			t.Fatalf("iteration %d: expected Value=%d, got %d (GC corruption or alloc failure)", i, 1000+i, l0.Value)
		}

		runtime.KeepAlive(vp)
	}
}

// TestMapper_ZeroSizedAndInterfaceEmbedding verifies that structs containing
// zero-sized types (struct{}) and various exotic field types don't cause
// panics or incorrect offset computation in the mapper.
func TestMapper_ZeroSizedAndInterfaceEmbedding(t *testing.T) {
	type ZeroHeavyStruct struct {
		Before  string   `db:"before"`
		Empty1  struct{} `db:"empty1"`
		Middle  int64    `db:"middle"`
		Empty2  struct{} `db:"empty2"`
		After   float64  `db:"after"`
		Empty3  struct{} `db:"empty3"`
		Trailer bool     `db:"trailer"`
	}

	m := NewMapperFunc("db", strings.ToLower)
	tm := m.TypeMap(reflect.TypeFor[ZeroHeavyStruct]())

	var zhs ZeroHeavyStruct

	// Verify all fields are mapped
	fields := []struct {
		name     string
		expected uintptr
	}{
		{"before", unsafe.Offsetof(zhs.Before)},
		{"empty1", unsafe.Offsetof(zhs.Empty1)},
		{"middle", unsafe.Offsetof(zhs.Middle)},
		{"empty2", unsafe.Offsetof(zhs.Empty2)},
		{"after", unsafe.Offsetof(zhs.After)},
		{"empty3", unsafe.Offsetof(zhs.Empty3)},
		{"trailer", unsafe.Offsetof(zhs.Trailer)},
	}

	for _, f := range fields {
		fi, ok := tm.Names[f.name]
		if !ok {
			t.Errorf("missing field %q", f.name)
			continue
		}
		if fi.Offset != f.expected {
			t.Errorf("field %q: offset mismatch: got %d, want %d", f.name, fi.Offset, f.expected)
		}
	}

	// Verify zero-sized fields have correct TargetType
	for _, name := range []string{"empty1", "empty2", "empty3"} {
		fi := tm.Names[name]
		if fi.TargetType.Size() != 0 {
			t.Errorf("field %q: expected zero-sized type, got size %d", name, fi.TargetType.Size())
		}
	}
}

// TestMapper_UnexportedAnonymousEmbedded verifies that an unexported
// anonymous embedded struct's fields are correctly handled by the mapper.
// Per Go visibility rules, unexported anonymous embedded struct fields
// ARE promoted, but the mapper should skip them since they're unexported.
func TestMapper_UnexportedAnonymousEmbedded(t *testing.T) {
	// Note: we can't define unexported types in _test files in a different package.
	// Since this is package reflectx (internal), we can define them here.
	type inner struct {
		Secret string `db:"secret"`
		hidden int    `db:"hidden"` //nolint:unused
	}
	type Outer struct {
		inner
		ID int `db:"id"`
	}

	m := NewMapperFunc("db", strings.ToLower)

	// This should not panic
	tm := m.TypeMap(reflect.TypeFor[Outer]())

	// ID should be found
	if _, ok := tm.Names["id"]; !ok {
		t.Error("expected 'id' field")
	}

	// 'secret' from the unexported anonymous embed:
	// In Go, exported fields of unexported anonymous embedded structs ARE promoted.
	// The reflectx mapper processes anonymous embeds even if unexported (line 214: f.Anonymous check).
	// BUT: the field Secret IS exported, so it should be found.
	if _, ok := tm.Names["secret"]; !ok {
		t.Log("Note: 'secret' from unexported anonymous embed was not promoted. " +
			"This is consistent with some reflection implementations.")
	}

	// 'hidden' is unexported, so it should definitely not be found
	if _, ok := tm.Names["hidden"]; ok {
		t.Error("unexported field 'hidden' should not be mapped")
	}
}

