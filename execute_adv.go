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
	"fmt"
	"reflect"
)

type find[E any] struct {
	query          *query[E]
	selectedSet    set[string]
	onDemand       *OnDemand
	orderBy        *orderBy
	page           *page
	condition      *Condition
	includeDeleted bool
	lastStr        string
}

func (f *find[E]) Must() *find[E] {
	f.query.Must()
	return f
}

func (f *find[E]) Describe(desc string) *find[E] {
	f.query.Describe(desc)
	return f
}

func (f *find[E]) SqlLogLevel(level Level) *find[E] {
	f.query.SqlLogLevel(level)
	return f
}

func (f *find[E]) Select(column_ ...string) *find[E] {
	f.selectedSet = newSet(column_...)
	return f
}

func (f *find[E]) OnDemand(onDemand *OnDemand) *find[E] {
	f.onDemand = onDemand
	return f
}

func (f *find[E]) Condition(cond *Condition) *find[E] {
	f.condition = cond
	return f
}

func (f *find[E]) OrderBy(orderBy *orderBy) *find[E] {
	f.orderBy = orderBy
	return f
}

func (f *find[E]) Page(page *page) *find[E] {
	f.page = page
	return f
}

func (f *find[E]) IncludeDeleted() *find[E] {
	f.includeDeleted = true
	return f
}

func (f *find[E]) LastStr(lastStr string) *find[E] {
	f.lastStr = lastStr
	return f
}

func (f *find[E]) Do() ([]*E, error) {
	es, err := f.query.BuildSql(func(b *SqlBuilder) {
		ei, err := f.query.db.getEntityInfo(reflect.TypeFor[E]())
		if err != nil {
			b.Error(err)
			return
		}

		columns := ei.getColumns(nil, f.onDemand, f.selectedSet, nil)
		if len(columns) == 0 {
			columns = ei.column_
		}
		b.Write("SELECT ")
		b.ForEach(b.Sep(", "), columns, func(_ int, column string) {
			b.WriteColumn(column)
		})
		b.Write(" FROM ").Write(ei.table).Accept(where{
			condition:      f.condition,
			policy:         ei.deleteSoftlyPolicy,
			includeDeleted: f.includeDeleted,
		}).Accept(f.orderBy)
		if f.page != nil {
			f.page.pageMode = f.query.db.pageMode
			b.Accept(f.page)
		}
		if f.lastStr != "" {
			b.Write(" ").Write(f.lastStr)
		}
	}).Do()
	return es, err
}

type findOne[E any] struct {
	find       *find[E]
	compatible bool
}

func (f *findOne[E]) Must() *findOne[E] {
	f.find.Must()
	return f
}

func (f *findOne[E]) Describe(desc string) *findOne[E] {
	f.find.Describe(desc)
	return f
}

func (f *findOne[E]) SqlLogLevel(level Level) *findOne[E] {
	f.find.SqlLogLevel(level)
	return f
}

func (f *findOne[E]) Select(column_ ...string) *findOne[E] {
	f.find.Select(column_...)
	return f
}

func (f *findOne[E]) OnDemand(onDemand *OnDemand) *findOne[E] {
	f.find.OnDemand(onDemand)
	return f
}

func (f *findOne[E]) Condition(cond *Condition) *findOne[E] {
	f.find.Condition(cond)
	return f
}

func (f *findOne[E]) OrderBy(orderBy *orderBy) *findOne[E] {
	f.find.OrderBy(orderBy)
	return f
}

func (f *findOne[E]) IncludeDeleted() *findOne[E] {
	f.find.IncludeDeleted()
	return f
}

func (f *findOne[E]) LastStr(lastStr string) *findOne[E] {
	f.find.LastStr(lastStr)
	return f
}

func (f *findOne[E]) Compatible() *findOne[E] {
	f.compatible = true
	return f
}

func (f *findOne[E]) Do() (*E, error) {
	entities, err := f.find.Do()
	var fst *E
	if len(entities) > 0 {
		if len(entities) > 1 && !f.compatible {
			err = errors.New("return more than one row")
		}
		fst = entities[0]
	}
	return fst, checkMust(f.find.query.must, err)
}

type insert[E any] struct {
	executor    *executor
	entity_     []*E
	nullableSet set[string]
	lastStr     string
}

func (i *insert[E]) Must() *insert[E] {
	i.executor.setMust()
	return i
}

func (i *insert[E]) Describe(desc string) *insert[E] {
	i.executor.setDescribe(desc)
	return i
}

func (i *insert[E]) SqlLogLevel(level Level) *insert[E] {
	i.executor.setSqlLogLevel(level)
	return i
}

func (i *insert[E]) Entities(entity_ ...*E) *insert[E] {
	i.entity_ = entity_
	return i
}

func (i *insert[E]) Nullable(column_ ...string) *insert[E] {
	i.nullableSet = newSet(column_...)
	return i
}

func (i *insert[E]) LastStr(lastStr string) *insert[E] {
	i.lastStr = lastStr
	return i
}

func (i *insert[E]) Do() (int64, error) {
	switch i.executor.db.GetGeneratedKeyMode.ID {
	case GetGeneratedKeyMode_.InsertReturning.ID, GetGeneratedKeyMode_.SQLServer.ID:
		_, err := newQuery[E](i.executor).MapTarget(i.entity_...).BuildSql(func(b *SqlBuilder) {
			i.buildSql(b)
		}).Do()
		return int64(len(i.entity_)), err
	default:
		m := newMutation(i.executor)
		if i.executor.db.GetGeneratedKeyMode.Is(GetGeneratedKeyMode_.FirstInsertId, GetGeneratedKeyMode_.LastInsertId) {
			m.MapTarget[E](i.entity_...)
		}
		return m.BuildSql(func(b *SqlBuilder) { i.buildSql(b) }).Do()
	}
}

func (i *insert[E]) buildSql(b *SqlBuilder) {
	if len(i.entity_) == 0 {
		b.Cancel()
		return
	}
	ei, err := i.executor.db.getEntityInfo(reflect.TypeFor[E]())
	if err != nil {
		b.Error(err)
		return
	}

	insertedColumns := ei.getColumns(i.entity_[0], nil,
		i.nullableSet.concat(ei.insertPolicy.forceColumnSet, ei.insertPolicy.defaultColumnSet),
		ei.insertPolicy.ignoredColumnSet)
	b.Write("INSERT INTO ").Write(ei.table)
	b.ForEach(b.SepFix("(", ", ", ")"), insertedColumns, func(_ int, column string) {
		b.WriteColumn(column)
	})
	if len(ei.autoColumn_) > 0 && i.executor.db.GetGeneratedKeyMode.IsPresent() {
		if GetGeneratedKeyMode_.SQLServer.Is(i.executor.db.GetGeneratedKeyMode) {
			i.executor.db.GetGeneratedKeyMode.writeSql(b, ei.autoColumn_)
			i._writeValuesClause(b, ei, insertedColumns)
		} else {
			i._writeValuesClause(b, ei, insertedColumns)
			if i.executor.db.GetGeneratedKeyMode.writeSql != nil {
				i.executor.db.GetGeneratedKeyMode.writeSql(b, ei.autoColumn_)
			}
		}
	} else {
		i._writeValuesClause(b, ei, insertedColumns)
	}
	if i.lastStr != "" {
		b.Write(" ").Write(i.lastStr)
	}
}

func (i *insert[E]) _writeValuesClause(b *SqlBuilder, ei *entityInfo, insertedColumn_ []string) {
	entityValueMap_ := make([]map[string]any, 0, len(i.entity_))
	for _, entity := range i.entity_ {
		entityValueMap_ = append(entityValueMap_, ei.getValueMap(entity, insertedColumn_))
	}
	am := newAssignedManager[E](i.executor.ctx, ei.insertPolicy, entityValueMap_, setColumns{}, "")
	b.Write(" VALUES ").ForEach(b.Sep(", "), i.entity_, func(i int, entity *E) {
		b.ForEach(b.SepFix("(", ", ", ")"), insertedColumn_, func(_ int, column string) {
			b.Accept(am.getValueWriter(column, i))
		})
	})
}

type update[E any] struct {
	mutation       *mutation
	entity         *E
	onDemand       *OnDemand
	setColumns     setColumns
	nullableSet    set[string]
	condition      *Condition
	includeDeleted bool
	skipSafety     bool
}

func (u *update[E]) Must() *update[E] {
	u.mutation.Must()
	return u
}

func (u *update[E]) Describe(desc string) *update[E] {
	u.mutation.Describe(desc)
	return u
}

func (u *update[E]) SqlLogLevel(level Level) *update[E] {
	u.mutation.SqlLogLevel(level)
	return u
}

func (u *update[E]) Entity(entity *E) *update[E] {
	u.entity = entity
	return u
}

func (u *update[E]) OnDemand(onDemand *OnDemand) *update[E] {
	u.onDemand = onDemand
	return u
}

func (u *update[E]) Set(column string, value any) *update[E] {
	u.setColumns.add(column, assignedValue{value: value})
	return u
}

func (u *update[E]) SetRaw(column string, sql string) *update[E] {
	u.setColumns.add(column, assignedRawSql{rawSql: sql})
	return u
}

func (u *update[E]) Nullable(column_ ...string) *update[E] {
	u.nullableSet = newSet(column_...)
	return u
}

func (u *update[E]) Condition(cond *Condition) *update[E] {
	u.condition = cond
	return u
}

func (u *update[E]) IncludeDeleted() *update[E] {
	u.includeDeleted = true
	return u
}

func (u *update[E]) SkipSafety() *update[E] {
	u.skipSafety = true
	return u
}

func (u *update[E]) Do() (int64, error) {
	if u.entity == nil && len(u.setColumns.columnSet) == 0 {
		return 0, nil
	}
	return u.mutation.BuildSql(func(b *SqlBuilder) {
		ei, err := u.mutation.db.getEntityInfo(reflect.TypeFor[E]())
		if err != nil {
			b.Error(err)
			return
		}

		updatedColumn_ := ei.getColumns(u.entity, u.onDemand,
			u.nullableSet.concat(u.setColumns.columnSet, ei.updatePolicy.forceColumnSet, ei.updatePolicy.defaultColumnSet),
			ei.updatePolicy.ignoredColumnSet)
		if len(updatedColumn_) == 0 {
			b.Cancel()
			return
		}
		entityValueMap := ei.getValueMap(u.entity, updatedColumn_, ei.pkColumn_)

		c := Cond()
		if len(ei.pkColumn_) > 0 && u.entity != nil {
			for _, column := range ei.pkColumn_ {
				if v, ok := entityValueMap[column]; ok {
					c.Eq(column, v)
				}
			}
		}
		c.Sub(u.condition)

		am := newAssignedManager[E](u.mutation.ctx, ei.updatePolicy, []map[string]any{ei.getValueMap(u.entity, updatedColumn_)}, u.setColumns, "")
		b.Write("UPDATE ").Write(ei.table).Write(" SET ")
		b.ForEach(b.Sep(", "), updatedColumn_, func(_ int, column string) {
			b.Write(column).Write(" = ").Accept(am.getValueWriter(column, 0))
		})
		b.Accept(where{
			condition:      c,
			policy:         ei.deleteSoftlyPolicy,
			includeDeleted: u.includeDeleted,
			safety:         !u.skipSafety,
		})
	}).Do()
}

type updateRow[E any] struct {
	mutation       *mutation
	entity_        []*E
	onDemand       *OnDemand
	setColumns     setColumns
	nullableSet    set[string]
	condition      *Condition
	includeDeleted bool
	skipSafety     bool
}

func (u *updateRow[E]) Must() *updateRow[E] {
	u.mutation.Must()
	return u
}

func (u *updateRow[E]) Describe(desc string) *updateRow[E] {
	u.mutation.Describe(desc)
	return u
}

func (u *updateRow[E]) SqlLogLevel(level Level) *updateRow[E] {
	u.mutation.SqlLogLevel(level)
	return u
}

func (u *updateRow[E]) OnDemand(onDemand *OnDemand) *updateRow[E] {
	u.onDemand = onDemand
	return u
}

func (u *updateRow[E]) Set(column string, value any) *updateRow[E] {
	u.setColumns.add(column, assignedValue{value: value})
	return u
}

func (u *updateRow[E]) SetRaw(column string, sql string) *updateRow[E] {
	u.setColumns.add(column, assignedRawSql{rawSql: sql})
	return u
}

func (u *updateRow[E]) Nullable(column_ ...string) *updateRow[E] {
	u.nullableSet = newSet(column_...)
	return u
}

func (u *updateRow[E]) Condition(cond *Condition) *updateRow[E] {
	u.condition = cond
	return u
}

func (u *updateRow[E]) IncludeDeleted() *updateRow[E] {
	u.includeDeleted = true
	return u
}

func (u *updateRow[E]) Entities(entity_ ...*E) *updateRow[E] {
	u.entity_ = entity_
	return u
}

func (u *updateRow[E]) Do() (int64, error) {
	if len(u.entity_) == 0 {
		return 0, nil
	}
	return u.mutation.BuildSql(func(b *SqlBuilder) {
		ei, err := u.mutation.db.getEntityInfo(reflect.TypeFor[E]())
		if err != nil {
			b.Error(err)
			return
		}
		if len(ei.pkColumn_) != 1 {
			b.Error(errors.New(ei.typ.String() + "\" must have exactly one field with the \"pk\" tag"))
			return
		}
		pkColumn := ei.pkColumn_[0]

		updatedColumns := ei.getColumns(u.entity_[0], u.onDemand,
			u.nullableSet.concat(u.setColumns.columnSet, ei.updatePolicy.forceColumnSet, ei.updatePolicy.defaultColumnSet),
			ei.updatePolicy.ignoredColumnSet)
		if len(updatedColumns) == 0 {
			b.Cancel()
			return
		}

		entityValueMap_ := make([]map[string]any, 0, len(u.entity_))
		pkValues := make([]any, 0, len(u.entity_))
		for i, entity := range u.entity_ {
			entityValueMap := ei.getValueMap(entity, updatedColumns, ei.pkColumn_)
			pkValue := entityValueMap[pkColumn]
			if pkValue == nil {
				b.Error(fmt.Errorf("field '%s' is nil at index %d of entities", ei.columnToFieldNameMap[pkColumn], i))
				return
			}
			pkValues = append(pkValues, pkValue)
			entityValueMap_ = append(entityValueMap_, entityValueMap)
		}

		am := newAssignedManager[E](u.mutation.ctx, ei.updatePolicy, entityValueMap_, u.setColumns, pkColumn)
		b.Write("UPDATE ").Write(ei.table).Write(" SET ")
		b.ForEach(b.Sep(", "), updatedColumns, func(i int, column string) {
			b.Write(column).Write(" = ").Accept(am.getValueWriter(column, -1))
		})
		var c *Condition
		if len(u.entity_) == 1 {
			c = Cond().Eq(pkColumn, pkValues[0]).Sub(u.condition)
		} else {
			c = Cond().In(pkColumn, pkValues).Sub(u.condition)
		}
		b.Accept(where{
			condition:      c,
			policy:         ei.deleteSoftlyPolicy,
			includeDeleted: u.includeDeleted,
			safety:         true,
		})
	}).Do()
}

type delete[E any] struct {
	mutation       *mutation
	condition      *Condition
	includeDeleted bool
	skipSafety     bool
}

func (d *delete[E]) Must() *delete[E] {
	d.mutation.Must()
	return d
}

func (d *delete[E]) Describe(desc string) *delete[E] {
	d.mutation.Describe(desc)
	return d
}

func (d *delete[E]) SqlLogLevel(level Level) *delete[E] {
	d.mutation.SqlLogLevel(level)
	return d
}

func (d *delete[E]) Condition(cond *Condition) *delete[E] {
	d.condition = cond
	return d
}

func (d *delete[E]) SkipSafety() *delete[E] {
	d.skipSafety = true
	return d
}

func (d *delete[E]) Do() (int64, error) {
	return d.mutation.BuildSql(func(b *SqlBuilder) {
		ei, err := d.mutation.db.getEntityInfo(reflect.TypeFor[E]())
		if err != nil {
			b.Error(err)
			return
		}
		b.Write("DELETE FROM ").Write(ei.table).Accept(where{
			condition: d.condition,
			safety:    !d.skipSafety,
		})
	}).Do()
}

type deleteSoftly[E any] struct {
	mutation   *mutation
	condition  *Condition
	skipSafety bool
}

func (d *deleteSoftly[E]) Must() *deleteSoftly[E] {
	d.mutation.Must()
	return d
}

func (d *deleteSoftly[E]) Describe(desc string) *deleteSoftly[E] {
	d.mutation.Describe(desc)
	return d
}

func (d *deleteSoftly[E]) SqlLogLevel(level Level) *deleteSoftly[E] {
	d.mutation.SqlLogLevel(level)
	return d
}

func (d *deleteSoftly[E]) Condition(cond *Condition) *deleteSoftly[E] {
	d.condition = cond
	return d
}

func (d *deleteSoftly[E]) SkipSafety() *deleteSoftly[E] {
	d.skipSafety = true
	return d
}

func (d *deleteSoftly[E]) Do() (int64, error) {
	return d.mutation.BuildSql(func(b *SqlBuilder) {
		ei, err := d.mutation.db.getEntityInfo(reflect.TypeFor[E]())
		if err != nil {
			b.Error(err)
			return
		}
		b.Write("UPDATE ").Write(ei.table).Write(" SET ").Write(ei.deleteSoftlyPolicy.deletedColumn).Write(" = ")
		switch ei.deleteSoftlyPolicy.mode.ID {
		case deleteSoftlyMode_.assignedPk.ID:
			b.Write(ei.deleteSoftlyPolicy.pkColumn)
		case deleteSoftlyMode_.assignedNull.ID:
			b.Write("NULL")
		default:
			b.Error(errors.New("feature is not supported"))
			return
		}
		b.Accept(where{
			condition: d.condition,
			policy:    ei.deleteSoftlyPolicy,
			safety:    !d.skipSafety,
		})
	}).Do()
}

type count[E any] struct {
	query          *query[Tuple[int64]]
	condition      *Condition
	includeDeleted bool
}

func (c *count[E]) Must() *count[E] {
	c.query.Must()
	return c
}

func (c *count[E]) Describe(desc string) *count[E] {
	c.query.Describe(desc)
	return c
}

func (c *count[E]) SqlLogLevel(level Level) *count[E] {
	c.query.SqlLogLevel(level)
	return c
}

func (c *count[E]) Condition(cond *Condition) *count[E] {
	c.condition = cond
	return c
}

func (c *count[E]) IncludeDeleted() *count[E] {
	c.includeDeleted = true
	return c
}

func (c *count[E]) Do() (i int64, err error) {
	es, err := c.query.BuildSql(func(b *SqlBuilder) {
		ei, err := c.query.db.getEntityInfo(reflect.TypeFor[E]())
		if err != nil {
			b.Error(err)
			return
		}
		b.Write("SELECT COUNT(*) FROM ").Write(ei.table).Accept(where{
			condition:      c.condition,
			policy:         ei.deleteSoftlyPolicy,
			includeDeleted: c.includeDeleted,
		})
	}).Do()
	if err == nil {
		i = es[0].Field
	}
	return
}

func newAssignedManager[E any](ctx context.Context, policy assignedPolicy, entitiesValue_ []map[string]any, setColumns setColumns, pk string) *assignedManager[E] {
	return &assignedManager[E]{
		ctx:                    ctx,
		policy:                 policy,
		entitiesValue_:         entitiesValue_,
		setColumns:             setColumns,
		pkColumn:               pk,
		reusePolicyValueWriter: make(map[string]SqlWriter, len(policy.reusedColumnSet)),
	}
}

type assignedManager[E any] struct {
	ctx                    context.Context
	policy                 assignedPolicy
	entitiesValue_         []map[string]any
	setColumns             setColumns
	pkColumn               string
	reusePolicyValueWriter map[string]SqlWriter
}

func (a *assignedManager[E]) getValueWriter(column string, entityIndex int) SqlWriter {
	if a.policy.forceColumnSet.contain(column) {
		return a._getPolicyValueWriter(column)
	}
	if vw, ok := a.setColumns.valueWriterMap[column]; ok {
		return vw
	}
	if entityIndex == -1 {
		if len(a.entitiesValue_) == 1 {
			entityIndex = 0
		} else {
			allNil := true
			caseItems := make([]assignedCaseItem, 0, len(a.entitiesValue_))
			for _, entityValueMap := range a.entitiesValue_ {
				value := entityValueMap[column]
				caseItems = append(caseItems, assignedCaseItem{
					caseValue: entityValueMap[a.pkColumn],
					thenValue: assignedValue{value: value},
				})
				if value != nil {
					allNil = false
				}
			}
			if allNil {
				return assignedValue{value: nil}
			}
			return assignedCases{pkColumn: a.pkColumn, caseItem_: caseItems}
		}
	}
	if value, ok := a.entitiesValue_[entityIndex][column]; ok {
		return assignedValue{value: value}
	}
	if a.policy.defaultColumnSet.contain(column) {
		return a._getPolicyValueWriter(column)
	}
	return assignedValue{value: nil}
}

func (a *assignedManager[E]) _getPolicyValueWriter(column string) SqlWriter {
	if vw, ok := a.reusePolicyValueWriter[column]; ok {
		return vw
	}
	var vm SqlWriter
	p := a.policy.assignedValueMap[column]
	if p.trueRawSqlFalseValue {
		vm = assignedRawSql{rawSql: p.rawSql(a.ctx)}
	} else {
		vm = assignedValue{value: p.value(a.ctx)}
	}
	if a.policy.reusedColumnSet.contain(column) {
		a.reusePolicyValueWriter[column] = vm
	}
	return vm
}

type setColumns struct {
	columnSet      set[string]
	valueWriterMap map[string]SqlWriter
}

func (s *setColumns) add(column string, valueWriter SqlWriter) {
	if s.columnSet.contain(column) {
		s.valueWriterMap[column] = valueWriter
	} else {
		if s.columnSet == nil {
			s.columnSet = newSet[string]()
			s.valueWriterMap = make(map[string]SqlWriter)
		}
		s.columnSet.add(column)
		s.valueWriterMap[column] = valueWriter
	}
}

type assignedValue struct {
	value any
}

func (a assignedValue) WriteSQL(b *SqlBuilder) {
	if a.value == nil {
		b.Write("NULL")
	} else {
		b.WritePh().Args(a.value)
	}
}

type assignedRawSql struct {
	rawSql string
}

func (a assignedRawSql) WriteSQL(b *SqlBuilder) {
	b.Write(a.rawSql)
}

type assignedCases struct {
	pkColumn  string
	caseItem_ []assignedCaseItem
}

func (a assignedCases) WriteSQL(b *SqlBuilder) {
	b.Write("CASE ").WriteColumn(a.pkColumn)
	for _, item := range a.caseItem_ {
		b.Accept(item)
	}
	b.Write(" END")
}

type assignedCaseItem struct {
	caseValue any
	thenValue assignedValue
}

func (a assignedCaseItem) WriteSQL(b *SqlBuilder) {
	b.Write(" WHEN ").WritePh().Args(a.caseValue).Write(" THEN ").Accept(a.thenValue)
}

type where struct {
	condition      *Condition
	policy         deleteSoftlyPolicy
	includeDeleted bool
	safety         bool
}

func (w where) WriteSQL(b *SqlBuilder) {
	c := w.condition
	if !c.notEmpty() {
		if w.safety {
			b.Error(errors.New("full table modification blocked"))
		}
		return
	}
	if w.policy.mode.IsPresent() && !w.includeDeleted {
		c = Cond().Sub(w.condition).Eq(w.policy.deletedColumn, w.policy.normalValue)
	}
	b.Write(" WHERE ").Accept(c)
}

func OrderBy() *orderBy {
	return &orderBy{}
}

func Page(offset, pageSize int) *page {
	return &page{offset: offset, pageSize: pageSize}
}

type orderBy struct {
	item_ []orderByItem
}

func (o *orderBy) Asc(column string) *orderBy {
	o.item_ = append(o.item_, orderByItem{column: column, seq: "ASC"})
	return o
}

func (o *orderBy) Desc(column string) *orderBy {
	o.item_ = append(o.item_, orderByItem{column: column, seq: "DESC"})
	return o
}

func (o *orderBy) WriteSQL(b *SqlBuilder) {
	if o != nil && len(o.item_) > 0 {
		b.Write(" ORDER BY ").ForEach(b.Sep(", "), o.item_, func(_ int, item orderByItem) {
			b.Accept(item)
		})
	}
}

type orderByItem struct {
	column string
	seq    string
}

func (o orderByItem) WriteSQL(b *SqlBuilder) {
	b.Write(o.column).Write(" ").Write(o.seq)
}

type page struct {
	pageMode         PageMode
	offset, pageSize int
}

func (p *page) WriteSQL(b *SqlBuilder) {
	if p != nil && p.pageMode.IsPresent() {
		p.pageMode.writeSql(b, p.offset, p.pageSize)
	}
}

func DemandFor[T any]() *OnDemand {
	return &OnDemand{t: reflect.TypeFor[T]()}
}

type OnDemand struct {
	t reflect.Type
}
