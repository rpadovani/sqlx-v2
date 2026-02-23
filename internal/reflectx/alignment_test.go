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
