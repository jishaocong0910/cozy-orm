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
		db, logger, mock := orm.MockDBAndLogger(r, nil)
		mock.ExpectBegin()
		mock.ExpectPrepare("UPDATE user set status=1 WHERE id=?").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		ctx := context.Background()
		var ctx2 context.Context
		tx := db.Tx(ctx)
		err := tx.TxOptions(nil).Do(func(ctx context.Context) error {
			ctx2 = ctx
			ti := orm.GetTxInfoInner(ctx, db)
			r.Equal(ti.Owner, tx)
			r.NotNil(ti.RawTx)
			_, err := db.Mutation(ctx).SqlLogLevel(orm.Level_.Info).BuildSql(func(b *orm.SQLBuilder) {
				b.Write("UPDATE user set status=1 WHERE id=?", 1)
			}).Do()
			return err
		})
		r.NoError(err)
		r.NoError(mock.ExpectationsWereMet())
		l := logger.Messages[0]
		r.Equal("SQL: UPDATE user set status=1 WHERE id=?; args: 1(int), tx: true, affected: 1, cost: 0ms", l.Msg)

		ti := orm.GetTxInfoInner(ctx2, db)
		r.Nil(ti)
	}
	{
		db, mock := orm.MockDB(r, nil)
		mock.ExpectBegin()
		mock.ExpectPrepare("UPDATE user set status=1 WHERE id=?").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectPrepare("UPDATE user set status=2 WHERE id=?").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			_, err := db.Mutation(ctx).BuildSql(func(b *orm.SQLBuilder) {
				b.Write("UPDATE user set status=1 WHERE id=?", 1)
			}).Do()
			if err != nil {
				return err
			}

			err = db.Tx(ctx).Do(func(ctx context.Context) error {
				_, err := db.Mutation(ctx).BuildSql(func(b *orm.SQLBuilder) {
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
		db, mock := orm.MockDB(r, nil)
		mock.ExpectBegin()
		mock.ExpectPrepare("UPDATE user set status=1 WHERE id=?").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectPrepare("UPDATE user set status=2 WHERE id=?").ExpectExec().WillReturnError(errors.New("test nested _rollback"))
		mock.ExpectRollback()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			_, err := db.Mutation(ctx).BuildSql(func(b *orm.SQLBuilder) {
				b.Write("UPDATE user set status=1 WHERE id=?", 1)
			}).Do()
			if err != nil {
				return err
			}

			err = db.Tx(ctx).Do(func(ctx context.Context) error {
				_, err := db.Mutation(ctx).BuildSql(func(b *orm.SQLBuilder) {
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
		db, mock := orm.MockDB(r, nil)
		db2, mock2 := orm.MockDB(r, nil)
		mock.ExpectBegin()
		mock.ExpectPrepare("UPDATE user set status = 1 WHERE id = ?").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectPrepare("UPDATE user set status = 2 WHERE id = ?").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
		mock2.ExpectBegin()
		mock2.ExpectPrepare("UPDATE user set status = 3 WHERE id = ?").ExpectExec().WillReturnResult(sqlmock.NewResult(0, 1))
		mock2.ExpectCommit()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			_, err := db.Mutation(ctx).BuildSql(func(b *orm.SQLBuilder) {
				b.Write("UPDATE user set status = 1 WHERE id = ?", 1)
			}).Do()
			if err != nil {
				return err
			}

			err = db2.Tx(ctx).Do(func(ctx context.Context) error {
				err = db.Tx(ctx).Do(func(ctx context.Context) error {
					_, err = db.Mutation(ctx).BuildSql(func(b *orm.SQLBuilder) {
						b.Write("UPDATE user set status = 2 WHERE id = ?", 1)
					}).Do()
					return err
				})
				if err != nil {
					return err
				}

				_, err := db2.Mutation(ctx).BuildSql(func(b *orm.SQLBuilder) {
					b.Write("UPDATE user set status = 3 WHERE id = ?", 1)
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
		db := orm.DBConfig{}.Build()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			_, err := db.Mutation(ctx).BuildSql(func(b *orm.SQLBuilder) {}).Do()
			return err
		})
		r.EqualError(err, "no available *sql.DB")
	}
	{
		db, mock := orm.MockDB(r, nil)
		mock.ExpectBegin()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			return errors.New("error")
		})
		r.EqualError(err, "error")
	}
	{
		db, mock := orm.MockDB(r, nil)
		mock.ExpectBegin().WillReturnError(errors.New("begin error"))
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			return nil
		})
		r.EqualError(err, "begin error")
	}
	{
		db, mock := orm.MockDB(r, nil)
		mock.ExpectBegin()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			panic(errors.New("panic error"))
		})
		r.ErrorContains(err, "panic error")
	}
	{
		db, mock := orm.MockDB(r, nil)
		mock.ExpectBegin()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			panic(123)
		})
		r.ErrorContains(err, "123")
	}
	{
		db, mock := orm.MockDB(r, nil)
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
		db, _ := orm.MockDB(r, nil)
		b := db.TxHook(context.Background())
		r.Nil(b)
		b = db.TxHook(context.Background())
		r.Nil(b)
	}
	{
		var wg sync.WaitGroup
		db, mock := orm.MockDB(r, nil)
		mock.ExpectBegin()
		mock.ExpectPrepare("UPDATE user SET `email` = ? WHERE `id` = ?").ExpectExec().WithArgs("a1", 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectRollback()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			_, err := db.Update[orm.User](ctx).Set("email", "a1").Cond(func(c *orm.Cond) {
				c.Eq("id", 1)
			}).Do()
			if err != nil {
				return err
			}
			hook := db.TxHook(ctx)
			r.NotNil(hook)

			wg.Add(1)
			hook.BeforeSync(func(ctx context.Context) error {
				return errors.New("before hook error")
			}).AfterAsync(func(ctx context.Context, commit bool) {
				r.False(commit)
				wg.Done()
			})
			ti := orm.GetTxInfoInner(ctx, db)
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
		db, mock := orm.MockDB(r, nil)
		mock.ExpectBegin()
		mock.ExpectPrepare("UPDATE user SET `email` = ? WHERE `id` = ?").ExpectExec().WithArgs("a1", 1).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectPrepare("UPDATE user SET `email` = ? WHERE `id` = ?").ExpectExec().WithArgs("a2", 2).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()
		ctx := context.WithValue(context.Background(), "key", "value")
		err := db.Tx(ctx).Do(func(ctx context.Context) error {
			_, err := db.Update[orm.User](ctx).Set("email", "a1").Cond(func(c *orm.Cond) {
				c.Eq("id", 1)
			}).Do()
			if err != nil {
				return err
			}

			wg.Add(1)
			hook := db.TxHook(ctx).BeforeSync(func(ctx context.Context) error {
				_, err := db.Update[orm.User](ctx).Set("email", "a2").Cond(func(c *orm.Cond) {
					c.Eq("id", 2)
				}).Do()
				return err
			}).AfterAsync(func(ctx context.Context, commit bool) {
				ti := orm.GetTxInfoInner(ctx, db)
				val := ctx.Value("key").(string)
				r.True(commit)
				r.Nil(ti)
				r.Equal(val, "value")
				wg.Done()
			}, "key")
			r.NotNil(hook)
			return nil
		})
		r.NoError(err)
		wg.Wait()
	}
	{
		db, logger, mock := orm.MockDBAndLogger(r, nil)
		mock.ExpectBegin()
		mock.ExpectCommit()
		err := db.Tx(nil).Do(func(ctx context.Context) error {
			hook := db.TxHook(ctx).AfterAsync(func(ctx context.Context, commit bool) {
				panic("after-hook panic")
			})
			r.NotNil(hook)
			return nil
		})
		r.NoError(err)
		time.Sleep(time.Second * 1)
		r.Len(logger.Messages, 1)
		r.Equal(orm.Level_.Error, logger.Messages[0].Level)
		r.Contains(logger.Messages[0].Msg, "panic in transaction after-hook after-hook panic")
	}
}

func TestTxContext(t *testing.T) {
	r := require.New(t)
	tc := orm.NewTxContext(nil)
	_, ok := tc.Deadline()
	r.False(ok)

	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	tc = orm.NewTxContext(ctx)
	_, ok = tc.Deadline()
	r.True(ok)

	cancelFunc()
	r.Equal(tc.Err(), context.Canceled)
}
