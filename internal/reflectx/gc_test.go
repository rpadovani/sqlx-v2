package reflectx

import (
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"unsafe"
)

type CanaryStruct struct {
	MagicValue int
}

const canaryMagic = 0xCAFEBA11

//go:noinline
func assignCanary(base unsafe.Pointer, traversal []Step, i int) {
	ptr := AddrByTraversal(base, traversal)
	*(*CanaryStruct)(ptr) = CanaryStruct{MagicValue: canaryMagic + i}
}

func TestGC_WriteBarrierViolation(t *testing.T) {
	type NestedStruct struct {
		Canary *CanaryStruct `db:"canary"`
	}

	type Row struct {
		ID     int           `db:"id"`
		Nested *NestedStruct `db:"nested"`
	}

	m := NewMapperFunc("db", strings.ToLower)
	tm := m.TypeMap(reflect.TypeFor[Row]())

	fiCanary := tm.Names["nested.canary"]
	if fiCanary == nil {
		t.Fatal("missing field 'nested.canary'")
	}
	
	oldGC := debug.SetGCPercent(1)
	defer debug.SetGCPercent(oldGC)

	for i := 0; i < 10000; i++ {
		rowVal := reflect.New(reflect.TypeFor[Row]())
		rowPtr := rowVal.Elem().Addr().UnsafePointer()

		assignCanary(rowPtr, fiCanary.Traversal, i)

		runtime.GC()
		debug.FreeOSMemory()

		// Allocate garbage to overwrite reclaimed memory
		garbage := make([][]byte, 100)
		for j := range garbage {
			garbage[j] = make([]byte, 1024)
			for k := range garbage[j] {
				garbage[j][k] = 0xFF
			}
		}

		row := rowVal.Interface().(*Row)
		if row.Nested == nil {
			t.Fatalf("iteration %d: Nested pointer is nil (GC reclaimed it!)", i)
		}
		if row.Nested.Canary == nil {
			t.Fatalf("iteration %d: Canary pointer is nil (GC reclaimed it!)", i)
		}
		if row.Nested.Canary.MagicValue != canaryMagic + i {
			t.Fatalf("iteration %d: Canary value corrupted, got %x but expected %x", i, row.Nested.Canary.MagicValue, canaryMagic + i)
		}

		runtime.KeepAlive(garbage)
		runtime.KeepAlive(rowVal)
	}
}
