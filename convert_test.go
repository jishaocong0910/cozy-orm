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
	"uuid"

	"github.com/stretchr/testify/require"
)

func TestFieldConv(t *testing.T) {
	r := require.New(t)
	var demo ConvDemo
	v := reflect.ValueOf(&demo).Elem()
	{
		field := v.FieldByName("Field1")
		c := getFieldConverter(field.Type())
		r.NotNil(c)
		c = getFieldConverter(reflect.TypeFor[*UserProperties]())
		r.NotNil(c)
		fc, ok := c.(ptrFieldConverter)
		r.True(ok)
		sd := fc.newScanDest()
		r.Equal(reflect.TypeFor[**string](), sd.p2pValue.Type())
		r.Equal(reflect.New(reflect.TypeFor[*string]()).Interface(), sd.dest())
		reflect.ValueOf(sd.dest()).Elem().Set(reflect.ValueOf((*string)(nil)))
		r.Nil(sd.value())
		reflect.ValueOf(sd.dest()).Elem().Set(reflect.ValueOf(new(`{"source":"a","country":"b"}`)))
		r.Equal(`{"source":"a","country":"b"}`, sd.value())
		fc.toField(field, `{"source":"a","country":"b"}`)
		r.Equal(&UserProperties{Source: "a", Country: "b"}, field.Interface().(*UserProperties))
	}
	{
		field := v.FieldByName("Field2")
		c := getFieldConverter(field.Type())
		r.NotNil(c)
		fc, ok := c.(valueFieldConverter)
		r.True(ok)
		sd := fc.newScanDest()
		r.Equal(reflect.TypeFor[**int8](), sd.p2pValue.Type())
		fc.toField(field, int8(3))
		r.Equal(UserLevel("3"), field.Interface().(UserLevel))
	}
	{
		field := v.FieldByName("Field3")
		c := getFieldConverter(field.Type())
		r.NotNil(c)
		fc, ok := c.(sliceFieldConverter)
		r.True(ok)
		sd := fc.newScanDest()
		r.Equal(reflect.TypeFor[**string](), sd.p2pValue.Type())
		fc.toField(field, "a,b,c")
		r.Equal(UserTags{"a", "b", "c"}, field.Interface().(UserTags))
	}
	{
		field := v.FieldByName("Field4")
		c := getFieldConverter(field.Type())
		r.NotNil(c)
		fc, ok := c.(mapFieldConverter)
		r.True(ok)
		sd := fc.newScanDest()
		r.Equal(reflect.TypeFor[**string](), sd.p2pValue.Type())
		fc.toField(field, `{"key1":"a","key2":"b"}`)
		r.Equal(UserAttributes{"key1": "a", "key2": "b"}, field.Interface().(UserAttributes))
	}
	{
		field := v.FieldByName("Field5")
		c := getFieldConverter(field.Type())
		r.NotNil(c)
		fc, ok := c.(zeroFieldConverter)
		r.True(ok)
		sd := fc.newScanDest()
		r.Equal(reflect.TypeFor[**string](), sd.p2pValue.Type())
		fc.toField(field, `test`)
		r.Equal("test", field.Interface().(string))
	}
	{
		tm := time.Now()
		field := v.FieldByName("Field6")
		c := getFieldConverter(field.Type())
		r.NotNil(c)
		fc, ok := c.(zeroFieldConverter)
		r.True(ok)
		sd := fc.newScanDest()
		r.Equal(reflect.TypeFor[**time.Time](), sd.p2pValue.Type())
		fc.toField(field, tm)
		r.Equal(tm, field.Interface().(time.Time))
	}
	{
		u := uuid.New()
		field := v.FieldByName("Field7")
		c := getFieldConverter(field.Type())
		r.NotNil(c)
		fc, ok := c.(zeroFieldConverter)
		r.True(ok)
		sd := fc.newScanDest()
		r.Equal(reflect.TypeFor[**uuid.UUID](), sd.p2pValue.Type())
		fc.toField(field, u)
		r.Equal(u, field.Interface().(uuid.UUID))
	}
	{
		field := v.FieldByName("Field8")
		c := getFieldConverter(field.Type())
		r.Nil(c)
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
	Field1  *UserProperties
	Field2  UserLevel
	Field3  UserTags
	Field4  UserAttributes
	Field5  string
	Field6  time.Time
	Field7  uuid.UUID
	Field8  UserCategory
	Field9  ConvBytes
	Field10 ConvTime
}

type ConvBytes struct {
	str string
}

func (c ConvBytes) ToValue() []byte {
	return []byte(c.str)
}

func (c *ConvBytes) ToField(val []byte) {
	c.str = string(val)
}

type ConvTime struct {
	str string
}

func (c ConvTime) ToValue() *time.Time {
	if t, err := time.Parse(time.DateTime, c.str); err != nil {
		return &t
	}
	return nil
}

func (c *ConvTime) ToField(val *time.Time) {
	c.str = val.Format(time.DateTime)
}
