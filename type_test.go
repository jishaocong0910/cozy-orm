// Copyright 2024-present jishaocong0910
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

package orm

import (
	"database/sql/driver"
	"reflect"
	"testing"
	"time"
	_ "unsafe"

	"github.com/stretchr/testify/require"
)

func TestIsValidFieldType(t *testing.T) {
	r := require.New(t)
	r.True(isValidFieldType(reflect.TypeFor[*int]()))
	r.True(isValidFieldType(reflect.TypeFor[*time.Time]()))
	r.False(isValidFieldType(reflect.TypeFor[int]()))
	r.True(isValidFieldType(reflect.TypeFor[[]byte]()))
	r.False(isValidFieldType(reflect.TypeFor[[]int]()))
	r.False(isValidFieldType(reflect.TypeFor[*[]int]()))
	r.False(isValidFieldType(reflect.TypeFor[*complex64]()))
	r.True(isValidFieldType(reflect.TypeFor[*ConvInt]()))
	r.False(isValidFieldType(reflect.TypeFor[ConvInt]()))
	r.True(isValidFieldType(reflect.TypeFor[*ConvStruct]()))
	r.False(isValidFieldType(reflect.TypeFor[ConvStruct]()))
	r.True(isValidFieldType(reflect.TypeFor[ConvSlice]()))
	r.True(isValidFieldType(reflect.TypeFor[*ConvSlice]()))
	r.True(isValidFieldType(reflect.TypeFor[ConvMap]()))
	r.True(isValidFieldType(reflect.TypeFor[*ConvMap]()))
	r.False(isValidFieldType(reflect.TypeFor[ScannerValuerInt]()))
	r.True(isValidFieldType(reflect.TypeFor[*ScannerValuerInt]()))
	r.False(isValidFieldType(reflect.TypeFor[ScannerValuerStruct]()))
	r.True(isValidFieldType(reflect.TypeFor[*ScannerValuerStruct]()))
	r.True(isValidFieldType(reflect.TypeFor[ScannerValuerSlice]()))
	r.True(isValidFieldType(reflect.TypeFor[*ScannerValuerSlice]()))
	r.True(isValidFieldType(reflect.TypeFor[ScannerValuerMap]()))
	r.True(isValidFieldType(reflect.TypeFor[*ScannerValuerMap]()))
}

func TestIsImplementConvert(t *testing.T) {
	r := require.New(t)
	r.True(isImplementConverter(reflect.TypeFor[*ImplementConvertDemo]()))
	r.True(isImplementConverter(reflect.TypeFor[*ImplementConvertDemo2]()))
	r.False(isImplementConverter(reflect.TypeFor[*ImplementConvertDemo3]()))
	r.False(isImplementConverter(reflect.TypeFor[*ImplementConvertDemo4]()))
	r.False(isImplementConverter(reflect.TypeFor[*ImplementConvertDemo5]()))
	r.False(isImplementConverter(reflect.TypeFor[*ImplementConvertDemo6]()))
	r.False(isImplementConverter(reflect.TypeFor[*ImplementConvertDemo7]()))
	r.False(isImplementConverter(reflect.TypeFor[*ImplementConvertDemo8]()))
	r.False(isImplementConverter(reflect.TypeFor[*ImplementConvertDemo9]()))
	r.False(isImplementConverter(reflect.TypeFor[*ImplementConvertDemo10]()))
}

func TestIsTupleType(t *testing.T) {
	r := require.New(t)
	r.False(isTupleType(reflect.TypeFor[ConvStruct]()))
	r.True(isTupleType(reflect.TypeFor[Tuple[string]]()))
	r.True(isTupleType(reflect.TypeFor[Tuple2[string, string]]()))
	r.True(isTupleType(reflect.TypeFor[Tuple3[string, string, string]]()))
	r.True(isTupleType(reflect.TypeFor[Tuple4[string, string, string, string]]()))
	r.True(isTupleType(reflect.TypeFor[Tuple5[string, string, string, string, string]]()))
	r.True(isTupleType(reflect.TypeFor[Tuple6[string, string, string, string, string, string]]()))
}

func TestIsEntityType(t *testing.T) {
	r := require.New(t)
	r.True(isEntityType(reflect.TypeFor[ConvStruct]()))
	r.False(isEntityType(reflect.TypeFor[Tuple[string]]()))
	r.False(isEntityType(reflect.TypeFor[Tuple2[string, string]]()))
	r.False(isEntityType(reflect.TypeFor[Tuple3[string, string, string]]()))
	r.False(isEntityType(reflect.TypeFor[Tuple4[string, string, string, string]]()))
	r.False(isEntityType(reflect.TypeFor[Tuple5[string, string, string, string, string]]()))
	r.False(isEntityType(reflect.TypeFor[Tuple6[string, string, string, string, string, string]]()))
}

type ConvInt int

func (c ConvInt) ToArg() string {
	return ""
}

func (c *ConvInt) ToField(string) {
}

type ConvStruct struct {
}

func (c ConvStruct) ToArg() string {
	return ""
}

func (c *ConvStruct) ToField(string) {
}

type ConvSlice []string

func (m ConvSlice) ToArg() string {
	return ""
}

func (m *ConvSlice) ToField(string) {
}

type ConvMap map[string]string

func (c ConvMap) ToArg() string {
	return ""
}

func (c *ConvMap) ToField(string) {
}

type ScannerValuerInt struct {
}

func (s ScannerValuerInt) Value() (driver.Value, error) {
	return nil, nil
}

func (s ScannerValuerInt) Scan(src any) error {
	return nil
}

type ScannerValuerStruct int

func (s ScannerValuerStruct) Value() (driver.Value, error) {
	return nil, nil
}

func (s ScannerValuerStruct) Scan(src any) error {
	return nil
}

type ScannerValuerSlice []string

func (s ScannerValuerSlice) Value() (driver.Value, error) {
	return nil, nil
}

func (s ScannerValuerSlice) Scan(src any) error {
	return nil
}

type ScannerValuerMap map[string]string

func (s ScannerValuerMap) Value() (driver.Value, error) {
	return nil, nil
}

func (s ScannerValuerMap) Scan(src any) error {
	return nil
}

type ImplementConvertDemo struct {
}

func (d ImplementConvertDemo) ToArg() string {
	return ""
}

func (d *ImplementConvertDemo) ToField(string) {
}

type ImplementConvertDemo2 struct {
}

func (d *ImplementConvertDemo2) ToArg() string {
	return ""
}

func (d *ImplementConvertDemo2) ToField(string) {
}

type ImplementConvertDemo3 struct {
}

func (d *ImplementConvertDemo3) ToValue(string) string {
	return ""
}

func (d *ImplementConvertDemo3) ToField(t string) {
}

type ImplementConvertDemo4 struct {
}

func (d *ImplementConvertDemo4) ToValue() (string, string) {
	return "", ""
}

func (d *ImplementConvertDemo4) ToField(t string) {
}

type ImplementConvertDemo5 struct {
}

func (d *ImplementConvertDemo5) ToValue() ConvInt {
	return 0
}

func (d *ImplementConvertDemo5) ToField(t string) {
}

type ImplementConvertDemo6 struct {
}

func (d ImplementConvertDemo6) ToValue() string {
	return ""
}

func (d *ImplementConvertDemo6) ToField(t, t2 string) {
}

type ImplementConvertDemo7 struct {
}

func (d ImplementConvertDemo7) ToValue() int {
	return 0
}

func (d *ImplementConvertDemo7) ToField(t ConvInt) {
}

type ImplementConvertDemo8 struct {
}

func (d ImplementConvertDemo8) ToValue() string {
	return ""
}

func (d *ImplementConvertDemo8) ToField(t string) string {
	return ""
}

type ImplementConvertDemo9 struct {
}

func (d ImplementConvertDemo9) ToArg() string {
	return ""
}

func (d ImplementConvertDemo9) ToField(string) {
}

type ImplementConvertDemo10 struct {
}

func (d ImplementConvertDemo10) ToValue() string {
	return ""
}
