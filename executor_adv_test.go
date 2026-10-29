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
	"context"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func TestFind(t *testing.T) {
	r := require.New(t)
	{
		db, _ := MockDB(r, nil)
		r.PanicsWithError("not a valid entity type", func() {
			db.Find[int](nil).Must().Do()
		})
	}
	{
		db, _ := MockDB(r, nil)
		_, err := db.FindOne[User](nil).Cond(func(c *Cond) {
			c.Error(errors.New("cond error"))
		}).Do()
		r.Error(err, "cond error")
	}
	{
		db, _ := MockDB(r, nil)
		_, err := db.FindOne[User](nil).OrderBy(func(c *OrderBy) {
			c.Error(errors.New("order by error"))
		}).Do()
		r.Error(err, "order by error")
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, logger, mock := MockDBAndLogger(r, &Config{IdentifierDelimiter: IdentifierDelimiter_.Backtick, PageMode: PageMode_.LimitOffset})
		mock.ExpectPrepare("SELECT `id`, `name`, `phone`, `email`, `level`, `version` FROM `user` WHERE `level` = ? ORDER BY `id` DESC, `name` ASC LIMIT 1 OFFSET 10 FOR UPDATE").ExpectQuery().WithArgs(1).
			WillReturnRows(mock.NewRows([]string{"id", "name"}).AddRow(9, "abc").AddRow(10, "efg"))
		users, err := db.Find[User](ctx).SqlLogLevel(LogLevel_.Info).Description("test desc").Select("id", "version").OnDemand(DemandFor[UserSimple]()).
			Cond(func(c *Cond) {
				c.Eq("level", 1)
			}).
			OrderBy(func(o *OrderBy) {
				o.Desc("id").Asc("name")
			}).Page(10, 1).LastClause("FOR UPDATE").Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(9), *users[0].Id)
		r.Equal("abc", *users[0].Name)
		r.Equal(int64(10), *users[1].Id)
		r.Equal("efg", *users[1].Name)
		lm := logger.Messages[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(LogLevel_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
	{
		db, mock := MockDB(r, &Config{
			IdentifierDelimiter: IdentifierDelimiter_.Backtick,
			PageMode:            PageMode_.OffsetFetch,
		})
		mock.ExpectPrepare("SELECT `id`, `name` FROM `user` OFFSET 10 ROWS FETCH NEXT 1 ROWS ONLY").
			ExpectQuery().WillReturnRows(mock.NewRows([]string{"id", "name"}).AddRow(9, "abc"))
		_, err := db.Find[User](nil).Select("id", "name").Page(10, 1).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		db, mock := MockDB(r, &Config{
			IdentifierDelimiter: IdentifierDelimiter_.Backtick,
			ColumnPolicyConfigs: ColumnPolicyConfigs{
				NewColumnPolicyConfig("deleted").OnDeleteSoftly().AssignedPkMode(0),
			},
		})
		mock.ExpectPrepare("SELECT `id`, `name` FROM `user` WHERE `level` = ? AND `deleted` = ?").ExpectQuery().WithArgs(1, 0).
			WillReturnRows(mock.NewRows([]string{}))
		_, err := db.Find[User](nil).Select("id", "name").Cond(func(c *Cond) {
			c.Eq("level", 1)
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		mock.ExpectPrepare("SELECT `id`, `name` FROM `user` WHERE `level` = ?").ExpectQuery().WithArgs(1).
			WillReturnRows(mock.NewRows([]string{}))
		_, err = db.Find[User](nil).Select("id", "name").Cond(func(c *Cond) {
			c.Eq("level", 1)
		}).IncludeDeleted().Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
}

func TestFindOne(t *testing.T) {
	r := require.New(t)
	{
		db, _ := MockDB(r, nil)
		r.PanicsWithError("not a valid entity type", func() {
			db.FindOne[int](nil).Must().Do()
		})
	}
	{
		db, mock := MockDB(r, nil)
		mock.ExpectPrepare("SELECT `id`, `name`, `phone`, `email`, `avatar_url`, `status`, `level`, `properties`, `category`, " +
			"`tags`, `attributes`, `uuid`, `create_at`, `update_at`, `version`, `deleted` FROM `user`").ExpectQuery().
			WillReturnRows(mock.NewRows([]string{}).AddRow().AddRow())
		_, err := db.FindOne[User](nil).Do()
		r.Error(err, "return more than one row")
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, logger, mock := MockDBAndLogger(r, &Config{PageMode: PageMode_.LimitOffset})
		mock.ExpectPrepare("SELECT `id`, `name`, `phone`, `email`, `level`, `version` FROM `user` " +
			"WHERE `level` = ? ORDER BY `id` DESC, `name` ASC  LIMIT 1 OFFSET 10 FOR UPDATE").ExpectQuery().WithArgs(1).
			WillReturnRows(mock.NewRows([]string{"id", "name"}).AddRow(9, "abc").AddRow(10, "efg"))
		user, err := db.FindOne[User](ctx).SqlLogLevel(LogLevel_.Info).Description("test desc").Lenient().Select("id", "version").
			OnDemand(DemandFor[UserSimple]()).Cond(func(c *Cond) {
			c.Eq("level", 1)
		}).OrderBy(func(o *OrderBy) {
			o.Desc("id").Asc("name")
		}).Page(10, 1).LastClause("FOR UPDATE").Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(9), *user.Id)
		r.Equal("abc", *user.Name)
		lm := logger.Messages[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(LogLevel_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
	{
		db, mock := MockDB(r, &Config{
			ColumnPolicyConfigs: ColumnPolicyConfigs{
				NewColumnPolicyConfig("deleted").OnDeleteSoftly().AssignedPkMode(0),
			},
		})

		mock.ExpectPrepare("SELECT `id`, `name` FROM `user` WHERE `level` = ? AND `deleted` = ?").ExpectQuery().WithArgs(1, 0).
			WillReturnRows(mock.NewRows([]string{}))
		_, err := db.FindOne[User](nil).Select("id", "name").Cond(func(c *Cond) {
			c.Eq("level", 1)
		}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		mock.ExpectPrepare("SELECT `id`, `name` FROM `user` WHERE `deleted` = ?").ExpectQuery().
			WillReturnRows(mock.NewRows([]string{}))
		_, err = db.FindOne[User](nil).Select("id", "name").Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())

		mock.ExpectPrepare("SELECT `id`, `name` FROM `user` WHERE `level` = ?").ExpectQuery().WithArgs(1).
			WillReturnRows(mock.NewRows([]string{}))
		_, err = db.FindOne[User](nil).Select("id", "name").Cond(func(c *Cond) {
			c.Eq("level", 1)
		}).IncludeDeleted().Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
}

func TestInsert_doDefault(t *testing.T) {
	r := require.New(t)
	{
		db, _ := MockDB(r, nil)
		r.PanicsWithError("not a valid entity type", func() {
			db.Insert[int](nil).Must().Entities(new(1)).Do()
		})
	}
	{
		db, _ := MockDB(r, nil)
		affected, err := db.Insert[User](nil).Do()
		r.Equal(int64(0), affected)
		r.NoError(err)
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, logger, mock := MockDBAndLogger(r, nil)
		mock.ExpectPrepare("INSERT INTO `user`(`name`, `status`) VALUES(?, NULL)").ExpectExec().WithArgs("abc").WillReturnResult(sqlmock.NewResult(1, 1))
		affected, err := db.Insert[User](ctx).SqlLogLevel(LogLevel_.Info).Description("test desc").Entities(&User{Name: new("abc")}).Required("status").Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
		lm := logger.Messages[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(LogLevel_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
	{
		ctx := context.WithValue(context.Background(), "status", 4)
		db, logger, mock := MockDBAndLogger(r, &Config{
			GetGeneratedKeyMode: GetGeneratedKeyMode_.FirstInsertId,
			ColumnPolicyConfigs: ColumnPolicyConfigs{
				NewColumnPolicyConfig("create_at").UseCreateTime(),
				NewColumnPolicyConfig("update_at").UseUpdateTime(),
				NewColumnPolicyConfig("version").UseRowVersion(),
				NewColumnPolicyConfig("status").OnInsert().Value(false, true, func(ctx context.Context) any {
					if s, ok := ctx.Value("status").(int); ok {
						return s
					}
					return 1
				}),
			},
		})
		mock.ExpectPrepare("INSERT INTO `user`(`name`, `status`, `create_at`, `update_at`, `version`) VALUES(?, ?, ?, ?, ?), (?, ?, ?, ?, ?)").ExpectExec().
			WillReturnResult(sqlmock.NewResult(1, 2)).
			WithArgs("name1", 2, AnyTime{}, AnyTime{}, 1, "name2", 4, AnyTime{}, AnyTime{}, 1)

		_, err := db.Insert[User](ctx).SqlLogLevel(LogLevel_.Info).Description("test desc").
			Entities(&User{Name: new("name1"), Status: new(UserStatus(2))}, &User{Name: new("name2")}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		lm := logger.Messages[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(LogLevel_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
}

func TestInsert_doGetIdentity(t *testing.T) {
	r := require.New(t)
	{
		u1 := &User{Name: new("name1")}
		u2 := &User{Name: new("name2")}
		db, mock := MockDB(r, &Config{GetGeneratedKeyMode: GetGeneratedKeyMode_.FirstInsertId})
		mock.ExpectPrepare("INSERT INTO `user`(`name`) VALUES(?), (?)").ExpectExec().WillReturnResult(sqlmock.NewResult(3, 2)).
			WithArgs("name1", "name2")
		_, err := db.Insert[User](nil).Entities(u1, u2).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(3), *u1.Id)
		r.Equal(int64(4), *u2.Id)
	}
	{
		u1 := &User{Name: new("name1")}
		u2 := &User{Name: new("name2")}
		db, mock := MockDB(r, &Config{GetGeneratedKeyMode: GetGeneratedKeyMode_.LastInsertId})
		mock.ExpectPrepare("INSERT INTO `user`(`name`) VALUES(?), (?)").ExpectExec().WillReturnResult(sqlmock.NewResult(3, 2)).
			WithArgs("name1", "name2")
		_, err := db.Insert[User](nil).Entities(u1, u2).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(2), *u1.Id)
		r.Equal(int64(3), *u2.Id)
	}
}

func TestInsert_doInsertReturning(t *testing.T) {
	r := require.New(t)
	{
		u1 := &User{Name: new("name1")}
		u2 := &User{Name: new("name2")}
		db, mock := MockDB(r, &Config{DBType: DBType_.PostgreSQL})
		mock.ExpectPrepare("INSERT INTO \"user\"(\"name\") VALUES($1), ($2) RETURNING \"id\"").ExpectQuery().
			WithArgs("name1", "name2").WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
		_, err := db.Insert[User](nil).Entities(u1, u2).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), *u1.Id)
		r.Equal(int64(2), *u2.Id)
	}
}

func TestInsert_doSQLServer(t *testing.T) {
	r := require.New(t)
	{
		u1 := &User{Name: new("name1")}
		u2 := &User{Name: new("name2")}
		db, mock := MockDB(r, &Config{DBType: DBType_.SQLServer})
		mock.ExpectPrepare("INSERT INTO [user]([name]) OUTPUT INSERTED.[id] VALUES(:1), (:2)").ExpectQuery().
			WithArgs("name1", "name2").WillReturnRows(mock.NewRows([]string{"id"}).AddRow(1).AddRow(2))
		_, err := db.Insert[User](nil).Entities(u1, u2).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), *u1.Id)
		r.Equal(int64(2), *u2.Id)
	}
}

func TestInsert_doOracle(t *testing.T) {
	r := require.New(t)
	{
		db := Config{
			DBType: DBType_.Oracle,
		}.Build()
		users := make([]*User, 0, 1001)
		for range 1001 {
			users = append(users, &User{})
		}
		_, err := db.Insert[User](nil).Entities(users...).Do()
		r.EqualError(err, "entity count cannot exceed 1000 in Oracle")
	}
	{
		db := Config{
			DBType: DBType_.Oracle,
		}.Build()
		u := &DemoMulPk{Name: new("name")}
		db, mock := MockDB(r, &Config{DBType: DBType_.Oracle})
		mock.ExpectPrepare("INSERT INTO \"DEMO_MUL_PK\"(\"name\") VALUES(:1)").ExpectExec().
			WithArgs("name").WillReturnResult(sqlmock.NewResult(0, 1))

		_, err := db.Insert[DemoMulPk](nil).Entities(u).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		db := Config{
			DBType: DBType_.Oracle,
		}.Build()
		u := &User{Name: new("name")}
		outArg := OutArg{MockValue: int64(20)}
		db, mock := MockDB(r, &Config{DBType: DBType_.Oracle})
		mock.ExpectPrepare("INSERT INTO \"USER\"(\"name\") VALUES(:1) RETURNING \"id\" INTO :out_id").ExpectExec().
			WithArgs("name", outArg).WillReturnResult(sqlmock.NewResult(0, 1))

		_, err := db.Insert[User](nil).Entities(u).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(20), *u.Id)
	}
	{
		db := Config{
			DBType: DBType_.Oracle,
		}.Build()
		u1 := &User{Name: new("name1")}
		u2 := &User{Name: new("name2")}
		outArg := OutArg{MockValue: ",1,2"}
		db, mock := MockDB(r, &Config{DBType: DBType_.Oracle})
		mock.ExpectPrepare("DECLARE l_id NUMBER; l_results VARCHAR2(32767) := '';"+
			" BEGIN"+
			" INSERT INTO \"USER\"(\"name\") VALUES(:1) RETURNING \"id\" INTO l_id; l_results := l_results || ',' || TO_CHAR(l_id);"+
			" INSERT INTO \"USER\"(\"name\") VALUES(:2) RETURNING \"id\" INTO l_id; l_results := l_results || ',' || TO_CHAR(l_id);"+
			" :out_results := l_results; EXCEPTION WHEN OTHERS THEN ROLLBACK; RAISE; END;").ExpectExec().
			WithArgs("name1", "name2", outArg).WillReturnResult(sqlmock.NewResult(0, 1))

		_, err := db.Insert[User](nil).Entities(u1, u2).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), *u1.Id)
		r.Equal(int64(2), *u2.Id)
	}
}

func TestUpdate(t *testing.T) {
	r := require.New(t)
	{
		db, _ := MockDB(r, nil)
		r.PanicsWithError("not a valid entity type", func() {
			db.Update[int](nil).Must().Entity(new(1)).Do()
		})
	}
	{
		db, _ := MockDB(r, nil)
		_, err := db.Update[User](nil).Entity(&User{Phone: new("123")}).Cond(func(c *Cond) {
			c.Error(errors.New("cond error"))
		}).Do()
		r.EqualError(err, "cond error")
	}
	{
		db, _ := MockDB(r, nil)
		_, err := db.Update[User](nil).Entity(&User{Phone: new("123")}).Assignment(func(a *Assignment) {
			a.Error(errors.New("assignment error"))
		}).Do()
		r.EqualError(err, "assignment error")
	}
	{
		db, mock := MockDB(r, nil)
		affected, err := db.Update[User](nil).Entity(&User{Phone: new("123")}).Do()
		r.Equal(int64(0), affected)
		r.EqualError(err, "full table modification blocked")

		mock.ExpectPrepare("UPDATE `user` SET `phone` = ?").ExpectExec().WithArgs("123").
			WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err = db.Update[User](nil).Entity(&User{Phone: new("123")}).SkipSafety().Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		db, _ := MockDB(r, nil)
		affected, err := db.Update[User](nil).Do()
		r.Equal(int64(0), affected)
		r.NoError(err)
	}
	{
		db, _ := MockDB(r, nil)
		affected, err := db.Update[User](nil).Entity(&User{}).Do()
		r.Equal(int64(0), affected)
		r.NoError(err)
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, logger, mock := MockDBAndLogger(r, nil)
		mock.ExpectPrepare("UPDATE `user` SET `name` = ? WHERE `id` = ?").
			ExpectExec().WithArgs("abc", 1).WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := db.Update[User](ctx).SqlLogLevel(LogLevel_.Info).Description("test desc").
			Entity(&User{Id: new(int64(1)), Name: new("abc")}).Do()
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		r.Equal(int64(1), affected)
		lm := logger.Messages[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(LogLevel_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, mock := MockDB(r, nil)
		mock.ExpectPrepare("UPDATE `user` SET `name` = ?, `phone` = NULL, `email` = ?, `status` = ?, `level` = level + ?, `properties` = NULL, `tags` = NULL WHERE `id` = ? AND `status` = ?").
			ExpectExec().WithArgs("abc", "bbb@example.com", 2, 3, 1, 4).WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := db.Update[User](ctx).Entity(&User{Id: new(int64(1)), Name: new("abc"), Email: new("abc@example.com"), Category: &UserCategory{Organization: "none", Class: 1}}).
			OnDemand(DemandFor[UserSimple]()).Required("properties").Assignment(func(a *Assignment) {
			a.SetValue("status", UserStatus(2)).SetValue("tags", "'t1,t2'").
				SetValue("email", "bbb@example.com").
				SetValue("tags", nil).
				SetExpr("level", func(e *Expr) {
					e.Text("level + ").Arg(3)
				})
		}).Cond(func(c *Cond) {
			c.Eq("status", 4)
		}).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
	}
	{
		db, mock := MockDB(r, &Config{
			ColumnPolicyConfigs: ColumnPolicyConfigs{
				NewColumnPolicyConfig("create_at").UseCreateTime(),
				NewColumnPolicyConfig("update_at").UseUpdateTime(),
				NewColumnPolicyConfig("version").UseRowVersion(),
				NewColumnPolicyConfig("deleted").OnDeleteSoftly().AssignedPkMode(0),
			},
		})

		mock.ExpectPrepare("UPDATE `user` SET `name` = ?, `update_at` = ?, `version` = version + 1 WHERE `id` = ? AND `deleted` = ?").ExpectExec().
			WithArgs("abc", AnyTime{}, 1, 0).
			WillReturnResult(sqlmock.NewResult(0, 0))
		_, err := db.Update[User](nil).Entity(&User{Name: new("abc")}).Cond(func(c *Cond) {
			c.Eq("id", 1)
		}).Do()
		r.NoError(err)

		mock.ExpectPrepare("UPDATE `user` SET `name` = ?, `update_at` = ?, `version` = version + 1 WHERE `id` = ?").ExpectExec().
			WithArgs("abc", AnyTime{}, 1).
			WillReturnResult(sqlmock.NewResult(0, 0))
		_, err = db.Update[User](nil).Entity(&User{Id: new(int64(1)), Name: new("abc"), CreateAt: new(time.Now()), Version: new(int64(5))}).
			Required("update_at").IncludeDeleted().Do()
		r.NoError(err)
	}
}

func TestUpdateBatch(t *testing.T) {
	r := require.New(t)
	{
		db, _ := MockDB(r, nil)
		r.PanicsWithError("not a valid entity type", func() {
			db.UpdateBatch[int](nil).Must().Entities(new(1)).Do()
		})
	}
	{
		db, _ := MockDB(r, nil)
		_, err := db.UpdateBatch[User](nil).Entities(&User{Phone: new("123")}).Cond(func(c *Cond) {
			c.Error(errors.New("cond error"))
		}).Do()
		r.EqualError(err, "cond error")
	}
	{
		db, _ := MockDB(r, nil)
		_, err := db.UpdateBatch[User](nil).Entities(&User{Phone: new("123")}).Assignment(func(a *Assignment) {
			a.Error(errors.New("assignment error"))
		}).Do()
		r.EqualError(err, "assignment error")
	}
	{
		db, _ := MockDB(r, nil)
		affected, err := db.UpdateBatch[DemoMulPk](nil).Entities(&DemoMulPk{}, &DemoMulPk{}).Do()
		r.Equal(int64(0), affected)
		r.EqualError(err, "\"orm.DemoMulPk\" must have exactly one field with the \"pk\" tag")
	}
	{
		db, mock := MockDB(r, nil)
		mock.ExpectPrepare("")
		affected, err := db.UpdateBatch[User](nil).Entities(&User{Name: new("abc")}).Do()
		r.Equal(int64(0), affected)
		r.EqualError(err, "field 'Id' is nil at index 0 of entities")
	}
	{
		db, _ := MockDB(r, nil)
		affected, err := db.UpdateBatch[User](nil).Do()
		r.Equal(int64(0), affected)
		r.NoError(err)
	}
	{
		db, _ := MockDB(r, nil)
		affected, err := db.UpdateBatch[User](nil).Entities(&User{Id: new(int64(1))}, &User{Id: new(int64(2))}).Do()
		r.Equal(int64(0), affected)
		r.NoError(err)
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, logger, mock := MockDBAndLogger(r, nil)
		mock.ExpectPrepare("UPDATE `user` SET `name` = CASE `id` WHEN ? THEN ? WHEN ? THEN ? END, `phone` = NULL, "+
			"`email` = CASE `id` WHEN ? THEN NULL WHEN ? THEN ? END, `status` = ?, `level` = NULL, `properties` = NULL, `tags` = NULL "+
			"WHERE `id` IN(?, ?) AND `level` = ?").ExpectExec().
			WithArgs(1, "a", 2, "b", 1, 2, "email", 2, 1, 2, 5).WillReturnResult(sqlmock.NewResult(1, 1))
		affected, err := db.UpdateBatch[User](ctx).SqlLogLevel(LogLevel_.Info).Description("test desc").OnDemand(DemandFor[UserSimple]()).
			Entities(&User{Id: new(int64(1)), Name: new("a")}, &User{Id: new(int64(2)), Name: new("b"), Email: new("email")}).
			Required("properties").Assignment(func(a *Assignment) {
			a.SetValue("status", UserStatus(2)).SetValue("tags", nil)
		}).Cond(func(c *Cond) {
			c.Eq("level", 5)
		}).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
		lm := logger.Messages[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(LogLevel_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
	{
		db, mock := MockDB(r, &Config{
			ColumnPolicyConfigs: ColumnPolicyConfigs{
				NewColumnPolicyConfig("create_at").UseCreateTime(),
				NewColumnPolicyConfig("update_at").UseUpdateTime(),
				NewColumnPolicyConfig("version").UseRowVersion(),
				NewColumnPolicyConfig("deleted").OnDeleteSoftly().AssignedPkMode(0),
			},
		})
		mock.ExpectPrepare("UPDATE `user` SET `name` = CASE `id` WHEN ? THEN ? WHEN ? THEN ? END, `update_at` = ?, `version` = version + 1 WHERE `id` IN(?, ?) AND `level` = ? AND `deleted` = ?").ExpectExec().
			WithArgs(1, "a", 2, "b", AnyTime{}, 1, 2, 5, 0).
			WillReturnResult(sqlmock.NewResult(1, 1))
		affected, err := db.UpdateBatch[User](nil).Entities(&User{Id: new(int64(1)), Name: new("a")}, &User{Id: new(int64(2)), Name: new("b")}).
			Cond(func(c *Cond) {
				c.Eq("level", 5)
			}).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)

		mock.ExpectPrepare("UPDATE `user` SET `name` = CASE `id` WHEN ? THEN ? WHEN ? THEN ? END, `update_at` = ?, `version` = version + 1 WHERE `id` IN(?, ?)").ExpectExec().
			WithArgs(1, "a", 2, "b", AnyTime{}, 1, 2).
			WillReturnResult(sqlmock.NewResult(1, 1))
		affected, err = db.UpdateBatch[User](nil).Entities(&User{Id: new(int64(1)), Name: new("a")}, &User{Id: new(int64(2)), Name: new("b"), Version: new(int64(5))}).
			Required("update_at", "version").IncludeDeleted().Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, mock := MockDB(r, nil)
		mock.ExpectPrepare("UPDATE `user` SET `name` = ?, `phone` = NULL, `email` = NULL, `status` = ?, `level` = NULL, `properties` = NULL, `tags` = NULL "+
			"WHERE `id` = ? AND `level` = ?").ExpectExec().
			WithArgs("a", 2, 1, 5).WillReturnResult(sqlmock.NewResult(1, 1))
		affected, err := db.UpdateBatch[User](ctx).SqlLogLevel(LogLevel_.Info).Description("test desc").OnDemand(DemandFor[UserSimple]()).
			Entities(&User{Id: new(int64(1)), Name: new("a")}).Required("properties").Assignment(func(a *Assignment) {
			a.SetValue("status", UserStatus(2)).SetValue("tags", nil)
		}).Cond(func(c *Cond) {
			c.Eq("level", 5)
		}).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
	}
}

func TestDelete(t *testing.T) {
	r := require.New(t)
	{
		db, _ := MockDB(r, nil)
		r.PanicsWithError("not a valid entity type", func() {
			db.Delete[int](nil).Must().Do()
		})
	}
	{
		db, _ := MockDB(r, nil)
		_, err := db.Delete[User](nil).Cond(func(c *Cond) {
			c.Error(errors.New("cond error"))
		}).Do()
		r.EqualError(err, "cond error")
	}
	{
		db, mock := MockDB(r, nil)
		affected, err := db.Delete[User](nil).Do()
		r.Equal(int64(0), affected)
		r.EqualError(err, "full table modification blocked")

		mock.ExpectPrepare("DELETE FROM `user`").ExpectExec().WithoutArgs().WillReturnResult(sqlmock.NewResult(0, 2))
		affected, err = db.Delete[User](nil).SkipSafety().Do()
		r.Equal(int64(2), affected)
		r.NoError(err)
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, logger, mock := MockDBAndLogger(r, nil)
		mock.ExpectPrepare("DELETE FROM `user` WHERE `id` = ?").ExpectExec().WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := db.Delete[User](ctx).SqlLogLevel(LogLevel_.Info).Description("test desc").Cond(func(c *Cond) {
			c.Eq("id", 1)
		}).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
		lm := logger.Messages[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(LogLevel_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
}

func TestDeleteSoftly(t *testing.T) {
	r := require.New(t)
	{
		db, _ := MockDB(r, nil)
		r.PanicsWithError("not a valid entity type", func() {
			db.DeleteSoftly[int](nil).Must().Do()
		})
	}
	{
		db, _ := MockDB(r, nil)
		_, err := db.DeleteSoftly[User](nil).Do()
		r.EqualError(err, "delete softly mode is undefined")
	}
	{
		db, _ := MockDB(r, nil)
		_, err := db.DeleteSoftly[User](nil).Cond(func(c *Cond) {
			c.Error(errors.New("cond error"))
		}).Do()
		r.EqualError(err, "cond error")
	}
	{
		db, _ := MockDB(r, nil)
		_, err := db.DeleteSoftly[User](nil).Assignment(func(a *Assignment) {
			a.Error(errors.New("assignment error"))
		}).Do()
		r.EqualError(err, "assignment error")
	}
	{
		db, mock := MockDB(r, &Config{
			ColumnPolicyConfigs: ColumnPolicyConfigs{
				NewColumnPolicyConfig("deleted").OnDeleteSoftly().AssignedPkMode(0),
			},
		})
		affected, err := db.DeleteSoftly[User](nil).Do()
		r.EqualError(err, "full table modification blocked")
		r.Equal(int64(0), affected)

		mock.ExpectPrepare("UPDATE `user` SET `deleted` = `id` WHERE `deleted` = ?").ExpectExec().WithArgs(0).
			WillReturnResult(sqlmock.NewResult(0, 2))
		affected, err = db.DeleteSoftly[User](nil).SkipSafety().Do()
		r.NoError(err)
		r.Equal(int64(2), affected)
	}
	{
		db, _ := MockDB(r, &Config{
			ColumnPolicyConfigs: ColumnPolicyConfigs{
				NewColumnPolicyConfig("deleted").OnDeleteSoftly().AssignedPkMode(0),
			},
		})
		_, err := db.DeleteSoftly[int](nil).Do()
		r.EqualError(err, "not a valid entity type")
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		log := &MockLogger{}
		db, mock := MockDB(r, &Config{
			ColumnPolicyConfigs: ColumnPolicyConfigs{
				NewColumnPolicyConfig("deleted").OnDeleteSoftly().AssignedPkMode(0),
			},
			Logger: log,
		})
		mock.ExpectPrepare("UPDATE `user` SET `deleted` = `id`, `status` = ?, `name` = 'anon' WHERE `id` = ? AND `deleted` = ?").ExpectExec().
			WithArgs(4, 1, 0).WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := db.DeleteSoftly[User](ctx).SqlLogLevel(LogLevel_.Info).Description("test desc").Assignment(func(a *Assignment) {
			a.SetValue("status", 4).SetExpr("name", func(e *Expr) {
				e.Text("'anon'")
			})
		}).Cond(func(c *Cond) {
			c.Eq("id", 1)
		}).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
		lm := log.Messages[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(LogLevel_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
	{
		db, mock := MockDB(r, &Config{
			ColumnPolicyConfigs: ColumnPolicyConfigs{
				NewColumnPolicyConfig("deleted").OnDeleteSoftly().AssignedPkMode(0),
			},
		})
		mock.ExpectPrepare("UPDATE `user` SET `deleted` = `id` WHERE `id` = ? AND `deleted` = ?").ExpectExec().WithArgs(1, 0).
			WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := db.DeleteSoftly[User](nil).Cond(func(c *Cond) {
			c.Eq("id", 1)
		}).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
	}
	{
		db, mock := MockDB(r, &Config{
			ColumnPolicyConfigs: ColumnPolicyConfigs{
				NewColumnPolicyConfig("deleted").OnDeleteSoftly().AssignedNullMode(0),
			},
		})
		mock.ExpectPrepare("UPDATE `user` SET `deleted` = NULL WHERE `id` = ? AND `deleted` = ?").ExpectExec().WithArgs(1, 0).
			WillReturnResult(sqlmock.NewResult(0, 1))
		affected, err := db.DeleteSoftly[User](nil).Cond(func(c *Cond) {
			c.Eq("id", 1)
		}).Do()
		r.NoError(err)
		r.Equal(int64(1), affected)
	}

}

func TestCount(t *testing.T) {
	r := require.New(t)
	{
		db, _ := MockDB(r, nil)
		r.PanicsWithError("not a valid entity type", func() {
			db.Count[int](nil).Must().Do()
		})
	}
	{
		db, _ := MockDB(r, nil)
		_, err := db.Count[User](nil).Cond(func(c *Cond) {
			c.Error(errors.New("cond error"))
		}).Do()
		r.EqualError(err, "cond error")
	}
	{
		ctx := context.WithValue(context.Background(), "test", "test")
		db, logger, mock := MockDBAndLogger(r, nil)
		mock.ExpectPrepare("SELECT COUNT(*) FROM `user` WHERE `level` = ?").ExpectQuery().WithArgs(2).WillReturnRows(mock.NewRows([]string{"count"}).AddRow(10))
		count, err := db.Count[User](ctx).SqlLogLevel(LogLevel_.Info).Description("test desc").Cond(func(c *Cond) {
			c.Eq("level", 2)
		}).Do()
		r.NoError(err)
		r.Equal(int64(10), count)
		lm := logger.Messages[0]
		r.Equal(ctx, lm.Ctx)
		r.Equal(LogLevel_.Info, lm.Level)
		r.Contains(lm.Msg, "desc: test desc")
	}
	{
		db, mock := MockDB(r, &Config{
			ColumnPolicyConfigs: ColumnPolicyConfigs{
				NewColumnPolicyConfig("deleted").OnDeleteSoftly().AssignedPkMode(0),
			},
		})
		mock.ExpectPrepare("SELECT COUNT(*) FROM `user` WHERE `level` = ? AND `deleted` = ?").ExpectQuery().WithArgs(2, 0).
			WillReturnRows(mock.NewRows([]string{"count"}).AddRow(10))
		_, err := db.Count[User](nil).SqlLogLevel(LogLevel_.Info).Description("test desc").Cond(func(c *Cond) {
			c.Eq("level", 2)
		}).Do()
		r.NoError(err)

		mock.ExpectPrepare("SELECT COUNT(*) FROM `user` WHERE `level` = ?").ExpectQuery().WithArgs(2).
			WillReturnRows(mock.NewRows([]string{"count"}).AddRow(10))
		_, err = db.Count[User](nil).SqlLogLevel(LogLevel_.Info).Description("test desc").Cond(func(c *Cond) {
			c.Eq("level", 2)
		}).IncludeDeleted().Do()
		r.NoError(err)
	}
}

func TestCond(t *testing.T) {
	r := require.New(t)
	{
		a := &argFetcher{}
		c := (&Cond{}).Eq("Eq", a.next("Eq")).Ne("Ne", a.next("Ne")).
			Gt("Gt", a.next("Gt")).Lt("Lt", a.next("Lt")).
			Ge("Ge", a.next("Ge")).Le("Le", a.next("Le")).
			Like("Like", a.next2("Like", func(next string) any { return "%" + next + "%" })).
			LikeLeft("LikeLeft", a.next2("LikeLeft", func(next string) any { return next + "%" })).
			LikeRight("LikeRight", a.next2("LikeRight", func(next string) any { return "%" + next })).
			LikePattern("LikePattern", a.next("LikePattern")).
			In("In", []string{a.next("In"), a.next("In")}).
			Between("Between", a.next("Between"), a.next("Between")).
			IsNull("IsNull").IsNotNull("IsNull").
			Not().Eq("Not", a.next("Not")).
			Or().Eq("OrEq1", a.next("OrEq1")).Eq("OrEq2", a.next("OrEq2"))

		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.accept(c)
		r.Equal("Eq = ? AND Ne <> ? AND Gt > ? AND Lt < ? AND Ge >= ? AND Le <= ? AND Like LIKE ? AND LikeLeft LIKE ?"+
			" AND LikeRight LIKE ? AND LikePattern LIKE ? AND In IN(?, ?) AND Between BETWEEN ? AND ? AND IsNull IS NULL"+
			" AND IsNull IS NOT NULL AND NOT Not = ? OR OrEq1 = ? AND OrEq2 = ?", b.b.String())
		r.Equal(a.args, b.args)
	}
}

func TestCond_Sub(t *testing.T) {
	r := require.New(t)
	{
		a := &argFetcher{}
		c := (&Cond{}).Sub(nil).Sub(func(c *Cond) {
			c.Eq("c1", a.next("c1")).Or().Eq("c2", a.next("c2"))
		})
		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.accept(c)
		r.Equal("c1 = ? OR c2 = ?", b.b.String())
		r.Equal([]any{"c1_1", "c2_2"}, b.args)
	}
	{
		a := &argFetcher{}
		c := (&Cond{}).Sub(func(c *Cond) {
			c.Eq("c1", a.next("c1")).Or().Eq("c2", a.next("c2"))
		}).Eq("c3", a.next("c3"))
		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.accept(c)
		r.Equal("(c1 = ? OR c2 = ?) AND c3 = ?", b.b.String())
		r.Equal([]any{"c1_1", "c2_2", "c3_3"}, b.args)
	}
	{
		a := &argFetcher{}
		c := (&Cond{}).Eq("c1", a.next("c1")).Sub(func(c *Cond) {
			c.Eq("c2", a.next("c2")).Or().Eq("c3", a.next("c3"))
		})
		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.accept(c)
		r.Equal("c1 = ? AND (c2 = ? OR c3 = ?)", b.b.String())
		r.Equal([]any{"c1_1", "c2_2", "c3_3"}, b.args)
	}
	{
		a := &argFetcher{}
		c := (&Cond{}).Not().Sub(func(c *Cond) {
			c.Eq("c1", a.next("c1")).Eq("c2", a.next("c2"))
		}).Not().Sub(func(c *Cond) {
			c.Eq("c3", a.next("c3")).Eq("c4", a.next("c4"))
		})
		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.accept(c)
		r.Equal("NOT (c1 = ? AND c2 = ?) AND NOT (c3 = ? AND c4 = ?)", b.b.String())
		r.Equal([]any{"c1_1", "c2_2", "c3_3", "c4_4"}, b.args)
	}
	{
		a := &argFetcher{}
		c := (&Cond{}).Eq("c1", a.next("c1")).Sub(func(c *Cond) {
			c.Error(errors.New("expr error"))
		})
		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.accept(c)
		r.Equal("c1 = ?", b.b.String())
		r.Equal(a.args, b.args)
		r.EqualError(c.err, "expr error")
	}
}

func TestCond_Expr(t *testing.T) {
	r := require.New(t)
	{
		a := &argFetcher{}
		c := (&Cond{}).Expr(func(c *CondExpr) {
			c.Text("c1 = 'c1'").Text(" AND c2 = ").Arg(a.next("c2"))
		})

		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.accept(c)
		r.Equal("c1 = 'c1' AND c2 = ?", b.b.String())
		r.Equal(a.args, b.args)
	}
	{
		a := &argFetcher{}
		c := (&Cond{}).Expr(func(c *CondExpr) {
			c.Text("c1 = 'c1'").Text(" AND c2 = ").Arg(a.next("c2"))
		}).Eq("c3", a.next("c3"))

		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.accept(c)
		r.Equal("(c1 = 'c1' AND c2 = ?) AND c3 = ?", b.b.String())
		r.Equal(a.args, b.args)
	}
	{
		a := &argFetcher{}
		c := (&Cond{}).Sub(func(c *Cond) {

		}).Expr(func(c *CondExpr) {
			c.Text("c1 = 'c1'").Text(" AND c2 = ").Arg(a.next("c2"))
		})

		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.accept(c)
		r.Equal("c1 = 'c1' AND c2 = ?", b.b.String())
		r.Equal(a.args, b.args)
	}
	{
		a := &argFetcher{}
		c := (&Cond{}).Sub(func(c *Cond) {
			c.Eq("c1", a.next("c1"))
		}).Expr(func(c *CondExpr) {
			c.Text("c2 = 'c2'").Text(" AND c3 = ").Arg(a.next("c3"))
		})

		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.accept(c)
		r.Equal("c1 = ? AND (c2 = 'c2' AND c3 = ?)", b.b.String())
		r.Equal(a.args, b.args)
	}
	{
		a := &argFetcher{}
		c := (&Cond{}).Eq("c1", a.next("c1")).Expr(func(c *CondExpr) {
			c.Error(errors.New("expr error"))
		})
		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.accept(c)
		r.Equal("c1 = ?", b.b.String())
		r.Equal(a.args, b.args)
		r.EqualError(c.err, "expr error")
	}
}

type argFetcher struct {
	args []any
	i    int
}

func (a *argFetcher) next(prefix string) string {
	a.i++
	arg := prefix + "_" + strconv.Itoa(a.i)
	a.args = append(a.args, arg)
	return arg
}

func (a *argFetcher) next2(prefix string, actualHandler func(next string) any) string {
	a.i++
	arg := prefix + "_" + strconv.Itoa(a.i)
	if actualHandler != nil {
		a.args = append(a.args, actualHandler(arg))
	} else {
		a.args = append(a.args, arg)
	}
	return arg
}
