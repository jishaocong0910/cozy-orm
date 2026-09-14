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
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	orm "github.com/jishaocong0910/cozy-orm"
	"github.com/stretchr/testify/require"
)

func TestTx(t *testing.T) {
	r := require.New(t)
	{
		db, mock := orm.MockDB(r)
		log := orm.MockLogger(db)
		mock.ExpectBegin()
		mock.ExpectPrepare("UPDATE user set status=1 WHERE id=?").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		ctx := context.Background()
		var ctx2 context.Context
		tx := db.Tx(ctx)
		err := tx.TxOptions(nil).Do(func(ctx context.Context) error {
			ctx2 = ctx
			ti := orm.GetTxInfoInner(ctx)
			r.Equal(ti.Creator, tx)
			r.NotNil(ti.SqlTx)
			_, err := db.Mutation(ctx).SqlLogLevel(orm.Level_.Info).BuildSql(func(b *orm.SqlBuilder) {
				b.Write("UPDATE user set status=1 WHERE id=?", 1)
			}).Do()
			return err
		})
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		l := log.Msgs[0]
		r.Equal("SQL: UPDATE user set status=1 WHERE id=?; args: 1(int), tx: true, affected: 1, cost: 0ms", l.Msg)

		ti := orm.GetTxInfoInner(ctx2)
		r.Nil(ti)
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectBegin()
		mock.ExpectPrepare("UPDATE user set status=1 WHERE id=?").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectPrepare("UPDATE user set status=2 WHERE id=?").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			_, err := db.Mutation(ctx).BuildSql(func(b *orm.SqlBuilder) {
				b.Write("UPDATE user set status=1 WHERE id=?", 1)
			}).Do()
			if err != nil {
				return err
			}

			err = db.Tx(ctx).Do(func(ctx context.Context) error {
				_, err := db.Mutation(ctx).BuildSql(func(b *orm.SqlBuilder) {
					b.Write("UPDATE user set status=2 WHERE id=?", 1)
				}).Do()
				return err
			})
			if err != nil {
				return err
			}

			return nil
		})
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectBegin()
		mock.ExpectPrepare("UPDATE user set status=1 WHERE id=?").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectPrepare("UPDATE user set status=2 WHERE id=?").ExpectExec().WillReturnError(errors.New("test nested _rollback"))
		mock.ExpectRollback()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			_, err := db.Mutation(ctx).BuildSql(func(b *orm.SqlBuilder) {
				b.Write("UPDATE user set status=1 WHERE id=?", 1)
			}).Do()
			if err != nil {
				return err
			}

			err = db.Tx(ctx).Do(func(ctx context.Context) error {
				_, err := db.Mutation(ctx).BuildSql(func(b *orm.SqlBuilder) {
					b.Write("UPDATE user set status=2 WHERE id=?", 1)
				}).Do()
				return err
			})
			if err != nil {
				return err
			}

			return nil
		})
		r.EqualError(err, "test nested _rollback")
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		db, mock := orm.MockDB(r)
		db2, mock2 := orm.MockDB(r)
		mock.ExpectBegin()
		mock.ExpectPrepare("UPDATE user set status=1 WHERE id=?").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		mock2.ExpectBegin()
		mock2.ExpectPrepare("UPDATE user set status=2 WHERE id=?").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock2.ExpectCommit()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			_, err := db.Mutation(ctx).BuildSql(func(b *orm.SqlBuilder) {
				b.Write("UPDATE user set status=1 WHERE id=?", 1)
			}).Do()
			if err != nil {
				return err
			}

			err = db2.Tx(ctx).Do(func(ctx context.Context) error {
				_, err := db2.Mutation(ctx).BuildSql(func(b *orm.SqlBuilder) {
					b.Write("UPDATE user set status=2 WHERE id=?", 1)
				}).Do()
				return err
			})
			if err != nil {
				return err
			}

			return nil
		})
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
	}
	{
		db := orm.DbInstConfig{}.Build()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			_, err := db.Mutation(ctx).BuildSql(func(b *orm.SqlBuilder) {}).Do()
			return err
		})
		r.EqualError(err, "no available *sql.DB")
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectBegin()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			return errors.New("error")
		})
		r.EqualError(err, "error")
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectBegin().WillReturnError(errors.New("begin error"))
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			return nil
		})
		r.EqualError(err, "begin error")
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectBegin()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			panic(errors.New("panic error"))
		})
		r.ErrorContains(err, "panic error")
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectBegin()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			panic(123)
		})
		r.ErrorContains(err, "123")
	}
	{
		db, mock := orm.MockDB(r)
		mock.ExpectBegin()
		r.Panics(func() {
			db.Tx(nil).Must().Do(func(ctx context.Context) error {
				panic(errors.New("test panic"))
			})
		})
	}

}

func TestTxHook(t *testing.T) {
	r := require.New(t)
	{
		b := orm.TxHook().Bind(context.Background())
		r.False(b)
		b = orm.TxHook().Bind(nil)
		r.False(b)
	}
	{
		var wg sync.WaitGroup
		db, mock := orm.MockDB(r)
		mock.ExpectBegin()
		mock.ExpectPrepare("UPDATE user SET email = ? WHERE id = ?").ExpectExec().WithArgs("a1", 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectRollback()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			_, err := db.Update[orm.User](ctx).Set("email", "a1").Condition(orm.Cond().Eq("id", 1)).Do()
			if err != nil {
				return err
			}
			hook := orm.TxHook()

			wg.Add(1)
			b := hook.BeforeSync(func(ctx context.Context) error {
				return errors.New("before hook error")
			}).AfterAsync(func(ctx context.Context, commit bool) {
				r.False(commit)
				wg.Done()
			}).Bind(ctx)
			r.True(b)
			ti := orm.GetTxInfoInner(ctx)
			r.Len(ti.TxHook, 1)
			r.Equal(ti.TxHook[0], hook)
			return nil
		})
		fmt.Println(err)
		r.EqualError(err, "before hook error")
		wg.Wait()
	}
	{
		var wg sync.WaitGroup
		db, mock := orm.MockDB(r)
		mock.ExpectBegin()
		mock.ExpectPrepare("UPDATE user SET email = ? WHERE id = ?").ExpectExec().WithArgs("a1", 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectPrepare("UPDATE user SET email = ? WHERE id = ?").ExpectExec().WithArgs("a2", 2).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		ctx := context.WithValue(context.Background(), "key", "value")
		err := db.Tx(ctx).Do(func(ctx context.Context) error {
			_, err := db.Update[orm.User](ctx).Set("email", "a1").Condition(orm.Cond().Eq("id", 1)).Do()
			if err != nil {
				return err
			}

			wg.Add(1)
			b := orm.TxHook().BeforeSync(func(ctx context.Context) error {
				_, err := db.Update[orm.User](ctx).Set("email", "a2").Condition(orm.Cond().Eq("id", 2)).Do()
				return err
			}).AfterAsync(func(ctx context.Context, commit bool) {
				ti := orm.GetTxInfoInner(ctx)
				val := ctx.Value("key").(string)
				r.True(commit)
				r.Nil(ti)
				r.Equal(val, "value")
				wg.Done()
			}, "key").Bind(ctx)
			r.True(b)
			return nil
		})
		r.NoError(err)
		wg.Wait()
	}
	{
		db, mock := orm.MockDB(r)
		log := orm.MockLogger(db)
		mock.ExpectBegin()
		mock.ExpectCommit()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			b := orm.TxHook().AfterAsync(func(ctx context.Context, commit bool) {
				panic("after-hook panic")
			}).Bind(ctx)
			r.True(b)
			return nil
		})
		r.NoError(err)
		time.Sleep(time.Second * 1)
		r.Len(log.Msgs, 1)
		r.Equal(orm.Level_.Error, log.Msgs[0].Level)
		r.Contains(log.Msgs[0].Msg, "panic in transaction after-hook after-hook panic")
	}
}
