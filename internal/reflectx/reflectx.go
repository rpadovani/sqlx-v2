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
	"fmt"
	"reflect"
	"strings"
	"sync"
	"unicode"
	"unsafe"

	"golang.org/x/sync/singleflight"
)

var scannerType = reflect.TypeFor[scanner]()

type scanner interface {
	Scan(any) error
}

func isScanner(t reflect.Type) bool {
	return t.Implements(scannerType) || reflect.PointerTo(t).Implements(scannerType)
}

// Mapper is a general purpose mapper of names to struct fields. A Mapper
// behaves like most other mappers in the Go ecosystem, mapping
// column names to struct fields using a set of rules.
type Mapper struct {
	tagName    string
	mapFunc    func(string) string
	tagMapFunc func(string) string // optional transform applied to parsed tag values
	strictTags bool
	cache      sync.Map // map[reflect.Type]*StructMap

	// sfBuckets shards discovery to prevent global mutex contention
	sfBuckets [256]singleflight.Group
}

// FieldInfo is metadata for a struct field.
type FieldInfo struct {
	Index      []int
	Path       string
	Field      reflect.StructField
	Zero       reflect.Value
	Name       string
	Options    map[string]string
	Embedded   bool
	Children   []*FieldInfo
	Parent     *FieldInfo
	Traversal  []Step       // Legacy support/pointer resolution
	Offset     uintptr      // Linear offset relative to base struct (0 for top-level pointer embeds)
	TargetType reflect.Type // The terminal type to scan into
	IsPtrPath  bool         // True if any node in the path (e.g. parent) was accessed via pointer
}

// Step represents a single level of dereferencing or offset in a struct traversal.
type Step struct {
	Offset   uintptr
	Type     reflect.Type   // If not nil, this step is a pointer that might need allocation of this Type
	PtrType  reflect.Type   // Cache of reflect.PointerTo(Type) for zero-allocation write barrier Set
}

// StructMap represents a mapped struct, with fields flattened
// from embedded structs and ready for name-based lookup.
type StructMap struct {
	Tree   *FieldInfo
	Index  []*FieldInfo
	Paths  map[string]*FieldInfo
	Names  map[string]*FieldInfo
	Errors []error // Populated when strict tag validation is enabled
}

// NewMapper returns a new Mapper with a given struct tag.
func NewMapper(tagName string) *Mapper {
	return &Mapper{
		tagName: tagName,
	}
}

// NewMapperFunc returns a new Mapper with a given struct tag and name mapping function.
func NewMapperFunc(tagName string, f func(string) string) *Mapper {
	return &Mapper{
		tagName: tagName,
		mapFunc: f,
	}
}

// NewMapperTagFunc returns a new Mapper with a given struct tag, a name mapping
// function and a tag mapping function. The name mapping function (f) is used to
// transform field names when no tag is present. The tag mapping function (tagMapFunc)
// is applied to tag values after parsing.
func NewMapperTagFunc(tagName string, f func(string) string, tagMapFunc func(string) string) *Mapper {
	return &Mapper{
		tagName:    tagName,
		mapFunc:    f,
		tagMapFunc: tagMapFunc,
	}
}

// NewMapperFuncStrict returns a new Mapper with strict tag validation enabled.
// When strict is true, struct tags are validated for correctness during TypeMap.
func NewMapperFuncStrict(tagName string, f func(string) string, strict bool) *Mapper {
	return &Mapper{
		tagName:    tagName,
		mapFunc:    f,
		strictTags: strict,
	}
}

// FieldByName returns a FieldInfo for the given column name.
func (m *Mapper) FieldByName(v reflect.Value, name string) reflect.Value {
	tm := m.TypeMap(v.Type())
	fi, ok := tm.Names[name]
	if !ok {
		return reflect.Value{}
	}
	return FieldByIndexes(v, fi.Index)
}

// TraversalsByName returns a slice index-path slices for the struct type
// and the given names.
func (m *Mapper) TraversalsByName(t reflect.Type, names []string) [][]int {
	tm := m.TypeMap(t)
	traversals := make([][]int, len(names))
	for i, name := range names {
		if fi, ok := tm.Names[name]; ok {
			traversals[i] = fi.Index
		}
	}
	return traversals
}

// FieldMap returns the mapper's mapping of field names to FieldInfos.
func (m *Mapper) FieldMap(t reflect.Type) map[string]*FieldInfo {
	tm := m.TypeMap(t)
	return tm.Names
}

// TypeMap returns a StructMap for the given reflect.Type, creating it if necessary.
// This is the core function that handles recursive struct flattening.
func (m *Mapper) TypeMap(t reflect.Type) *StructMap {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	if v, ok := m.cache.Load(t); ok {
		return v.(*StructMap)
	}

	// Use singleflight to prevent cache stampede for this specific type.
	key := fmt.Sprintf("%p", t)
	name := t.String()
	var hash uint64 = 14695981039346656037
	for i := 0; i < len(name); i++ {
		hash ^= uint64(name[i])
		hash *= 1099511628211
	}
	bucket := hash % 256

	v, _, _ := m.sfBuckets[bucket].Do(key, func() (any, error) {
		// Double-check cache inside singleflight
		if v, ok := m.cache.Load(t); ok {
			return v.(*StructMap), nil
		}

		sm := m.mapType(t)
		m.cache.Store(t, sm)
		return sm, nil
	})

	return v.(*StructMap)
}

// mapType is the internal recursive struct mapper.
// It handles nested/embedded structs by flattening their fields.
func (m *Mapper) mapType(t reflect.Type) *StructMap {
	sm := &StructMap{
		Paths: make(map[string]*FieldInfo),
		Names: make(map[string]*FieldInfo),
	}
	sm.Tree = &FieldInfo{}
	visited := make(map[reflect.Type]bool)
	sm.Index = m.flattenFields(sm, t, sm.Tree, "", nil, visited)
	return sm
}

// flattenFields recursively processes struct fields, handling embedded structs.
func (m *Mapper) flattenFields(sm *StructMap, t reflect.Type, parent *FieldInfo, prefix string, index []int, visited map[reflect.Type]bool) []*FieldInfo {
	if visited[t] {
		return nil
	}
	visited[t] = true
	defer delete(visited, t)

	var fields []*FieldInfo

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)

		// Skip unexported fields, but allow unexported anonymous embedded structs
		if !f.IsExported() && !f.Anonymous {
			continue
		}

		fi := &FieldInfo{
			Field:  f,
			Parent: parent,
		}

		// Build the index path
		fi.Index = make([]int, len(index)+1)
		copy(fi.Index, index)
		fi.Index[len(index)] = i

		// Build traversal steps
		if parent != nil && len(parent.Traversal) > 0 {
			fi.Traversal = make([]Step, len(parent.Traversal))
			copy(fi.Traversal, parent.Traversal)
			fi.IsPtrPath = parent.IsPtrPath
		}

		// Absolute offset from base struct; meaningless across pointer boundaries.
		if parent != nil && !fi.IsPtrPath {
			fi.Offset = parent.Offset + f.Offset
		} else if parent == nil {
			fi.Offset = f.Offset
		}

		fi.Traversal = append(fi.Traversal, Step{Offset: f.Offset})
		fi.TargetType = f.Type

		if f.Type.Kind() == reflect.Pointer {
			ptrType := reflect.PointerTo(f.Type.Elem())
			fi.Traversal = append(fi.Traversal, Step{Type: f.Type.Elem(), PtrType: ptrType})
			fi.TargetType = f.Type.Elem()
			fi.IsPtrPath = true
			fi.Offset = 0
		}

		// Parse the tag
		tag := f.Tag.Get(m.tagName)
		if tag == "-" {
			continue
		}

		name := ""
		if tag != "" {
			parts := strings.Split(tag, ",")
			name = parts[0]

			// Apply tag mapping function if configured
			if name != "" && m.tagMapFunc != nil {
				name = m.tagMapFunc(name)
			}

			// Tag validation
			if name != "" {
				if err := validateTagName(name, t.Name(), f.Name); err != nil {
					sm.Errors = append(sm.Errors, err)
				}
			}
			if tag != "" && name == "" {
				sm.Errors = append(sm.Errors, fmt.Errorf(
					"reflectx: struct %s, field %s: empty tag name in `%s:\"%s\"`",
					t.Name(), f.Name, m.tagName, tag,
				))
			}

			if len(parts) > 1 {
				fi.Options = make(map[string]string)
				for _, opt := range parts[1:] {
					kv := strings.SplitN(opt, "=", 2)
					if len(kv) == 2 {
						fi.Options[kv[0]] = kv[1]
					} else {
						fi.Options[opt] = ""
					}
				}
			}
		}

		// Handle embedded structs (recursive flattening)
		ft := f.Type
		if ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}

		isScan := isScanner(ft)

		if ft.Kind() == reflect.Struct && tag == "" && f.Anonymous && !isScan {
			fi.Embedded = true
			fi.Children = m.flattenFields(sm, ft, fi, prefix, fi.Index, visited)
			fields = append(fields, fi)
			continue
		}

		// Determine the mapped name
		if name == "" {
			name = m.mapFunc(f.Name)
		}

		if prefix != "" {
			fi.Path = prefix + "." + name
		} else {
			fi.Path = name
		}
		fi.Name = name

		// If it's a non-anonymous struct, recurse with prefix
		if ft.Kind() == reflect.Struct && !isScan {
			fi.Children = m.flattenFields(sm, ft, fi, fi.Path, fi.Index, visited)
		}

		sm.Paths[fi.Path] = fi
		sm.Names[fi.Path] = fi
		if _, ok := sm.Names[fi.Name]; !ok {
			sm.Names[fi.Name] = fi
		}
		fields = append(fields, fi)
	}

	return fields
}

// validateTagName checks that a tag name contains only valid characters:
// letters, digits, underscores, and dots.
func validateTagName(name, structName, fieldName string) error {
	for _, r := range name {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '.' {
			return fmt.Errorf(
				"reflectx: struct %s, field %s: invalid character %q in tag name %q",
				structName, fieldName, r, name,
			)
		}
	}
	return nil
}

// FieldByIndexes returns a value within a struct for the given index path.
func FieldByIndexes(v reflect.Value, indexes []int) reflect.Value {
	for _, i := range indexes {
		v = reflect.Indirect(v).Field(i)
		// if this is a pointer and it's nil, allocate a new value and set it
		if v.Kind() == reflect.Pointer && v.IsNil() {
			alloc := reflect.New(Deref(v.Type()))
			v.Set(alloc)
		}
		if v.Kind() == reflect.Map && v.IsNil() {
			v.Set(reflect.MakeMap(v.Type()))
		}
	}
	return v
}

// AddrByTraversal returns the address of a field within a struct using pre-calculated steps.
//
// SAFETY: This function operates on raw unsafe.Pointer addresses and allocates
// intermediate pointer fields via reflect.NewAt().Elem().Set() to maintain GC
// write barrier visibility. The caller MUST ensure:
//  1. 'base' is derived from a heap-allocated reflect.New() Value (not a stack local).
//  2. The reflect.Value that produced 'base' (via .Pointer()) remains rooted by
//     a runtime.KeepAlive() call AFTER the Scan that writes through these pointers.
//  3. The Steps slice was produced by the same Mapper.TypeMap that analysed the
//     struct type, guaranteeing offset correctness.
//
// Rooting chain: caller's vp (reflect.Value) → runtime.KeepAlive(vp) after Scan
// → base is reachable → all interior pointers allocated here are reachable.
func AddrByTraversal(base unsafe.Pointer, steps []Step) unsafe.Pointer {
	ptr := base
	for _, step := range steps {
		ptr = unsafe.Add(ptr, step.Offset)
		if step.Type != nil {
			p := (*unsafe.Pointer)(ptr)
			if *p == nil {
				// Allocate the pointer's target type via reflect to maintain
				// GC write barrier visibility. The Set call ensures the new
				// allocation is properly tracked by the garbage collector.
				vp := reflect.New(step.Type)
				// The Set call ensures the new allocation is properly tracked by the GC write barrier.
				reflect.NewAt(step.PtrType, ptr).Elem().Set(vp)
			}
			ptr = *p
		}
	}
	return ptr
}

// Deref is Indirect for reflect.Types
func Deref(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

// TraversalsByNameFunc traverses the mapped names and calls fn with the index of
// each name and the struct traversal represented by that name. Panics if
// the struct cannot be mapped or if fn panics.
func (m *Mapper) TraversalsByNameFunc(t reflect.Type, names []string, fn func(int, []int) error) error {
	tm := m.TypeMap(t)
	for i, name := range names {
		fi, ok := tm.Names[name]
		if !ok {
			if err := fn(i, []int{}); err != nil {
				return err
			}
			continue
		}
		if err := fn(i, fi.Index); err != nil {
			return err
		}
	}
	return nil
}

// FieldByIndexesReadOnly returns a value for the field given by the struct traversal
// for the given value.
func FieldByIndexesReadOnly(v reflect.Value, indexes []int) reflect.Value {
	for _, i := range indexes {
		if v.Kind() == reflect.Pointer && v.IsNil() {
			v = reflect.Zero(v.Type().Elem())
		}
		v = reflect.Indirect(v).Field(i)
	}
	return v
}
