// Package reflectx is a legacy shim re-exporting internal/reflectx.
//
// Deprecated: Do not use this package in new code. It exists solely to maintain build compatibility with legacy jmoiron/sqlx consumers.
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

	"github.com/rpadovani/sqlx-v2/internal/reflectx"
)

type (
	Mapper    = reflectx.Mapper
	FieldInfo = reflectx.FieldInfo
	StructMap = reflectx.StructMap
)

func NewMapper(tag string) *Mapper {
	return reflectx.NewMapper(tag)
}

func NewMapperFunc(tag string, f func(string) string) *Mapper {
	return reflectx.NewMapperFunc(tag, f)
}

func NewMapperTagFunc(tag string, f func(string) string, mapFunc func(string) string) *Mapper {
	return reflectx.NewMapperTagFunc(tag, f, mapFunc)
}

func FieldByIndexes(v reflect.Value, indexes []int) reflect.Value {
	return reflectx.FieldByIndexes(v, indexes)
}

func FieldByIndexesReadOnly(v reflect.Value, indexes []int) reflect.Value {
	return reflectx.FieldByIndexesReadOnly(v, indexes)
}

func Deref(t reflect.Type) reflect.Type {
	return reflectx.Deref(t)
}
