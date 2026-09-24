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
	"context"
	"errors"
	"testing"
	"time"

	orm "github.com/jishaocong0910/cozy-orm"

	"github.com/DATA-DOG/go-sqlmock"

	"github.com/stretchr/testify/require"
)

func TestQuery(t *testing.T) {
	r := require.New(t)
	{
		db := orm.DBConfig{}.Build()
		_, err := db.Query[orm.User](nil).BuildSql(func(b *orm.SqlBuilder) {}).Do()
		r.EqualError(err, "no available *sql.DB")
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("").ExpectQuery().WillReturnRows(mock.NewRows([]string{"unused"}))
		_, err := db.Query[int](nil).BuildSql(func(b *orm.SqlBuilder) {}).Do()
		r.EqualError(err, "unsupported mapping type")
	}
	{
		db, _ := orm.MockDB(r)
		_, err := db.Query[orm.User](nil).BuildSql(func(b *orm.SqlBuilder) {
			b.Error(errors.New("build sql cause an error"))
		}).Do()
		r.EqualError(err, "build sql cause an error")
	}
	{
		db, _ := orm.MockDB(r)
		r.PanicsWithError("build sql cause an error", func() {
			db.Query[orm.User](nil).Must().BuildSql(func(b *orm.SqlBuilder) {
				b.Error(errors.New("build sql cause an error"))
			}).Do()
		})
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("").ExpectQuery().WillReturnRows(mock.NewRows([]string{"id"}).AddRow("abc"))
		_, err := db.Query[orm.User](nil).BuildSql(func(b *orm.SqlBuilder) {}).Do()
		r.EqualError(err, "sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"abc\") to a int64: invalid syntax")
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("").ExpectQuery().WillReturnRows(mock.NewRows([]string{"id"}).AddRow("abc"))
		_, err := db.Query[orm.User](nil).MapTarget(&orm.User{}).BuildSql(func(b *orm.SqlBuilder) {}).Do()
		r.EqualError(err, "sql: Scan error on column index 0, name \"id\": converting driver.Value type string (\"abc\") to a int64: invalid syntax")
	}
	{
		db, _ := orm.MockDB(r)
		entities, err := db.Query[orm.User](nil).Do()
		r.NoError(err)
		r.NotNil(entities)
		r.Len(entities, 0)
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("")
		_, err := db.Query[orm.User](nil).BuildSql(func(b *orm.SqlBuilder) {
			b.Write("SELECT * FROM user WHERE id = 1")
			b.Cancel()
		}).Do()
		r.NoError(err)
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, mock := orm.MockDB(r)
		log := orm.MockLogger(db)
		mock.ExpectPrepare("sql").ExpectQuery().WillReturnRows(mock.NewRows([]string{"unused"}))
		_, err := db.Query[orm.User](ctx).SqlLogLevel(orm.Level_.Info).Description("test desc").BuildSql(func(b *orm.SqlBuilder) {
			b.Write("sql")
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		lm := log.Msgs[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(orm.Level_.Info, lm.Level)
		r.Contains(lm.Msg, "SQL: sql; args: , tx: false, desc: test desc")
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("").ExpectQuery().WillReturnError(errors.New("_prepare error"))
		_, err := db.Query[orm.User](nil).BuildSql(func(b *orm.SqlBuilder) {}).Do()
		r.EqualError(err, "_prepare error")
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("").ExpectQuery().WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1).RowError(0, errors.New("rows error")))
		_, err := db.Query[orm.User](nil).BuildSql(func(b *orm.SqlBuilder) {}).Do()
		r.EqualError(err, "rows error")
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("").ExpectQuery().WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1).RowError(0, errors.New("rows error")))
		_, err := db.Query[orm.User](nil).MapTarget(&orm.User{}).BuildSql(func(b *orm.SqlBuilder) {}).Do()
		r.EqualError(err, "rows error")
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		createAt := time.UnixMilli(1703659380000)
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("SELECT * FROM user WHERE id = ? AND level = ? AND phone = ?").
			ExpectQuery().WithArgs(1, nil, nil).WillReturnRows(mock.NewRows([]string{"id", "name", "phone",
			"email", "unused", "avatar_url", "status", "level", "properties", "category", "tags", "attributes", "create_at"}).
			AddRow(1, "a", "123456", "example@gmail", "unused", []byte("jpg"), 2, 3, "{\"source\": \"web\",\"Country\": \"CN\"}",
				"{\"organization\": \"none\",\"class\": 1}", "a,b,c", "{\"key1\": \"value1\",\"key2\": \"value2\"}", createAt))
		users, err := db.Query[orm.User](nil).BuildSql(func(b *orm.SqlBuilder) {
			b.Write("SELECT * FROM user")
			b.Write(" WHERE id = ? AND level = ? AND phone = ?", new(1), nil, (*string)(nil))
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		user := users[0]
		r.Equal(int64(1), *user.Id)
		r.Equal("a", *user.Name)
		r.Equal("123456", *user.Phone)
		r.Equal("example@gmail", *user.Email)
		r.Equal(createAt, *user.CreateAt)
		r.Equal([]byte("jpg"), user.AvatarUrl)
		r.Equal(orm.UserStatus(2), *user.Status)
		r.Equal(orm.UserLevel("3"), *user.Level)
		r.Equal(orm.UserTags{"a", "b", "c"}, user.Tags)
		r.Equal(orm.UserProperties{Source: "web", Country: "CN"}, *user.Properties)
		r.Equal(orm.UserCategory{Organization: "none", Class: 1}, *user.Category)
		r.Equal(orm.UserTags{"a", "b", "c"}, user.Tags)
		r.Equal(orm.UserAttributes{"key1": "value1", "key2": "value2"}, user.Attributes)
	}
	{
		targets := []*orm.User{{}, {}}
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("").ExpectQuery().WillReturnRows(mock.NewRows([]string{"id"}).
			AddRow(1).AddRow(2))
		users, err := db.Query[orm.User](nil).BuildSql(func(b *orm.SqlBuilder) {}).MapTarget(targets...).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Nil(users)

		r.Equal(int64(1), *targets[0].Id)
		r.Equal(int64(2), *targets[1].Id)
	}
}

func TestMutation(t *testing.T) {
	r := require.New(t)
	{
		db := orm.DBConfig{}.Build()
		_, err := db.Mutation(nil).BuildSql(func(b *orm.SqlBuilder) {}).Do()
		r.EqualError(err, "no available *sql.DB")
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("").ExpectExec().WillReturnError(errors.New("exec error"))
		_, err := db.Mutation(nil).BuildSql(func(b *orm.SqlBuilder) {}).Do()
		r.EqualError(err, "exec error")
	}
	{
		db, _ := orm.MockDB(r)
		_, err := db.Mutation(nil).BuildSql(func(b *orm.SqlBuilder) {
			b.Error(errors.New("build sql cause an error"))
		}).Do()
		r.EqualError(err, "build sql cause an error")
	}
	{
		db, _ := orm.MockDB(r)
		r.PanicsWithError("build sql cause an error", func() {
			db.Mutation(nil).Must().BuildSql(func(b *orm.SqlBuilder) {
				b.Error(errors.New("build sql cause an error"))
			}).Do()
		})
	}
	{
		db, _ := orm.MockDB(r)
		affected, err := db.Mutation(nil).Do()
		r.NoError(err)
		r.Equal(int64(0), affected)
	}
	{
		db, _ := orm.MockDB(r)
		affected, err := db.Mutation(nil).BuildSql(func(b *orm.SqlBuilder) {
			b.Cancel()
		}).Do()
		r.NoError(err)
		r.Equal(int64(0), affected)
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, mock := orm.MockDB(r)
		log := orm.MockLogger(db)
		mock.ExpectPrepare("sql").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		_, err := db.Mutation(ctx).SqlLogLevel(orm.Level_.Info).Description("test desc").BuildSql(func(b *orm.SqlBuilder) {
			b.Write("sql")
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		lm := log.Msgs[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(orm.Level_.Info, lm.Level)
		r.Contains(lm.Msg, "SQL: sql; args: , tx: false, desc: test desc")
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("UPDATE user SET status = ?, level = ? WHERE id = ?").
			ExpectExec().WithArgs(1, 2, 1).WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := db.Mutation(nil).BuildSql(func(b *orm.SqlBuilder) {
			b.Write("UPDATE user SET status = ?, level = ? WHERE id = ?")
			b.Args(new(orm.UserLevel("1")), 2, 1)
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
	}
	{
		users := []*orm.User{{}, {}}
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DBConfig{SqlDB: sqlDB, GetGeneratedKeyMode: orm.GetGeneratedKeyMode_.FirstInsertId}.Build()
		mock.ExpectPrepare("").ExpectExec().WillReturnResult(sqlmock.NewResult(1, 1))
		affected, err := db.Mutation(nil).MapTarget(users...).BuildSql(func(b *orm.SqlBuilder) {
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
		r.Equal(int64(1), *users[0].Id)
		r.Equal(int64(2), *users[1].Id)
	}
	{
		users := []*orm.User{{}, {}}
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DBConfig{SqlDB: sqlDB, GetGeneratedKeyMode: orm.GetGeneratedKeyMode_.LastInsertId}.Build()
		mock.ExpectPrepare("").ExpectExec().WillReturnResult(sqlmock.NewResult(2, 1))
		affected, err := db.Mutation(nil).MapTarget(users...).BuildSql(func(b *orm.SqlBuilder) {
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
		r.Equal(int64(1), *users[0].Id)
		r.Equal(int64(2), *users[1].Id)
	}
	{
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DBConfig{SqlDB: sqlDB, GetGeneratedKeyMode: orm.GetGeneratedKeyMode_.FirstInsertId}.Build()
		log := orm.MockLogger(db)
		mock.ExpectPrepare("").ExpectExec().WillReturnResult(sqlmock.NewResult(2, 1))
		affected, err := db.Mutation(nil).MapTarget(new(1), new(1)).BuildSql(func(b *orm.SqlBuilder) {
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
		lm := log.Msgs[0]
		r.Equal(orm.Level_.Warn, lm.Level)
		r.Equal("get generated key fail, not a valid entity type", lm.Msg)
	}
	{
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DBConfig{SqlDB: sqlDB, GetGeneratedKeyMode: orm.GetGeneratedKeyMode_.FirstInsertId}.Build()
		log := orm.MockLogger(db)
		mock.ExpectPrepare("").ExpectExec().WillReturnResult(orm.UnsupportedLastInsertIdResult{})
		_, err := db.Mutation(nil).MapTarget(new(1), new(1)).BuildSql(func(b *orm.SqlBuilder) {
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		lm := log.Msgs[0]
		r.Equal(orm.Level_.Warn, lm.Level)
		r.Equal("get generated key fail, lastInsertId is not supported", lm.Msg)
	}
	{
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DBConfig{SqlDB: sqlDB, GetGeneratedKeyMode: orm.GetGeneratedKeyMode_.FirstInsertId}.Build()
		log := orm.MockLogger(db)
		mock.ExpectPrepare("").ExpectExec().WillReturnResult(sqlmock.NewResult(2, 1))
		affected, err := db.Mutation(nil).MapTarget(&orm.DemoMulPk{}, &orm.DemoMulPk{}).BuildSql(func(b *orm.SqlBuilder) {
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
		lm := log.Msgs[0]
		r.Equal(orm.Level_.Warn, lm.Level)
		r.Equal(`get generated key fail, the entity "orm.DemoMulPk" must have exactly one field with "auto" tag`, lm.Msg)
	}
}
