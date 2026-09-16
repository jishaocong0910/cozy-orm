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

package orm_test

import (
	"testing"
	"time"

	orm "github.com/jishaocong0910/cozy-orm"
	"github.com/stretchr/testify/require"
)

func TestTuple(t *testing.T) {
	r := require.New(t)
	tm := time.UnixMilli(1749198596000)
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("").ExpectQuery().WillReturnRows(mock.NewRows([]string{"id", "create_at", "level", "tags", "attributes", "category"}).
			AddRow(1, tm, 5, "aa,bb,cc", `{"key1": "value1","key2": "value2"}`, "{\"organization\": \"none\",\"class\": 1}"))
		tuples, err := db.Query[orm.Tuple6[*int64, *time.Time, *orm.UserLevel, orm.UserTags, orm.UserAttributes, *orm.UserCategory]](nil).BuildSql(func(b *orm.SqlBuilder) {}).Do()
		r.NoError(err)
		r.Equal(new(int64(1)), tuples[0].Field1)
		r.Equal(new(tm), tuples[0].Field2)
		r.Equal(new(orm.UserLevel("5")), tuples[0].Field3)
		r.Equal(orm.UserTags{"aa", "bb", "cc"}, tuples[0].Field4)
		r.Equal(orm.UserAttributes{"key1": "value1", "key2": "value2"}, tuples[0].Field5)
		r.Equal(new(orm.UserCategory{Organization: "none", Class: 1}), tuples[0].Field6)
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("").ExpectQuery().WillReturnRows(mock.NewRows([]string{"id", "create_at", "level", "tags", "attributes", "category"}).
			AddRow(1, tm, 5, "aa,bb,cc", `{"key1": "value1","key2": "value2"}`, "{\"organization\": \"none\",\"class\": 1}"))
		tuples, err := db.Query[orm.Tuple6[int64, time.Time, orm.UserLevel, *orm.UserTags, *orm.UserAttributes, orm.UserCategory]](nil).BuildSql(func(b *orm.SqlBuilder) {}).Do()
		r.NoError(err)
		r.Equal(int64(1), tuples[0].Field1)
		r.Equal(tm, tuples[0].Field2)
		r.Equal(orm.UserLevel("5"), tuples[0].Field3)
		r.Equal(new(orm.UserTags{"aa", "bb", "cc"}), tuples[0].Field4)
		r.Equal(new(orm.UserAttributes{"key1": "value1", "key2": "value2"}), tuples[0].Field5)
		r.Equal(orm.UserCategory{Organization: "none", Class: 1}, tuples[0].Field6)
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("").ExpectQuery().WillReturnRows(mock.NewRows([]string{"name", "phone", "properties", "tags"}).
			AddRow(nil, nil, nil, nil))
		tuples, err := db.Query[orm.Tuple4[string, *time.Time, orm.UserProperties, orm.UserCategory]](nil).BuildSql(func(b *orm.SqlBuilder) {}).Do()
		r.NoError(err)
		r.Equal("", tuples[0].Field1)
		r.Nil(tuples[0].Field2)
		r.Equal(orm.UserProperties{}, tuples[0].Field3)
		r.Equal(orm.UserCategory{}, tuples[0].Field4)
	}
}
