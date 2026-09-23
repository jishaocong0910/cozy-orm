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
	"reflect"
	"time"
)

type query[E any] struct {
	*executor
	mapTargets []*E
}

func (q *query[E]) Must() *query[E] {
	q.setMust()
	return q
}

func (q *query[E]) Describe(desc string) *query[E] {
	q.setDescribe(desc)
	return q
}

func (q *query[E]) SqlLogLevel(level Level) *query[E] {
	q.setSqlLogLevel(level)
	return q
}

func (q *query[E]) BuildSql(buildSql func(b *SqlBuilder)) *query[E] {
	q.setBuildSql(buildSql)
	return q
}

func (q *query[E]) MapTarget(entities ...*E) *query[E] {
	q.mapTargets = entities
	return q
}

func (q *query[E]) Do() (entities []*E, err error) {
	rows, columns, cost, cancel, err := q.doQuery()
	if err != nil {
		q.printSqlError(err)
		return []*E{}, checkMust(q.must, err)
	}
	if cancel {
		return []*E{}, nil
	}
	defer func() {
		printWarn(q.ctx, q.db.logger, rows.Close())
	}()

	mp, err := q.db.getMapper(reflect.TypeFor[E]())
	if err != nil {
		q.printSqlError(err)
		return []*E{}, checkMust(q.must, err)
	}

	var rowCount int
	if len(q.mapTargets) > 0 {
		rowCount, err = q._mapTargetEntities(rows, columns, mp)
	} else {
		rowCount, entities, err = q._mapNewEntities(rows, columns, mp)
	}
	q.printSqlRowCount(int64(rowCount), cost)
	return
}

func (q *query[E]) _mapTargetEntities(rows *sql.Rows, columns []string, mp mapper) (rowCount int, err error) {
	for rows.Next() {
		if len(q.mapTargets) > rowCount {
			e := q.mapTargets[rowCount]
			dest_, afterScan := mp.mapping(columns, e)
			err = rows.Scan(dest_...)
			if err != nil {
				return 0, checkMust(q.must, err)
			}
			afterScan()
		}
		rowCount++
	}
	err = rows.Err()
	if rows.Err() != nil {
		rowCount = -1
	}
	return
}

func (q *query[E]) _mapNewEntities(rows *sql.Rows, columns []string, mp mapper) (rowCount int, entities []*E, err error) {
	for rows.Next() {
		e := new(E)
		dest, afterScan := mp.mapping(columns, e)
		err = rows.Scan(dest...)
		if err != nil {
			return 0, []*E{}, checkMust(q.must, err)
		}
		afterScan()
		entities = append(entities, e)
		rowCount++
	}
	err = rows.Err()
	if err != nil {
		rowCount = -1
		entities = []*E{}
	}
	return
}

func newQuery[E any](e *executor) *query[E] {
	return &query[E]{executor: e}
}

type mutation struct {
	*executor
	mapTargetType reflect.Type
	mapTargets    []any
}

func (m *mutation) Must() *mutation {
	m.setMust()
	return m
}

func (m *mutation) Describe(desc string) *mutation {
	m.setDescribe(desc)
	return m
}

func (m *mutation) SqlLogLevel(level Level) *mutation {
	m.setSqlLogLevel(level)
	return m
}

func (m *mutation) BuildSql(buildSql func(b *SqlBuilder)) *mutation {
	m.setBuildSql(buildSql)
	return m
}

func (m *mutation) MapTarget[E any](entities ...*E) *mutation {
	m.mapTargetType = reflect.TypeFor[E]()
	m.mapTargets = make([]any, 0, len(entities))
	for _, e := range entities {
		m.mapTargets = append(m.mapTargets, e)
	}
	return m
}

func (m *mutation) Do() (affected int64, err error) {
	result, cost, cancel, err := m.doExec()
	if err != nil {
		m.printSqlError(err)
		return 0, checkMust(m.must, err)
	}
	if cancel {
		return 0, nil
	}
	affected, warn := result.RowsAffected()
	printWarn(m.ctx, m.db.logger, warn)
	m.printSqlAffected(affected, cost)
	m._getGenKey(result)
	return
}

func (m *mutation) _getGenKey(result sql.Result) {
	if m.executor.db.GetGeneratedKeyMode.Is(GetGeneratedKeyMode_.FirstInsertId, GetGeneratedKeyMode_.LastInsertId) && len(m.mapTargets) > 0 {
		id, warn := result.LastInsertId()
		if warn != nil {
			printWarn(m.ctx, m.db.logger, errors.New("get generated key fail, "+warn.Error()))
			return
		}
		ei, warn := m.db.getEntityInfo(m.mapTargetType)
		if warn != nil {
			printWarn(m.ctx, m.db.logger, errors.New("get generated key fail, "+warn.Error()))
			return
		}
		if len(ei.autoColumns) != 1 {
			printWarn(m.ctx, m.db.logger, errors.New("get generated key fail, the entity \""+
				m.mapTargetType.String()+"\" must have exactly one field with \"auto\" tag"))
			return
		}
		fieldIndex := ei.columnToFieldIndexMap[ei.autoColumns[0]]
		if m.db.GetGeneratedKeyMode.Is(GetGeneratedKeyMode_.LastInsertId) {
			id = id - int64(len(m.mapTargets)-1)*ei.lastInsertIdStep
		}
		for _, et := range m.mapTargets {
			v := reflect.ValueOf(et).Elem()
			field := v.Field(fieldIndex)
			field.Set(ei.lastInsertIdConversion(id))
			id += ei.lastInsertIdStep
		}
	}
}

func newMutation(e *executor) *mutation {
	return &mutation{executor: e}
}

type executor struct {
	ctx         context.Context
	db          *DB
	must        bool
	desc        string
	sqlLogLevel Level
	builder     *SqlBuilder
	buildSql    func(b *SqlBuilder)
}

func (e *executor) setMust() {
	e.must = true
}

func (e *executor) setDescribe(desc string) {
	e.desc = desc
}

func (e *executor) setSqlLogLevel(level Level) {
	e.sqlLogLevel = level
}

func (e *executor) setBuildSql(buildSql func(b *SqlBuilder)) {
	e.buildSql = buildSql
}

func (e *executor) printSqlRowCount(rowCount int64, cost time.Duration) {
	s, a := e.builder.sqlAndArgs()
	printSql(e.ctx, e.db.logger, e.sqlLogLevel, e._inTx(), e.desc, s, a, rowCount, -1, cost, nil)
}

func (e *executor) printSqlAffected(affected int64, cost time.Duration) {
	s, a := e.builder.sqlAndArgs()
	printSql(e.ctx, e.db.logger, e.sqlLogLevel, e._inTx(), e.desc, s, a, -1, affected, cost, nil)
}

func (e *executor) printSqlError(err error) {
	s, a := e.builder.sqlAndArgs()
	printSql(e.ctx, e.db.logger, e.sqlLogLevel, e._inTx(), e.desc, s, a, -1, -1, -1, err)
}

func (e *executor) doQuery() (*sql.Rows, []string, time.Duration, bool, error) {
	if e.buildSql == nil {
		return nil, nil, -1, true, nil
	}
	builder := newSqlBuilder(e.db.paramPrefix, e.db.quotedIdentifier)
	e.builder = builder
	e.buildSql(builder)
	if builder.err != nil {
		return nil, nil, -1, false, builder.err
	}
	if builder.cancel {
		return nil, nil, -1, true, nil
	}

	stmt, err := e._prepare(builder.b.String())
	if err != nil {
		return nil, nil, -1, false, err
	}
	defer func() {
		printWarn(e.ctx, e.db.logger, stmt.Close())
	}()

	start := time.Now()
	rows, err := stmt.QueryContext(e.ctx, convertArgs(builder.args)...)
	if err != nil {
		return nil, nil, -1, false, err
	}
	cost := time.Since(start)

	columns, err := rows.Columns()
	if err != nil { // coverage-ignore
		printWarn(e.ctx, e.db.logger, rows.Close())
		return nil, nil, -1, false, err
	}
	return rows, columns, cost, false, nil
}

func (e *executor) doExec() (sql.Result, time.Duration, bool, error) {
	if e.buildSql == nil {
		return nil, -1, true, nil
	}
	builder := newSqlBuilder(e.db.paramPrefix, e.db.quotedIdentifier)
	e.builder = builder
	e.buildSql(builder)
	if builder.err != nil {
		return nil, -1, false, builder.err
	}
	if builder.cancel {
		return nil, -1, true, nil
	}

	stmt, err := e._prepare(builder.b.String())
	if err != nil {
		return nil, -1, false, err
	}
	defer func() {
		printWarn(e.ctx, e.db.logger, stmt.Close())
	}()

	start := time.Now()
	result, err := stmt.ExecContext(e.ctx, convertArgs(builder.args)...)
	if err != nil {
		return nil, -1, false, err
	}
	cost := time.Since(start)

	return result, cost, false, nil
}

func (e *executor) _inTx() bool {
	return cvTx.get(e.ctx).matchingDb(e.db)
}

func (e *executor) _prepare(sqlStr string) (*sql.Stmt, error) {
	if ti := cvTx.get(e.ctx); ti.matchingDb(e.db) {
		return ti.sqlTx.PrepareContext(e.ctx, sqlStr)
	}
	if e.db.sqlDB == nil {
		return nil, errors.New("no available *sql.DB")
	}
	return e.db.sqlDB.PrepareContext(e.ctx, sqlStr)
}

func newExecutor(ctx context.Context, db *DB) *executor {
	if ctx == nil {
		ctx = context.Background()
	}
	return &executor{
		ctx:         ctx,
		db:          db,
		sqlLogLevel: db.sqlLogLevel,
	}
}
