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
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFieldConv(t *testing.T) {
	r := require.New(t)
	var demo ConvDemo
	v := reflect.ValueOf(&demo).Elem()
	{
		field := v.FieldByName("Field1")
		fc := getFieldConverter(field.Type())
		r.NotNil(fc)
		fc = getFieldConverter(reflect.TypeFor[*ConvStruct]())
		r.NotNil(fc)
		r.Equal(reflect.TypeFor[*string](), fc.ptrMediumType)
		r.Equal(reflect.Pointer, fc.fieldKind)
		fc.convert(field, `{"field1":"a","field2":"b"}`)
		r.Equal(&ConvStruct{Field1: "a", Field2: "b"}, field.Interface().(*ConvStruct))
	}
	{
		field := v.FieldByName("Field2")
		fc := getFieldConverter(field.Type())
		r.NotNil(fc)
		r.Equal(reflect.TypeFor[*string](), fc.ptrMediumType)
		r.Equal(reflect.Int, fc.fieldKind)
		fc.convert(field, "3")
		r.Equal(ConvInt(3), field.Interface().(ConvInt))
	}
	{
		field := v.FieldByName("Field3")
		fc := getFieldConverter(field.Type())
		r.NotNil(fc)
		r.Equal(reflect.TypeFor[*string](), fc.ptrMediumType)
		r.Equal(reflect.Slice, fc.fieldKind)
		fc.convert(field, "a,b,c")
		r.Equal(ConvSlice{"a", "b", "c"}, field.Interface().(ConvSlice))
	}
	{
		field := v.FieldByName("Field4")
		fc := getFieldConverter(field.Type())
		r.NotNil(fc)
		r.Equal(reflect.TypeFor[*string](), fc.ptrMediumType)
		r.Equal(reflect.Map, fc.fieldKind)
		fc.convert(field, `{"key1":"a","key2":"b"}`)
		r.Equal(ConvMap{"key1": "a", "key2": "b"}, field.Interface().(ConvMap))
	}
	{
		field := v.FieldByName("Field5")
		fc := getFieldConverter(field.Type())
		r.Nil(fc)
	}
	{
		field := v.FieldByName("Field6")
		fc := getFieldConverter(field.Type())
		r.NotNil(fc)
		r.Equal(reflect.TypeFor[*[]byte](), fc.ptrMediumType)
		r.Equal(reflect.Struct, fc.fieldKind)
		fc.convert(field, []byte("abc"))
		r.Equal(ConvBytes{str: "abc"}, field.Interface().(ConvBytes))
	}
	{
		tm, _ := time.Parse(time.DateTime, "2026-09-23 23:50:43")
		field := v.FieldByName("Field7")
		fc := getFieldConverter(field.Type())
		r.NotNil(fc)
		r.Equal(reflect.TypeFor[**time.Time](), fc.ptrMediumType)
		r.Equal(reflect.Struct, fc.fieldKind)
		fc.convert(field, new(tm))
		r.Equal(ConvTime{str: "2026-09-23 23:50:43"}, field.Interface().(ConvTime))
	}
}

func TestConvertArgs(t *testing.T) {
	r := require.New(t)
	db, mock := MockDB(r)
	mock.ExpectPrepare("").ExpectQuery().WithArgs(nil, (*string)(nil), "test",
		`{"source":"a","country":"b"}`, `{"source":"a","country":"b"}`,
		1, 1,
		"a,b,c", "a,b,c",
		`{"key1":"a","key2":"b"}`, `{"key1":"a","key2":"b"}`,
	).WillReturnRows(mock.NewRows([]string{"unused"}))
	_, err := db.Query[User](nil).BuildSql(func(b *SqlBuilder) {
		b.Args(nil, (*string)(nil), "test",
			UserProperties{Source: "a", Country: "b"}, &UserProperties{Source: "a", Country: "b"},
			UserLevel("1"), new(UserLevel("1")),
			UserTags{"a", "b", "c"}, &UserTags{"a", "b", "c"},
			UserAttributes{"key1": "a", "key2": "b"}, &UserAttributes{"key1": "a", "key2": "b"},
		)
	}).Do()
	r.NoError(err)
	r.NoError(mock.ExpectationsWereMet())
}

type ConvDemo struct {
	Field1 *ConvStruct
	Field2 ConvInt
	Field3 ConvSlice
	Field4 ConvMap
	Field5 string
	Field6 ConvBytes
	Field7 ConvTime
}
