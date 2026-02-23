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

import "github.com/rpadovani/sqlx-v2/internal/bind"

var (
	// ErrBindMismatch is returned when there is a mismatch between query variables and arguments.
	ErrBindMismatch = bind.ErrBindMismatch

	// ErrNamedPropertyNotFound is returned when a named parameter cannot be found in the provided struct or map.
	ErrNamedPropertyNotFound = bind.ErrNamedPropertyNotFound

	// ErrUnsupportedType is returned when an unsupported type is provided for binding.
	ErrUnsupportedType = bind.ErrUnsupportedType

	// ErrSyntax is returned when there is a syntax error in the query parsing.
	ErrSyntax = bind.ErrSyntax

	// ErrColumnNotFound is returned when a column from the database cannot be mapped to a destination.
	ErrColumnNotFound = bind.ErrColumnNotFound
)
