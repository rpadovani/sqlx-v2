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

package sqlx_test

type SmallStruct struct {
	ID    int64  `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

type MediumStruct struct {
	ID        int64   `db:"id"`
	FirstName string  `db:"first_name"`
	LastName  string  `db:"last_name"`
	Email     string  `db:"email"`
	Phone     string  `db:"phone"`
	Address   string  `db:"address"`
	City      string  `db:"city"`
	State     string  `db:"state"`
	Zip       string  `db:"zip"`
	Country   string  `db:"country"`
	Age       int64   `db:"age"`
	Score     float64 `db:"score"`
	Active    bool    `db:"active"`
	CreatedAt string  `db:"created_at"`
	UpdatedAt string  `db:"updated_at"`
}

type NestedStruct1 struct {
	Field10 string  `db:"field_10"`
	Field11 float64 `db:"field_11"`
	Field12 int64   `db:"field_12"`
	Field13 string  `db:"field_13"`
	Field14 float64 `db:"field_14"`
}

type NestedStruct2 struct {
	Field20 float64 `db:"field_20"`
	Field21 int64   `db:"field_21"`
	Field22 string  `db:"field_22"`
	Field23 float64 `db:"field_23"`
	Field24 int64   `db:"field_24"`
}

type EmbeddedStruct struct {
	Field30 int64   `db:"field_30"`
	Field31 string  `db:"field_31"`
	Field32 float64 `db:"field_32"`
	Field33 int64   `db:"field_33"`
	Field34 string  `db:"field_34"`
}

type LargeStruct struct {
	Field00 int64   `db:"field_00"`
	Field01 string  `db:"field_01"`
	Field02 float64 `db:"field_02"`
	Field03 int64   `db:"field_03"`
	Field04 string  `db:"field_04"`
	Field05 float64 `db:"field_05"`
	Field06 int64   `db:"field_06"`
	Field07 string  `db:"field_07"`
	Field08 float64 `db:"field_08"`
	Field09 int64   `db:"field_09"`

	Nested1 NestedStruct1 `db:""`

	Field15 int64   `db:"field_15"`
	Field16 string  `db:"field_16"`
	Field17 float64 `db:"field_17"`
	Field18 int64   `db:"field_18"`
	Field19 string  `db:"field_19"`

	NestedPtr *NestedStruct2 `db:""`

	Field25 string  `db:"field_25"`
	Field26 float64 `db:"field_26"`
	Field27 int64   `db:"field_27"`
	Field28 string  `db:"field_28"`
	Field29 float64 `db:"field_29"`

	*EmbeddedStruct

	Field35 float64 `db:"field_35"`
	Field36 int64   `db:"field_36"`
	Field37 string  `db:"field_37"`
	Field38 float64 `db:"field_38"`
	Field39 int64   `db:"field_39"`

	Field40 string  `db:"field_40"`
	Field41 float64 `db:"field_41"`
	Field42 int64   `db:"field_42"`
	Field43 string  `db:"field_43"`
	Field44 float64 `db:"field_44"`
	Field45 int64   `db:"field_45"`
	Field46 string  `db:"field_46"`
	Field47 float64 `db:"field_47"`
	Field48 int64   `db:"field_48"`
	Field49 string  `db:"field_49"`
}
