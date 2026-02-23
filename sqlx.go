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

package sqlx

import (
	"strings"

	"github.com/rpadovani/sqlx-v2/internal/bind"
)

// Rebind a query from QUESTION to the specified bind type.
func Rebind(bindType int, query string) string {
	return bind.Rebind(bindType, query)
}

// In expands slice values in args, returning the modified query string and a
// new arg list that can be executed by a database. The query should use the ?
// bindVar. The return value uses the ? bindVar.
func In(query string, args ...any) (string, []any, error) {
	return bind.In(query, args...)
}

// Named takes a query using named parameters and an argument and returns a
// new query with the named parameters replaced with positional parameters and
// the corresponding argument list.
func Named(query string, arg any) (string, []any, error) {
	return bind.Named(query, arg)
}

// NameMapper maps column names to struct field names. By default,
// it uses strings.ToLower to lowercase struct field names.  It can be set to
// whatever you want, but it is encouraged to be set before sqlx is used as
// name-to-field mappings are cached after first use on a type.
var NameMapper = strings.ToLower
