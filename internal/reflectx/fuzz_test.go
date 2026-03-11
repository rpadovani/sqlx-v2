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

package reflectx_test

import (
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/rpadovani/sqlx-v2/internal/reflectx"
)

// types that can be safely embedded as fields for fuzzing
var baseTypes = []reflect.Type{
	reflect.TypeFor[int](),
	reflect.TypeFor[int8](),
	reflect.TypeFor[int16](),
	reflect.TypeFor[int32](),
	reflect.TypeFor[int64](),
	reflect.TypeFor[uint](),
	reflect.TypeFor[uint8](),
	reflect.TypeFor[uint16](),
	reflect.TypeFor[uint32](),
	reflect.TypeFor[uint64](),
	reflect.TypeFor[uintptr](),
	reflect.TypeFor[float32](),
	reflect.TypeFor[float64](),
	reflect.TypeFor[complex64](),
	reflect.TypeFor[complex128](),
	reflect.TypeFor[bool](),
	reflect.TypeFor[string](),
	reflect.TypeFor[[]byte](),
	reflect.TypeFor[[]int](),
	reflect.TypeFor[map[string]int](),
	// Zero-sized struct
	reflect.TypeOf(struct{}{}),
}

var (
	typeCache sync.Map
	typeCount int32
)

func buildDynamicStruct(data []byte) reflect.Type {
	if len(data) == 0 {
		return reflect.TypeOf(struct{}{})
	}

	key := string(data)
	if v, ok := typeCache.Load(key); ok {
		return v.(reflect.Type)
	}

	if atomic.LoadInt32(&typeCount) > 5000 {
		return reflect.TypeOf(struct{}{})
	}

	var fields []reflect.StructField
	seen := make(map[string]bool)

	for i := 0; i < len(data) && len(fields) < 50; i++ {
		b := data[i]

		t := baseTypes[int(b)%len(baseTypes)]
		if b%11 == 0 {
			t = reflect.PointerTo(t)
		}

		name := string(rune('A' + (b % 26)))

		for seen[name] {
			name += "X"
		}
		seen[name] = true

		f := reflect.StructField{
			Name: name,
			Type: t,
		}

		fields = append(fields, f)
	}

	if len(fields) == 0 {
		return reflect.TypeOf(struct{}{})
	}

	var result reflect.Type
	func() {
		defer func() { _ = recover() }()
		result = reflect.StructOf(fields)
	}()

	if result != nil {
		atomic.AddInt32(&typeCount, 1)
		typeCache.Store(key, result)
	}
	return result
}

func FuzzTypeMap(f *testing.F) {
	// Seed corpus with some basic fuzz values
	f.Add([]byte{0, 1, 2, 3})
	f.Add([]byte{255, 0, 127})
	f.Add([]byte{10, 10, 10}) // Anonymous triggers
	f.Add([]byte{})

	f.Fuzz(func(t *testing.T, data []byte) {
		m := reflectx.NewMapperFunc("db", strings.ToLower)

		// Construct a random struct based on the byte payload
		tType := buildDynamicStruct(data)
		if tType == nil {
			return
		}

		// Ask the mapper to map it
		tm := m.TypeMap(tType)

		// Assertions to verify offset boundary calculus
		size := tType.Size()
		for _, fi := range tm.Names {
			if size == 0 {
				continue // Zero-sized structs (e.g. struct{ p struct{} }) are edge cases with 0 bounds
			}
			if fi.IsPtrPath {
				continue // pointer paths exist outside the literal contiguous memory of the struct
			}

			// Target 1 Invariant: Offset + sizeof(Field) <= Total Size
			fieldSize := fi.TargetType.Size()
			endBound := fi.Offset + fieldSize

			if endBound > size {
				t.Fatalf("OOB Boundary Violation! Field %s spans offset %d to %d (size %d), but total Struct Size is only %d",
					fi.Name, fi.Offset, endBound, fieldSize, size)
			}
		}
	})
}
