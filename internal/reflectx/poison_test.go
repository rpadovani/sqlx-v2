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
	"testing"

	"github.com/rpadovani/sqlx-v2/internal/reflectx"
)

// IdentityA and IdentityB have the exact same memory layout
// but represent distinctly different semantic structs.
type IdentityA struct {
	SharedID int64 `db:"shared_id"`
}

type IdentityB struct {
	SharedID int64 `db:"shared_id"`
}

// IdentityC adds a distinct DB tag to an otherwise identical field
type IdentityC struct {
	SharedID int64 `db:"custom_id"`
}

func TestCachePoisonIsolation(t *testing.T) {
	m := reflectx.NewMapperFunc("db", strings.ToLower)

	// 1. Generate mappings for A, B, and C
	mapA := m.TypeMap(reflect.TypeFor[IdentityA]())
	mapB := m.TypeMap(reflect.TypeFor[IdentityB]())
	mapC := m.TypeMap(reflect.TypeFor[IdentityC]())

	// 2. Validate memory isolation of the cached StructMap objects
	if mapA == mapB {
		t.Fatal("SEVERE: Type cache collision! IdentityA and IdentityB share the same StructMap pointer.")
	}
	if mapA == mapC || mapB == mapC {
		t.Fatal("SEVERE: Type cache collision! IdentityC was entangled in the cache.")
	}

	// 3. Verify semantic isolation is preserved
	if _, ok := mapA.Names["shared_id"]; !ok {
		t.Fatal("IdentityA lost 'shared_id' mapping")
	}
	if _, ok := mapB.Names["shared_id"]; !ok {
		t.Fatal("IdentityB lost 'shared_id' mapping")
	}

	// C should be entirely disjoint
	if _, ok := mapC.Names["shared_id"]; ok {
		t.Fatal("IdentityC erroneously inherited the 'shared_id' tag from A/B!")
	}
	if _, ok := mapC.Names["custom_id"]; !ok {
		t.Fatal("IdentityC lost its distinct 'custom_id' mapping")
	}
}
