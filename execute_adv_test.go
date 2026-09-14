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
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	orm "github.com/jishaocong0910/cozy-orm"
	"github.com/stretchr/testify/require"
)

func TestFind(t *testing.T) {
	r := require.New(t)
	{
		db, _ := orm.MockDB(r)
		r.PanicsWithError("not a valid entity type", func() {
			db.Find[int](nil).Must().Do()
		})
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{SqlDB: sqlDB, PageType: orm.PageType_.LimitOffset}.Build()
		log := orm.MockLogger(db)
		mock.ExpectPrepare("SELECT id, name, phone, email, level, version FROM user WHERE level = ? ORDER BY id DESC, name ASC LIMIT 1 OFFSET 10 FOR UPDATE").ExpectQuery().WithArgs(1).
			WillReturnRows(mock.NewRows([]string{"id", "name"}).AddRow(9, "abc").AddRow(10, "efg"))
		users, err := db.Find[orm.User](ctx).SqlLogLevel(orm.Level_.Info).Describe("test desc").Select("id", "version").OnDemand(orm.DemandFor[orm.UserSimple]()).
			Condition(orm.Cond().Eq("level", 1)).OrderBy(orm.OrderBy().Desc("id").Asc("name")).
			Page(orm.Page(10, 1)).LastStr("FOR UPDATE").Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(9), *users[0].Id)
		r.Equal("abc", *users[0].Name)
		r.Equal(int64(10), *users[1].Id)
		r.Equal("efg", *users[1].Name)
		lm := log.Msgs[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(orm.Level_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
	{
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{
			SqlDB:    sqlDB,
			PageType: orm.PageType_.FetchNext,
		}.Build()
		mock.ExpectPrepare("SELECT id, name FROM user OFFSET 10 ROWS FETCH NEXT 1 ROWS ONLY").
			ExpectQuery().WillReturnRows(mock.NewRows([]string{"id", "name"}).AddRow(9, "abc"))
		_, err := db.Find[orm.User](nil).Select("id", "name").Page(orm.Page(10, 1)).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{
			SqlDB: sqlDB,
			ColumnPolicyConfigs: orm.ColumnPolicyConfigs{
				orm.NewColumnPolicyConfig("deleted").OnDeleteSoftly().PkMode(0),
			},
		}.Build()
		mock.ExpectPrepare("SELECT id, name FROM user WHERE level = ? AND deleted = ?").ExpectQuery().WithArgs(1, 0).
			WillReturnRows(mock.NewRows([]string{}))
		_, err := db.Find[orm.User](nil).Select("id", "name").Condition(orm.Cond().Eq("level", 1)).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		mock.ExpectPrepare("SELECT id, name FROM user WHERE level = ?").ExpectQuery().WithArgs(1).
			WillReturnRows(mock.NewRows([]string{}))
		_, err = db.Find[orm.User](nil).Select("id", "name").Condition(orm.Cond().Eq("level", 1)).IncludeDeleted().Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
}

func TestFindOne(t *testing.T) {
	r := require.New(t)
	{
		db, _ := orm.MockDB(r)
		r.PanicsWithError("not a valid entity type", func() {
			db.FindOne[int](nil).Must().Do()
		})
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("SELECT id, name, phone, email, avatar_url, status, level, properties, category, " +
			"tags, attributes, uuid, create_at, update_at, version, deleted FROM user").ExpectQuery().
			WillReturnRows(mock.NewRows([]string{}).AddRow().AddRow())
		_, err := db.FindOne[orm.User](nil).Do()
		r.Error(err, "return more than one row")
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{SqlDB: sqlDB, PageType: orm.PageType_.LimitOffset}.Build()
		log := orm.MockLogger(db)
		mock.ExpectPrepare("SELECT id, name, phone, email, level, version FROM user " +
			"WHERE level = ? ORDER BY id DESC, name ASC LIMIT 1 OFFSET 10 FOR UPDATE").ExpectQuery().WithArgs(1).
			WillReturnRows(mock.NewRows([]string{"id", "name"}).AddRow(9, "abc").AddRow(10, "efg"))
		user, err := db.FindOne[orm.User](ctx).SqlLogLevel(orm.Level_.Info).Describe("test desc").Compatible().Select("id", "version").
			OnDemand(orm.DemandFor[orm.UserSimple]()).Condition(orm.Cond().Eq("level", 1)).OrderBy(orm.OrderBy().Desc("id").Asc("name")).
			Page(orm.Page(10, 1)).LastStr("FOR UPDATE").Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(9), *user.Id)
		r.Equal("abc", *user.Name)
		lm := log.Msgs[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(orm.Level_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
	{
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{
			SqlDB: sqlDB,
			ColumnPolicyConfigs: orm.ColumnPolicyConfigs{
				orm.NewColumnPolicyConfig("deleted").OnDeleteSoftly().PkMode(0),
			},
		}.Build()
		mock.ExpectPrepare("SELECT id, name FROM user WHERE level = ? AND deleted = ?").ExpectQuery().WithArgs(1, 0).
			WillReturnRows(mock.NewRows([]string{}))
		_, err := db.FindOne[orm.User](nil).Select("id", "name").Condition(orm.Cond().Eq("level", 1)).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		mock.ExpectPrepare("SELECT id, name FROM user WHERE level = ?").ExpectQuery().WithArgs(1).
			WillReturnRows(mock.NewRows([]string{}))
		_, err = db.FindOne[orm.User](nil).Select("id", "name").Condition(orm.Cond().Eq("level", 1)).IncludeDeleted().Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
}

func TestInsertBatch(t *testing.T) {
	r := require.New(t)
	{
		db, _ := orm.MockDB(r)
		r.PanicsWithError("not a valid entity type", func() {
			db.Insert[int](nil).Must().Entities(new(1)).Do()
		})
	}
	{
		db, _ := orm.MockDB(r)
		affected, err := db.Insert[orm.User](nil).Do()
		r.Equal(int64(0), affected)
		r.NoError(err)
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, mock := orm.MockDB(r)
		log := orm.MockLogger(db)
		mock.ExpectPrepare("INSERT INTO user(name) VALUES (?) ON DUPLICATE KEY UPDATE id = id").ExpectExec().WithArgs("abc").WillReturnResult(sqlmock.NewResult(1, 1))
		affected, err := db.Insert[orm.User](ctx).SqlLogLevel(orm.Level_.Info).Describe("test desc").
			Entities(&orm.User{Name: new("abc")}).LastStr("ON DUPLICATE KEY UPDATE id = id").Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
		lm := log.Msgs[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(orm.Level_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("INSERT INTO user(name, phone, email, status) VALUES (?, ?, ?, NULL)").ExpectExec().
			WithArgs("abc", "phone", "email").WillReturnResult(sqlmock.NewResult(1, 1))
		_, err := db.Insert[orm.User](nil).Entities(&orm.User{Name: new("abc"), Phone: new("phone"), Email: new("email")}).Nullable("status").Do()
		r.NoError(err)
	}
	{
		u1 := &orm.User{Name: new("name1")}
		u2 := &orm.User{Name: new("name2")}
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{
			SqlDB:      sqlDB,
			GenKeyType: orm.GenKeyType_.FirstInsertId,
			ColumnPolicyConfigs: orm.ColumnPolicyConfigs{
				orm.NewColumnPolicyConfig("create_at").UseCreateTime(),
			},
		}.Build()
		mock.ExpectPrepare("INSERT INTO user(name, create_at) VALUES (?, ?), (?, ?)").ExpectExec().
			WillReturnResult(sqlmock.NewResult(1, 2)).
			WithArgs("name1", orm.AnyTime{}, "name2", orm.AnyTime{})
		_, err := db.Insert[orm.User](nil).Entities(u1, u2).Do()
		r.NoError(err)
		r.Equal(int64(1), *u1.Id)
		r.Equal(int64(2), *u2.Id)
	}
	{
		u1 := &orm.User{Name: new("name1")}
		u2 := &orm.User{Name: new("name2")}
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{SqlDB: sqlDB, GenKeyType: orm.GenKeyType_.LastInsertId}.Build()
		mock.ExpectPrepare("INSERT INTO user(name) VALUES (?), (?)").ExpectExec().WillReturnResult(sqlmock.NewResult(3, 2)).
			WithArgs("name1", "name2")
		_, err := db.Insert[orm.User](nil).Entities(u1, u2).Do()
		r.NoError(err)
		r.Equal(int64(2), *u1.Id)
		r.Equal(int64(3), *u2.Id)
	}
	{
		u1 := &orm.User{Name: new("name1")}
		u2 := &orm.User{Name: new("name2")}
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{SqlDB: sqlDB, GenKeyType: orm.GenKeyType_.Returning}.Build()
		mock.ExpectPrepare("INSERT INTO user(name) VALUES (?), (?) RETURNING id").ExpectQuery().
			WithArgs("name1", "name2").WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
		_, err := db.Insert[orm.User](nil).Entities(u1, u2).Do()
		r.NoError(err)
		r.Equal(int64(1), *u1.Id)
		r.Equal(int64(2), *u2.Id)
	}
	{
		u1 := &orm.User{Name: new("name1")}
		u2 := &orm.User{Name: new("name2")}
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{SqlDB: sqlDB, GenKeyType: orm.GenKeyType_.Output}.Build()
		mock.ExpectPrepare("INSERT INTO user(name) OUTPUT INSERTED.id VALUES (?), (?)").ExpectQuery().
			WithArgs("name1", "name2").WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
		_, err := db.Insert[orm.User](nil).Entities(u1, u2).Do()
		r.NoError(err)
		r.Equal(int64(1), *u1.Id)
		r.Equal(int64(2), *u2.Id)
	}
}

func TestUpdate(t *testing.T) {
	r := require.New(t)
	{
		db, _ := orm.MockDB(r)
		r.PanicsWithError("not a valid entity type", func() {
			db.Update[int](nil).Must().Entity(new(1)).Do()
		})
	}
	{
		db, mock := orm.MockDB(r)
		affected, err := db.Update[orm.User](nil).Entity(&orm.User{Phone: new("123")}).Do()
		r.Equal(int64(0), affected)
		r.EqualError(err, "full table modification blocked")

		mock.ExpectPrepare("UPDATE user SET phone = ?").ExpectExec().WithArgs("123").
			WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err = db.Update[orm.User](nil).Entity(&orm.User{Phone: new("123")}).SkipSafety().Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		db, _ := orm.MockDB(r)
		affected, err := db.Update[orm.User](nil).Do()
		r.Equal(int64(0), affected)
		r.NoError(err)
	}
	{
		db, _ := orm.MockDB(r)
		affected, err := db.Update[orm.User](nil).Entity(&orm.User{}).Do()
		r.Equal(int64(0), affected)
		r.NoError(err)
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, mock := orm.MockDB(r)
		log := orm.MockLogger(db)
		mock.ExpectPrepare("UPDATE user SET name = ? WHERE id = ?").
			ExpectExec().WithArgs("abc", 1).WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := db.Update[orm.User](ctx).SqlLogLevel(orm.Level_.Info).Describe("test desc").
			Entity(&orm.User{Id: new(int64(1)), Name: new("abc")}).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
		lm := log.Msgs[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(orm.Level_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("UPDATE user SET name = ?, phone = NULL, email = NULL, status = ?, level = '2', properties = NULL, tags = NULL WHERE id = ? AND status = ?").
			ExpectExec().WithArgs("abc", 2, 1, 4).WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := db.Update[orm.User](ctx).Entity(&orm.User{Id: new(int64(1)), Name: new("abc"), Category: &orm.UserCategory{Organization: "none", Class: 1}}).
			OnDemand(orm.DemandFor[orm.UserSimple]()).Nullable("properties").Set("status", orm.UserStatus(2)).Set("tags", "'t1,t2'").Set("tags", nil).
			SetRaw("level", "'2'").
			Condition(orm.Cond().Eq("status", 4)).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
	}
	{
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{
			SqlDB: sqlDB,
			ColumnPolicyConfigs: orm.ColumnPolicyConfigs{
				orm.NewColumnPolicyConfig("create_at").UseCreateTime(),
				orm.NewColumnPolicyConfig("update_at").UseUpdateTime(),
				orm.NewColumnPolicyConfig("version").UseRowVersion(),
				orm.NewColumnPolicyConfig("deleted").OnDeleteSoftly().PkMode(0),
			},
		}.Build()

		mock.ExpectPrepare("UPDATE user SET name = ?, update_at = ?, version = version + 1 WHERE id = ? AND deleted = ?").ExpectExec().
			WithArgs("abc", orm.AnyTime{}, 1, 0).
			WillReturnResult(sqlmock.NewResult(0, 0))
		_, err := db.Update[orm.User](nil).Entity(&orm.User{Name: new("abc")}).Condition(orm.Cond().Eq("id", 1)).Do()
		r.NoError(err)

		mock.ExpectPrepare("UPDATE user SET name = ?, update_at = ?, version = version + 1 WHERE id = ?").ExpectExec().
			WithArgs("abc", orm.AnyTime{}, 1).
			WillReturnResult(sqlmock.NewResult(0, 0))
		_, err = db.Update[orm.User](nil).Entity(&orm.User{Id: new(int64(1)), Name: new("abc"), CreateAt: new(time.Now()), Version: new(int64(5))}).
			Nullable("update_at").IncludeDeleted().Do()
		r.NoError(err)
	}
}

func TestUpdateRows(t *testing.T) {
	r := require.New(t)
	{
		db, _ := orm.MockDB(r)
		r.PanicsWithError("not a valid entity type", func() {
			db.UpdateRow[int](nil).Must().Entities(new(1)).Do()
		})
	}
	{
		db, _ := orm.MockDB(r)
		affected, err := db.UpdateRow[orm.DemoMulPk](nil).Entities(&orm.DemoMulPk{}, &orm.DemoMulPk{}).Do()
		r.Equal(int64(0), affected)
		r.EqualError(err, "orm.DemoMulPk\" must have exactly one field with the \"pk\" tag")
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("")
		affected, err := db.UpdateRow[orm.User](nil).Entities(&orm.User{Name: new("abc")}).Do()
		r.Equal(int64(0), affected)
		r.EqualError(err, "field 'Id' is nil at index 0 of entities")
	}
	{
		db, _ := orm.MockDB(r)
		affected, err := db.UpdateRow[orm.User](nil).Do()
		r.Equal(int64(0), affected)
		r.NoError(err)
	}
	{
		db, _ := orm.MockDB(r)
		affected, err := db.UpdateRow[orm.User](nil).Entities(&orm.User{Id: new(int64(1))}, &orm.User{Id: new(int64(2))}).Do()
		r.Equal(int64(0), affected)
		r.NoError(err)
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, mock := orm.MockDB(r)
		log := orm.MockLogger(db)
		mock.ExpectPrepare("UPDATE user SET name = CASE id WHEN ? THEN ? WHEN ? THEN ? END, phone = NULL, "+
			"email = CASE id WHEN ? THEN NULL WHEN ? THEN ? END, status = ?, level = '2', properties = NULL, tags = NULL "+
			"WHERE id IN(?, ?) AND level = ?").ExpectExec().
			WithArgs(1, "a", 2, "b", 1, 2, "email", 2, 1, 2, 5).WillReturnResult(sqlmock.NewResult(1, 1))
		affected, err := db.UpdateRow[orm.User](ctx).SqlLogLevel(orm.Level_.Info).Describe("test desc").OnDemand(orm.DemandFor[orm.UserSimple]()).
			Entities(&orm.User{Id: new(int64(1)), Name: new("a")}, &orm.User{Id: new(int64(2)), Name: new("b"), Email: new("email")}).
			Nullable("properties").Set("status", orm.UserStatus(2)).Set("tags", nil).SetRaw("level", "'2'").
			Condition(orm.Cond().Eq("level", 5)).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
		lm := log.Msgs[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(orm.Level_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
	{
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{
			SqlDB: sqlDB,
			ColumnPolicyConfigs: orm.ColumnPolicyConfigs{
				orm.NewColumnPolicyConfig("create_at").UseCreateTime(),
				orm.NewColumnPolicyConfig("update_at").UseUpdateTime(),
				orm.NewColumnPolicyConfig("version").UseRowVersion(),
				orm.NewColumnPolicyConfig("deleted").OnDeleteSoftly().PkMode(0),
			},
		}.Build()
		mock.ExpectPrepare("UPDATE user SET name = CASE id WHEN ? THEN ? WHEN ? THEN ? END, update_at = ?, version = version + 1 WHERE id IN(?, ?) AND level = ? AND deleted = ?").ExpectExec().
			WithArgs(1, "a", 2, "b", orm.AnyTime{}, 1, 2, 5, 0).
			WillReturnResult(sqlmock.NewResult(1, 1))
		affected, err := db.UpdateRow[orm.User](nil).Entities(&orm.User{Id: new(int64(1)), Name: new("a")}, &orm.User{Id: new(int64(2)), Name: new("b")}).
			Condition(orm.Cond().Eq("level", 5)).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)

		mock.ExpectPrepare("UPDATE user SET name = CASE id WHEN ? THEN ? WHEN ? THEN ? END, update_at = ?, version = version + 1 WHERE id IN(?, ?)").ExpectExec().
			WithArgs(1, "a", 2, "b", orm.AnyTime{}, 1, 2).
			WillReturnResult(sqlmock.NewResult(1, 1))
		affected, err = db.UpdateRow[orm.User](nil).Entities(&orm.User{Id: new(int64(1)), Name: new("a")}, &orm.User{Id: new(int64(2)), Name: new("b"), Version: new(int64(5))}).
			Nullable("update_at", "version").IncludeDeleted().Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, mock := orm.MockDB(r)
		mock.ExpectPrepare("UPDATE user SET name = ?, phone = NULL, email = NULL, status = ?, level = '2', properties = NULL, tags = NULL "+
			"WHERE id = ? AND level = ?").ExpectExec().
			WithArgs("a", 2, 1, 5).WillReturnResult(sqlmock.NewResult(1, 1))
		affected, err := db.UpdateRow[orm.User](ctx).SqlLogLevel(orm.Level_.Info).Describe("test desc").OnDemand(orm.DemandFor[orm.UserSimple]()).
			Entities(&orm.User{Id: new(int64(1)), Name: new("a")}, ).Nullable("properties").
			Set("status", orm.UserStatus(2)).Set("tags", nil).SetRaw("level", "'2'").
			Condition(orm.Cond().Eq("level", 5)).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
	}
}

func TestDelete(t *testing.T) {
	r := require.New(t)
	{
		db, _ := orm.MockDB(r)
		r.PanicsWithError("not a valid entity type", func() {
			db.Delete[int](nil).Must().Do()
		})
	}
	{
		db, mock := orm.MockDB(r)
		affected, err := db.Delete[orm.User](nil).Do()
		r.Equal(int64(0), affected)
		r.EqualError(err, "full table modification blocked")

		mock.ExpectPrepare("DELETE FROM user").ExpectExec().WithoutArgs().WillReturnResult(sqlmock.NewResult(0, 2))
		affected, err = db.Delete[orm.User](nil).SkipSafety().Do()
		r.Equal(int64(2), affected)
		r.NoError(err)
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, mock := orm.MockDB(r)
		log := orm.MockLogger(db)
		mock.ExpectPrepare("DELETE FROM user WHERE id = ?").ExpectExec().WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := db.Delete[orm.User](ctx).SqlLogLevel(orm.Level_.Info).Describe("test desc").Condition(orm.Cond().Eq("id", 1)).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
		lm := log.Msgs[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(orm.Level_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
}

func TestDeleteSoftly(t *testing.T) {
	r := require.New(t)
	{
		db, _ := orm.MockDB(r)
		r.PanicsWithError("not a valid entity type", func() {
			db.DeleteSoftly[int](nil).Must().Do()
		})
	}
	{
		db, _ := orm.MockDB(r)
		r.PanicsWithError("feature is not supported", func() {
			db.DeleteSoftly[orm.User](nil).Must().Do()
		})
	}
	{
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{
			SqlDB: sqlDB,
			ColumnPolicyConfigs: orm.ColumnPolicyConfigs{
				orm.NewColumnPolicyConfig("deleted").OnDeleteSoftly().PkMode(0),
			},
		}.Build()
		affected, err := db.DeleteSoftly[orm.User](nil).Do()
		r.EqualError(err, "full table modification blocked")
		r.Equal(int64(0), affected)

		mock.ExpectPrepare("UPDATE user SET deleted = id").ExpectExec().WithoutArgs().WillReturnResult(sqlmock.NewResult(0, 2))
		affected, err = db.DeleteSoftly[orm.User](nil).SkipSafety().Do()
		r.NoError(err)
		r.Equal(int64(2), affected)
	}
	{
		sqlDB, _ := orm.MockSqlDB(r)
		db := orm.DbInstConfig{
			SqlDB: sqlDB,
			ColumnPolicyConfigs: orm.ColumnPolicyConfigs{
				orm.NewColumnPolicyConfig("deleted").OnDeleteSoftly().PkMode(0),
			},
		}.Build()
		_, err := db.DeleteSoftly[int](nil).Do()
		r.EqualError(err, "not a valid entity type")
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{
			SqlDB: sqlDB,
			ColumnPolicyConfigs: orm.ColumnPolicyConfigs{
				orm.NewColumnPolicyConfig("deleted").OnDeleteSoftly().PkMode(0),
			},
		}.Build()
		log := orm.MockLogger(db)
		mock.ExpectPrepare("UPDATE user SET deleted = id WHERE id = ? AND deleted = ?").ExpectExec().WithArgs(1, 0).WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := db.DeleteSoftly[orm.User](ctx).SqlLogLevel(orm.Level_.Info).Describe("test desc").Condition(orm.Cond().Eq("id", 1)).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
		lm := log.Msgs[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(orm.Level_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
	{
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{
			SqlDB: sqlDB,
			ColumnPolicyConfigs: orm.ColumnPolicyConfigs{
				orm.NewColumnPolicyConfig("deleted").OnDeleteSoftly().PkMode(0),
			},
		}.Build()
		mock.ExpectPrepare("UPDATE user SET deleted = id WHERE id = ? AND deleted = ?").ExpectExec().WithArgs(1, 0).
			WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := db.DeleteSoftly[orm.User](nil).Condition(orm.Cond().Eq("id", 1)).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
	}
	{
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{
			SqlDB: sqlDB,
			ColumnPolicyConfigs: orm.ColumnPolicyConfigs{
				orm.NewColumnPolicyConfig("deleted").OnDeleteSoftly().NullMode(0),
			},
		}.Build()
		mock.ExpectPrepare("UPDATE user SET deleted = NULL WHERE id = ? AND deleted = ?").ExpectExec().WithArgs(1, 0).
			WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := db.DeleteSoftly[orm.User](nil).Condition(orm.Cond().Eq("id", 1)).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
	}

}

func TestCount(t *testing.T) {
	r := require.New(t)
	{
		db, _ := orm.MockDB(r)
		r.PanicsWithError("not a valid entity type", func() {
			db.Count[int](nil).Must().Do()
		})
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, mock := orm.MockDB(r)
		log := orm.MockLogger(db)
		mock.ExpectPrepare("SELECT COUNT(*) FROM user WHERE level = ?").ExpectQuery().WithArgs(2).WillReturnRows(mock.NewRows([]string{"count"}).AddRow(10))
		count, err := db.Count[orm.User](ctx).SqlLogLevel(orm.Level_.Info).Describe("test desc").Condition(orm.Cond().Eq("level", 2)).Do()
		r.NoError(err)
		r.Equal(int64(10), count)
		lm := log.Msgs[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(orm.Level_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
	{
		sqlDB, mock := orm.MockSqlDB(r)
		db := orm.DbInstConfig{
			SqlDB: sqlDB,
			ColumnPolicyConfigs: orm.ColumnPolicyConfigs{
				orm.NewColumnPolicyConfig("deleted").OnDeleteSoftly().PkMode(0),
			},
		}.Build()
		mock.ExpectPrepare("SELECT COUNT(*) FROM user WHERE level = ? AND deleted = ?").ExpectQuery().WithArgs(2, 0).
			WillReturnRows(mock.NewRows([]string{"count"}).AddRow(10))
		_, err := db.Count[orm.User](nil).SqlLogLevel(orm.Level_.Info).Describe("test desc").Condition(orm.Cond().Eq("level", 2)).Do()
		r.NoError(err)

		db = orm.DbInstConfig{
			SqlDB: sqlDB,
			ColumnPolicyConfigs: orm.ColumnPolicyConfigs{
				orm.NewColumnPolicyConfig("deleted").OnDeleteSoftly().PkMode(0),
			},
		}.Build()
		mock.ExpectPrepare("SELECT COUNT(*) FROM user WHERE level = ?").ExpectQuery().WithArgs(2).
			WillReturnRows(mock.NewRows([]string{"count"}).AddRow(10))
		_, err = db.Count[orm.User](nil).SqlLogLevel(orm.Level_.Info).Describe("test desc").Condition(orm.Cond().Eq("level", 2)).IncludeDeleted().Do()
		r.NoError(err)
	}
}
