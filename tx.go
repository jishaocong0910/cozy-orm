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
	"database/sql"
	"errors"
	"fmt"
)

type tx struct {
	ctx       context.Context
	db        *DB
	must      bool
	txOptions *sql.TxOptions
	ti        *txInfo
}

func (t *tx) Must() *tx {
	t.must = true
	return t
}

func (t *tx) TxOptions(txOptions *sql.TxOptions) *tx {
	t.txOptions = txOptions
	return t
}

func (t *tx) Do(do func(ctx context.Context) error) (err error) {
	err = t._open()
	if err != nil {
		return checkMust(t.must, err)
	}
	err = t._safeDo(do)
	return checkMust(t.must, err)
}

func (t *tx) _safeDo(do func(ctx context.Context) error) (err error) {
	defer func() {
		if err != nil {
			printWarn(t.ctx, t.db.logger, t._rollback())
			return
		}

		if r := recover(); r != nil {
			printWarn(t.ctx, t.db.logger, t._rollback())
			err = fmt.Errorf("%v\n%s", r, deferStack())
			return
		}

		err = t._commit()
	}()

	err = do(t.ctx)

	for _, hook := range t.ti.txHook_ {
		if hook.beforeHandler != nil {
			err = hook.beforeHandler(t.ctx)
			if err != nil {
				break
			}
		}
	}
	return
}

func (t *tx) _open() error {
	if t.ti = cvTx.get(t.ctx); !t.ti.matchingDb(t.db) {
		if t.db.sqlDB == nil {
			return checkMust(t.must, errors.New("no available *sql.DB"))
		}

		sqlTx, err := t.db.sqlDB.BeginTx(t.ctx, t.txOptions)
		if err != nil {
			return checkMust(t.must, err)
		}

		t.ti = &txInfo{&txInfoInner{sqlTx: sqlTx, creator: t}}
		t.ctx = cvTx.set(t.ctx, t.ti)
	}
	return nil
}

func (t *tx) _commit() error {
	if t.ti.creator == t {
		err := t.ti.sqlTx.Commit()
		if err == nil {
			t._runAfterHook(true)
		}
		cvTx.clean(t.ctx)
		return err
	}
	return nil
}

func (t *tx) _rollback() error {
	if t.ti.creator == t {
		err := t.ti.sqlTx.Rollback()
		if err == nil {
			t._runAfterHook(false)
		}
		cvTx.clean(t.ctx)
		return err
	}
	return nil
}

func (t *tx) _runAfterHook(commit bool) {
	for _, hook := range t.ti.txHook_ {
		if hook.afterHandler != nil {
			ctx := context.Background()
			for _, key := range hook.inheritCtxKey_ {
				val := t.ctx.Value(key)
				if val != nil {
					ctx = context.WithValue(ctx, key, val)
				}
			}
			go func() {
				defer func() {
					if r := recover(); r != nil {
						printLog(ctx, t.db.logger, Level_.Error, "panic in transaction after-hook %v\n%s", r, deferStack())
					}
				}()
				hook.afterHandler(ctx, commit)
			}()
		}
	}
}

type txInfoInner struct {
	creator *tx
	sqlTx   *sql.Tx
	txHook_ []*txHook
}

type txInfo struct {
	*txInfoInner
}

func (t *txInfo) isValid() bool {
	return t != nil && t.txInfoInner != nil
}

func (t *txInfo) matchingDb(db *DB) bool {
	return t.isValid() && t.creator.db == db
}

type txHook struct {
	beforeHandler  func(ctx context.Context) error
	afterHandler   func(ctx context.Context, commit bool)
	inheritCtxKey_ []any
}

func (t *txHook) BeforeSync(handler func(ctx context.Context) error) *txHook {
	t.beforeHandler = handler
	return t
}

func (t *txHook) AfterAsync(handler func(ctx context.Context, commit bool), inheritCtxKey_ ...any) *txHook {
	t.afterHandler = handler
	t.inheritCtxKey_ = inheritCtxKey_
	return t
}

func (t *txHook) Bind(ctx context.Context) bool {
	if ti := cvTx.get(ctx); ti.isValid() {
		ti.txHook_ = append(ti.txHook_, t)
		return true
	}
	return false
}

func TxHook() *txHook {
	return &txHook{}
}

func newTx(ctx context.Context, db *DB) *tx {
	if ctx == nil {
		ctx = context.Background()
	}
	return &tx{ctx: ctx, db: db}
}
