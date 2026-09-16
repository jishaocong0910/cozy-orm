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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	r := require.New(t)
	{
		log := &mockLogger{}
		sqlDB, mock := MockSqlDB(r)
		db := DBConfig{SqlDB: sqlDB, Logger: log}.Build()
		mock.ExpectPrepare("test").ExpectQuery().WillReturnRows(mock.NewRows([]string{"id"}))
		_, err := db.Query[User](nil).SqlLogLevel(Level_.Info).BuildSql(func(b *SqlBuilder) {
			b.Write("test")
		}).Do()
		r.NoError(err)
		r.Equal(sqlDB, db.Raw())
		r.Len(log.Msgs, 1)
		lm := log.Msgs[0]
		r.Equal(Level_.Info, lm.Level)
		r.Contains(lm.Msg, "test")
	}
	{
		sqlDB, mock := MockSqlDB(r)
		db := DBConfig{SqlDB: sqlDB, ParamPrefix: ":"}.Build()
		mock.ExpectPrepare(":1:2:3").ExpectQuery().WillReturnRows(mock.NewRows([]string{"id"}))
		_, err := db.Query[User](nil).BuildSql(func(b *SqlBuilder) {
			b.WritePh().WritePh().WritePh()
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		db := DBConfig{DBType: DBType_.MySQL}.Build()
		r.Equal("", db.paramPrefix)
		r.Equal(QuotedIdentifier_.Backtick.ID, db.quotedIdentifier.ID)
		r.Equal(GetGeneratedKeyMode_.FirstInsertId.ID, db.GetGeneratedKeyMode.ID)
		r.Equal(PageMode_.LimitOffset.ID, db.pageMode.ID)
	}
	{
		db := DBConfig{DBType: DBType_.Oracle}.Build()
		r.Equal(":", db.paramPrefix)
		r.Equal(QuotedIdentifier_.DoubleQuote.ID, db.quotedIdentifier.ID)
		r.Equal(GetGeneratedKeyMode_.Oracle.ID, db.GetGeneratedKeyMode.ID)
		r.Equal(PageMode_.OffsetFetch.ID, db.pageMode.ID)
	}
	{
		db := DBConfig{DBType: DBType_.Postgres}.Build()
		r.Equal("$", db.paramPrefix)
		r.Equal(QuotedIdentifier_.DoubleQuote.ID, db.quotedIdentifier.ID)
		r.Equal(GetGeneratedKeyMode_.InsertReturning.ID, db.GetGeneratedKeyMode.ID)
		r.Equal(PageMode_.LimitOffset.ID, db.pageMode.ID)
	}
	{
		db := DBConfig{DBType: DBType_.SQLServer}.Build()
		r.Equal(":", db.paramPrefix)
		r.Equal(QuotedIdentifier_.Bracket.ID, db.quotedIdentifier.ID)
		r.Equal(GetGeneratedKeyMode_.SQLServer.ID, db.GetGeneratedKeyMode.ID)
		r.Equal(PageMode_.OffsetFetch.ID, db.pageMode.ID)
	}
	{
		db := DBConfig{DBType: DBType_.SQLite}.Build()
		r.Equal("", db.paramPrefix)
		r.Equal(QuotedIdentifier_.Backtick.ID, db.quotedIdentifier.ID)
		r.Equal(GetGeneratedKeyMode_.LastInsertId.ID, db.GetGeneratedKeyMode.ID)
		r.Equal(PageMode_.LimitOffset.ID, db.pageMode.ID)
	}

}
