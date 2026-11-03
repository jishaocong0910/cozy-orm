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
			if cause, ok := r.(error); ok {
				err = fmt.Errorf("%w\n%s", cause, deferStack())
			} else {
				err = fmt.Errorf("%v\n%s", r, deferStack())
			}
			return
		}

		err = t._commit()
	}()

	err = do(t.ctx)
	if err != nil {
		return
	}

	err = t._runBeforeCommitHook()
	if err != nil {
		return
	}
	return
}

func (t *tx) _open() error {
	if ti := t.db.getTxInfo(t.ctx); ti == nil {
		if t.db.rawDB == nil {
			return checkMust(t.must, errors.New("no available *sql.DB"))
		}

		rawTx, err := t.db.rawDB.BeginTx(t.ctx, t.txOptions)
		if err != nil {
			return checkMust(t.must, err)
		}

		t.ctx = context.WithValue(t.ctx, t.db, &txInfo{&txInfoData{rawTx: rawTx, owner: t}})
	}
	return nil
}

func (t *tx) _commit() error {
	if ti := t.db.getTxInfo(t.ctx); ti != nil && ti.owner == t {
		err := ti.rawTx.Commit()
		if err == nil {
			t._runAfterCompleteHook(true)
		}
		t._cleanTxInfo(t.db)
		return err
	}
	return nil
}

func (t *tx) _rollback() error {
	if ti := t.db.getTxInfo(t.ctx); ti != nil && ti.owner == t {
		err := ti.rawTx.Rollback()
		if err == nil {
			t._runAfterCompleteHook(false)
		}
		t._cleanTxInfo(t.db)
		return err
	}
	return nil
}

func (t *tx) _cleanTxInfo(db *DB) {
	if ti := db.getTxInfo(t.ctx); ti != nil {
		*ti = txInfo{}
	}
}

func (t *tx) _runBeforeCommitHook() error {
	if ti := t.db.getTxInfo(t.ctx); ti.owner == t {
		ti.hookRunning = true
		for _, hook := range t.db.getTxInfo(t.ctx).txHooks {
			if hook.beforeHandler != nil {
				err := hook.beforeHandler(hook.ctx)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *tx) _runAfterCompleteHook(commit bool) {
	for _, hook := range t.db.getTxInfo(t.ctx).txHooks {
		if hook.afterHandler != nil {
			ctx := context.Background()
			for _, key := range hook.copyCtxKeys {
				val := hook.ctx.Value(key)
				if val != nil {
					ctx = context.WithValue(ctx, key, val)
				}
			}
			go func() {
				defer func() {
					if r := recover(); r != nil {
						printLog(ctx, t.db.logger, LogLevel_.Error, "panic in transaction after-hook %v\n%s", r, deferStack())
					}
				}()
				hook.afterHandler(ctx, commit)
			}()
		}
	}
}

func newTx(ctx context.Context, db *DB) *tx {
	if ctx == nil {
		ctx = context.Background()
	}
	return &tx{ctx: ctx, db: db}
}

type txHook struct {
	ctx           context.Context
	beforeHandler func(ctx context.Context) error
	afterHandler  func(ctx context.Context, commit bool)
	copyCtxKeys   []any
}

func (t *txHook) BeforeCommitSync(handler func(ctx context.Context) error) {
	if t != nil {
		t.beforeHandler = handler
	}
}

func (t *txHook) AfterCompleteAsync(handler func(ctx context.Context, commit bool), copyCtxKeys ...any) {
	if t != nil {
		t.afterHandler = handler
		t.copyCtxKeys = copyCtxKeys
	}
}

type txInfoData struct {
	owner       *tx
	rawTx       *sql.Tx
	hookRunning bool
	txHooks     []*txHook
}

type txInfo struct {
	*txInfoData
}

func (t *txInfo) registerHook(ctx context.Context) (*txHook, bool) {
	if t.hookRunning {
		return nil, false
	}
	th := &txHook{ctx: ctx}
	t.txHooks = append(t.txHooks, th)
	return th, true
}
