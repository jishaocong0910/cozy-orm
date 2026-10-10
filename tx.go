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
	"time"
)

type tx struct {
	ctx       *TxContext
	db        *DB
	must      bool
	txOptions *sql.TxOptions
}

func (t *tx) Must() *tx {
	t.must = true
	return t
}

func (t *tx) TxOptions(txOptions *sql.TxOptions) *tx {
	t.txOptions = txOptions
	return t
}

func (t *tx) Do(handler func(ctx context.Context) error) (err error) {
	err = t._open()
	if err != nil {
		return checkMust(t.must, err)
	}
	err = t._safeDo(handler)
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

	for _, hook := range t.ctx.getTxInfo(t.db).txHooks {
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
	if ti := t.ctx.getTxInfo(t.db); ti == nil {
		if t.db.rawDB == nil {
			return checkMust(t.must, errors.New("no available *sql.DB"))
		}

		rawTx, err := t.db.rawDB.BeginTx(t.ctx, t.txOptions)
		if err != nil {
			return checkMust(t.must, err)
		}

		t.ctx.setTxInfo(&txInfo{&txInfoInner{rawTx: rawTx, owner: t}})
	}
	return nil
}

func (t *tx) _commit() error {
	if ti := t.ctx.getTxInfo(t.db); ti != nil && ti.owner == t {
		err := ti.rawTx.Commit()
		if err == nil {
			t._runAfterHook(true)
		}
		t.ctx.cleanTxInfo(t.db)
		return err
	}
	return nil
}

func (t *tx) _rollback() error {
	if ti := t.ctx.getTxInfo(t.db); ti != nil && ti.owner == t {
		err := ti.rawTx.Rollback()
		if err == nil {
			t._runAfterHook(false)
		}
		t.ctx.cleanTxInfo(t.db)
		return err
	}
	return nil
}

func (t *tx) _runAfterHook(commit bool) {
	for _, hook := range t.ctx.getTxInfo(t.db).txHooks {
		if hook.afterHandler != nil {
			ctx := context.Background()
			for _, key := range hook.inheritCtxKeys {
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

func newTx(ctx context.Context, db *DB) *tx {
	return &tx{ctx: newTxContext(ctx), db: db}
}

type txHook struct {
	beforeHandler  func(ctx context.Context) error
	afterHandler   func(ctx context.Context, commit bool)
	inheritCtxKeys []any
}

func (t *txHook) BeforeSync(handler func(ctx context.Context) error) *txHook {
	if t != nil {
		t.beforeHandler = handler
	}
	return t
}

func (t *txHook) AfterAsync(handler func(ctx context.Context, commit bool), inheritCtxKeys ...any) *txHook {
	if t != nil {
		t.afterHandler = handler
		t.inheritCtxKeys = inheritCtxKeys
	}
	return t
}

type TxContext struct {
	ctx context.Context
}

func (t *TxContext) Deadline() (deadline time.Time, ok bool) {
	return t.ctx.Deadline()
}

func (t *TxContext) Done() <-chan struct{} {
	return t.ctx.Done()
}

func (t *TxContext) Err() error {
	return t.ctx.Err()
}

func (t *TxContext) Value(key any) any {
	return t.ctx.Value(key)
}

func (t *TxContext) setTxInfo(ti *txInfo) {
	t.ctx = context.WithValue(t.ctx, ti.owner.db, ti)
}

func (t *TxContext) getTxInfo(db *DB) *txInfo {
	ti, _ := t.ctx.Value(db).(*txInfo)
	if ti != nil && ti.txInfoInner != nil {
		return ti
	}
	return nil
}

func (t *TxContext) cleanTxInfo(db *DB) {
	if ti := t.getTxInfo(db); ti != nil {
		*ti = txInfo{}
	}
}

func newTxContext(ctx context.Context) *TxContext {
	if ctx == nil {
		ctx = context.Background()
	}
	return &TxContext{ctx: ctx}
}

type txInfoInner struct {
	owner   *tx
	rawTx   *sql.Tx
	txHooks []*txHook
}

type txInfo struct {
	*txInfoInner
}
