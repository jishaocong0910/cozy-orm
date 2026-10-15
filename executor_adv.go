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
	"reflect"
	"strconv"
	"strings"
)

func DemandFor[D any]() *Demand {
	return &Demand{t: reflect.TypeFor[D]()}
}

type Demand struct {
	t reflect.Type
}

type find[E any] struct {
	query          *query[E]
	selectedSet    set[string]
	demand         *Demand
	orderBy        *OrderBy
	page           *page
	cond           *Cond
	includeDeleted bool
	lastClause     string
}

func (f *find[E]) Must() *find[E] {
	f.query.Must()
	return f
}

func (f *find[E]) Description(desc string) *find[E] {
	f.query.Description(desc)
	return f
}

func (f *find[E]) SqlLogLevel(level Level) *find[E] {
	f.query.SqlLogLevel(level)
	return f
}

func (f *find[E]) Select(columns ...string) *find[E] {
	f.selectedSet = newSet(columns...)
	return f
}

func (f *find[E]) OnDemand(demand *Demand) *find[E] {
	f.demand = demand
	return f
}

func (f *find[E]) Cond(handler func(c *Cond)) *find[E] {
	if handler != nil {
		f.cond = &Cond{}
		handler(f.cond)
	}
	return f
}

func (f *find[E]) OrderBy(handler func(o *OrderBy)) *find[E] {
	if handler != nil {
		f.orderBy = &OrderBy{}
		handler(f.orderBy)
	}
	return f
}

func (f *find[E]) Page(offset, rowCount int) *find[E] {
	f.page = &page{offset: offset, rowCount: rowCount}
	return f
}

func (f *find[E]) IncludeDeleted() *find[E] {
	f.includeDeleted = true
	return f
}

func (f *find[E]) LastClause(lastClause string) *find[E] {
	f.lastClause = lastClause
	return f
}

func (f *find[E]) Do() ([]*E, error) {
	es, err := f.query.BuildSql(func(b *SQLBuilder) {
		ei, err := f.query.db.getEntityInfo(reflect.TypeFor[E]())
		if err != nil {
			b.Error(err)
			return
		}

		columns := ei.getColumns(nil, f.demand, f.selectedSet, nil)
		if len(columns) == 0 {
			columns = ei.columns
		}

		b.Text("SELECT ")
		b.ForEach(b.Sep(", "), columns, func(_ int, column string) {
			b.Quoted(column)
		})
		b.Text(" FROM ").Quoted(ei.table).accept(where{
			cond:           f.cond,
			policy:         ei.deleteSoftlyPolicy,
			includeDeleted: f.includeDeleted,
		}).accept(f.orderBy)
		if f.page != nil {
			f.page.pageMode = f.query.db.pageMode
			b.accept(f.page)
		}
		if f.lastClause != "" {
			b.Text(" ").Text(f.lastClause)
		}
	}).Do()
	return es, err
}

type findOne[E any] struct {
	find    *find[E]
	lenient bool
}

func (f *findOne[E]) Must() *findOne[E] {
	f.find.Must()
	return f
}

func (f *findOne[E]) Description(desc string) *findOne[E] {
	f.find.Description(desc)
	return f
}

func (f *findOne[E]) SqlLogLevel(level Level) *findOne[E] {
	f.find.SqlLogLevel(level)
	return f
}

func (f *findOne[E]) Select(columns ...string) *findOne[E] {
	f.find.Select(columns...)
	return f
}

func (f *findOne[E]) OnDemand(demand *Demand) *findOne[E] {
	f.find.OnDemand(demand)
	return f
}

func (f *findOne[E]) Cond(handler func(c *Cond)) *findOne[E] {
	f.find.Cond(handler)
	return f
}

func (f *findOne[E]) OrderBy(handler func(o *OrderBy)) *findOne[E] {
	f.find.OrderBy(handler)
	return f
}

func (f *findOne[E]) Page(offset, rowCount int) *findOne[E] {
	f.find.Page(offset, rowCount)
	return f
}

func (f *findOne[E]) IncludeDeleted() *findOne[E] {
	f.find.IncludeDeleted()
	return f
}

func (f *findOne[E]) LastClause(lastClause string) *findOne[E] {
	f.find.LastClause(lastClause)
	return f
}

func (f *findOne[E]) Lenient() *findOne[E] {
	f.lenient = true
	return f
}

func (f *findOne[E]) Do() (*E, error) {
	entities, err := f.find.Do()
	var fst *E
	if len(entities) > 0 {
		if len(entities) > 1 && !f.lenient {
			err = errors.New("return more than one row")
		}
		fst = entities[0]
	}
	return fst, checkMust(f.find.query.must, err)
}

type insert[E any] struct {
	executor    *executor
	entities    []*E
	requiredSet set[string]
}

func (i *insert[E]) Must() *insert[E] {
	i.executor.setMust()
	return i
}

func (i *insert[E]) Description(desc string) *insert[E] {
	i.executor.setDescription(desc)
	return i
}

func (i *insert[E]) SqlLogLevel(level Level) *insert[E] {
	i.executor.setSqlLogLevel(level)
	return i
}

func (i *insert[E]) Entities(entities ...*E) *insert[E] {
	i.entities = entities
	return i
}

func (i *insert[E]) Required(columns ...string) *insert[E] {
	i.requiredSet = newSet(columns...)
	return i
}

func (i *insert[E]) Do() (int64, error) {
	if len(i.entities) == 0 {
		return 0, nil
	}

	ei, err := i.executor.db.getEntityInfo(reflect.TypeFor[E]())
	if err != nil {
		return 0, checkMust(i.executor.must, err)
	}

	insertedColumns := ei.getColumns(i.entities[0], nil,
		i.requiredSet.concat(ei.insertPolicy.forceColumnSet, ei.insertPolicy.defaultColumnSet),
		ei.insertPolicy.ignoredColumnSet)

	if len(ei.autoColumns) > 0 {
		switch i.executor.db.getGeneratedKeyMode.ID {
		case GetGeneratedKeyMode_.FirstInsertId.ID, GetGeneratedKeyMode_.LastInsertId.ID:
			return i._doGetIdentify(ei, insertedColumns)
		case GetGeneratedKeyMode_.InsertReturning.ID:
			return i._doInsertReturning(ei, insertedColumns)
		case GetGeneratedKeyMode_.SQLServer.ID:
			return i._doSQLServer(ei, insertedColumns)
		case GetGeneratedKeyMode_.Oracle.ID:
			return i._doOracle(ei, insertedColumns)
		}
	}
	return i._doDefault(ei, insertedColumns)
}

func (i *insert[E]) _doDefault(ei *entityInfo, insertedColumns []string) (int64, error) {
	return newMutation(i.executor).BuildSql(func(b *SQLBuilder) {
		i._buildInsertColumnsClause(b, ei, insertedColumns)
		i._buildValuesClause(b, ei, insertedColumns)
	}).Do()
}

func (i *insert[E]) _doGetIdentify(ei *entityInfo, insertedColumns []string) (int64, error) {
	return newMutation(i.executor).MapTo[E](i.entities...).BuildSql(func(b *SQLBuilder) {
		i._buildInsertColumnsClause(b, ei, insertedColumns)
		i._buildValuesClause(b, ei, insertedColumns)
	}).Do()
}

func (i *insert[E]) _doInsertReturning(ei *entityInfo, insertedColumns []string) (int64, error) {
	_, err := newQuery[E](i.executor).MapTo(i.entities...).BuildSql(func(b *SQLBuilder) {
		i._buildInsertColumnsClause(b, ei, insertedColumns)
		i._buildValuesClause(b, ei, insertedColumns)
		b.ForEach(b.SepWrapOpt(" RETURNING ", ", ", ""), ei.autoColumns, func(_ int, column string) {
			b.Quoted(column)
		})
	}).Do()
	return int64(len(i.entities)), err
}

func (i *insert[E]) _doSQLServer(ei *entityInfo, insertedColumns []string) (int64, error) {
	_, err := newQuery[E](i.executor).MapTo(i.entities...).BuildSql(func(b *SQLBuilder) {
		i._buildInsertColumnsClause(b, ei, insertedColumns)
		b.ForEach(b.SepWrapOpt(" OUTPUT ", ", ", ""), ei.autoColumns, func(_ int, column string) {
			b.Text("INSERTED.").Quoted(column)
		})
		i._buildValuesClause(b, ei, insertedColumns)
	}).Do()
	return int64(len(i.entities)), err
}

func (i *insert[E]) _doOracle(ei *entityInfo, insertedColumns []string) (int64, error) {
	if ei.identityConverter == nil {
		return i._doDefault(ei, insertedColumns)
	}
	if len(i.entities) > 1000 {
		return 0, checkMust(i.executor.must, errors.New("entity count cannot exceed 1000 in Oracle"))
	}

	identifyColumn := ei.autoColumns[0]

	if len(i.entities) == 1 {
		v := reflect.ValueOf(i.entities[0]).Elem()
		fieldIndex := ei.columnToFieldIndexMap[ei.autoColumns[0]]
		field := v.Field(fieldIndex)
		affected, err := newMutation(i.executor).MapTo[E](i.entities...).BuildSql(func(b *SQLBuilder) {
			i._buildInsertColumnsClause(b, ei, insertedColumns)
			i._buildValuesClause(b, ei, insertedColumns)
			b.Text(" RETURNING ").Quoted(identifyColumn).Text(" INTO out_id")
			b.Args(sql.Named("out_id", sql.Out{Dest: field.Addr().Interface()}))
		}).Do()
		return affected, err
	}

	results := string(make([]byte, 32767))
	m := newMutation(i.executor)
	affected, err := m.BuildSql(func(b *SQLBuilder) {
		am := i._getAssignedManager(ei, insertedColumns)
		b.Text("DECLARE l_id NUMBER; l_results VARCHAR2(32767) := ''; BEGIN")
		for j := 0; j < len(i.entities); j++ {
			b.Text(" ")
			i._buildInsertColumnsClause(b, ei, insertedColumns)
			b.ForEach(b.SepWrap(" VALUES(", ", ", ")"), insertedColumns, func(_ int, column string) {
				b.accept(am.getValueWriter(column, j))
			}).Text(" RETURNING ").Quoted(identifyColumn).Text(" INTO l_id; l_results := l_results || ',' || TO_CHAR(l_id);")
		}
		b.Text(" :out_results := l_results; EXCEPTION WHEN OTHERS THEN ROLLBACK; RAISE; END;")
		b.Args(sql.Named("out_results", sql.Out{Dest: &results}))
	}).Do()
	if err == nil {
		for j, r := range strings.Split(results, ",")[1:] {
			id, _ := strconv.ParseInt(r, 10, 64)
			m.getIdentify(i.entities[j], ei, id)
		}
	}
	return affected, err
}

func (i *insert[E]) _buildInsertColumnsClause(b *SQLBuilder, ei *entityInfo, insertedColumns []string) {
	b.Text("INSERT INTO ").Quoted(ei.table)
	b.ForEach(b.SepWrap("(", ", ", ")"), insertedColumns, func(_ int, column string) {
		b.Quoted(column)
	})
}

func (i *insert[E]) _buildValuesClause(b *SQLBuilder, ei *entityInfo, insertedColumns []string) {
	am := i._getAssignedManager(ei, insertedColumns)
	b.Text(" VALUES").ForEach(b.Sep(", "), i.entities, func(i int, entity *E) {
		b.ForEach(b.SepWrap("(", ", ", ")"), insertedColumns, func(_ int, column string) {
			b.accept(am.getValueWriter(column, i))
		})
	})
}

func (i *insert[E]) _getAssignedManager(ei *entityInfo, insertedColumns []string) *assignedManager[E] {
	entityValueMaps := make([]map[string]any, 0, len(i.entities))
	for _, entity := range i.entities {
		entityValueMaps = append(entityValueMaps, ei.getValueMap(entity, insertedColumns))
	}
	return newAssignedManager[E](i.executor.ctx, ei.insertPolicy, entityValueMaps, nil, "")
}

type update[E any] struct {
	mutation       *mutation
	entity         *E
	demand         *Demand
	setColumns     setColumns
	requiredSet    set[string]
	cond           *Cond
	includeDeleted bool
	skipSafety     bool
}

func (u *update[E]) Must() *update[E] {
	u.mutation.Must()
	return u
}

func (u *update[E]) Description(desc string) *update[E] {
	u.mutation.Description(desc)
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

func (u *update[E]) OnDemand(demand *Demand) *update[E] {
	u.demand = demand
	return u
}

func (u *update[E]) Set(column string, value any) *update[E] {
	u.setColumns.add(column, assignedValue{value: value})
	return u
}

func (u *update[E]) SetExpr(column string, expr string) *update[E] {
	u.setColumns.add(column, assignedExpr{expr: expr})
	return u
}

func (u *update[E]) Required(columns ...string) *update[E] {
	u.requiredSet = newSet(columns...)
	return u
}

func (u *update[E]) Cond(handler func(c *Cond)) *update[E] {
	if handler != nil {
		u.cond = &Cond{}
		handler(u.cond)
	}
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
	return u.mutation.BuildSql(func(b *SQLBuilder) {
		ei, err := u.mutation.db.getEntityInfo(reflect.TypeFor[E]())
		if err != nil {
			b.Error(err)
			return
		}

		updatedColumns := ei.getColumns(u.entity, u.demand,
			u.requiredSet.concat(u.setColumns.columnSet, ei.updatePolicy.forceColumnSet, ei.updatePolicy.defaultColumnSet),
			ei.updatePolicy.ignoredColumnSet)
		if len(updatedColumns) == 0 {
			b.Cancel()
			return
		}

		entityValueMap := ei.getValueMap(u.entity, updatedColumns, ei.pkColumns)
		c := &Cond{}
		if len(ei.pkColumns) > 0 && u.entity != nil {
			for _, column := range ei.pkColumns {
				if v, ok := entityValueMap[column]; ok {
					c.Eq(column, v)
				}
			}
		}
		c.add(u.cond)

		am := newAssignedManager[E](u.mutation.ctx, ei.updatePolicy, []map[string]any{ei.getValueMap(u.entity, updatedColumns)}, &u.setColumns, "")
		b.Text("UPDATE ").Quoted(ei.table).Text(" SET ")
		b.ForEach(b.Sep(", "), updatedColumns, func(_ int, column string) {
			b.Quoted(column).Text(" = ").accept(am.getValueWriter(column, 0))
		})

		b.accept(where{
			cond:           c,
			policy:         ei.deleteSoftlyPolicy,
			includeDeleted: u.includeDeleted,
			safety:         !u.skipSafety,
		})
	}).Do()
}

type updateRow[E any] struct {
	mutation       *mutation
	entities       []*E
	demand         *Demand
	setColumns     setColumns
	requiredSet    set[string]
	cond           *Cond
	includeDeleted bool
	skipSafety     bool
}

func (u *updateRow[E]) Must() *updateRow[E] {
	u.mutation.Must()
	return u
}

func (u *updateRow[E]) Description(desc string) *updateRow[E] {
	u.mutation.Description(desc)
	return u
}

func (u *updateRow[E]) SqlLogLevel(level Level) *updateRow[E] {
	u.mutation.SqlLogLevel(level)
	return u
}

func (u *updateRow[E]) Entities(entities ...*E) *updateRow[E] {
	u.entities = entities
	return u
}

func (u *updateRow[E]) Required(columns ...string) *updateRow[E] {
	u.requiredSet = newSet(columns...)
	return u
}

func (u *updateRow[E]) OnDemand(demand *Demand) *updateRow[E] {
	u.demand = demand
	return u
}

func (u *updateRow[E]) Set(column string, value any) *updateRow[E] {
	u.setColumns.add(column, assignedValue{value: value})
	return u
}

func (u *updateRow[E]) SetExpr(column string, expr string) *updateRow[E] {
	u.setColumns.add(column, assignedExpr{expr: expr})
	return u
}

func (u *updateRow[E]) Cond(handler func(c *Cond)) *updateRow[E] {
	if handler != nil {
		u.cond = &Cond{}
		handler(u.cond)
	}
	return u
}

func (u *updateRow[E]) IncludeDeleted() *updateRow[E] {
	u.includeDeleted = true
	return u
}

func (u *updateRow[E]) Do() (int64, error) {
	if len(u.entities) == 0 {
		return 0, nil
	}
	return u.mutation.BuildSql(func(b *SQLBuilder) {
		ei, err := u.mutation.db.getEntityInfo(reflect.TypeFor[E]())
		if err != nil {
			b.Error(err)
			return
		}

		if len(ei.pkColumns) != 1 {
			b.Error(errors.New(ei.typ.String() + "\" must have exactly one field with the \"pk\" tag"))
			return
		}

		pkColumn := ei.pkColumns[0]
		updatedColumns := ei.getColumns(u.entities[0], u.demand,
			u.requiredSet.concat(u.setColumns.columnSet, ei.updatePolicy.forceColumnSet, ei.updatePolicy.defaultColumnSet),
			ei.updatePolicy.ignoredColumnSet)
		if len(updatedColumns) == 0 {
			b.Cancel()
			return
		}

		entityValueMaps := make([]map[string]any, 0, len(u.entities))
		pkValues := make([]any, 0, len(u.entities))
		for i, entity := range u.entities {
			entityValueMap := ei.getValueMap(entity, updatedColumns, ei.pkColumns)
			pkValue := entityValueMap[pkColumn]
			if pkValue == nil {
				b.Error(fmt.Errorf("field '%s' is nil at index %d of entities", ei.columnToFieldNameMap[pkColumn], i))
				return
			}
			pkValues = append(pkValues, pkValue)
			entityValueMaps = append(entityValueMaps, entityValueMap)
		}

		am := newAssignedManager[E](u.mutation.ctx, ei.updatePolicy, entityValueMaps, &u.setColumns, pkColumn)
		b.Text("UPDATE ").Quoted(ei.table).Text(" SET ")
		b.ForEach(b.Sep(", "), updatedColumns, func(i int, column string) {
			b.Quoted(column).Text(" = ").accept(am.getValueWriter(column, -1))
		})

		c := &Cond{}
		if len(u.entities) == 1 {
			c = c.Eq(pkColumn, pkValues[0]).add(u.cond)
		} else {
			c = c.In(pkColumn, pkValues).add(u.cond)
		}
		b.accept(where{
			cond:           c,
			policy:         ei.deleteSoftlyPolicy,
			includeDeleted: u.includeDeleted,
			safety:         true,
		})
	}).Do()
}

type delete[E any] struct {
	mutation       *mutation
	cond           *Cond
	includeDeleted bool
	skipSafety     bool
}

func (d *delete[E]) Must() *delete[E] {
	d.mutation.Must()
	return d
}

func (d *delete[E]) Description(desc string) *delete[E] {
	d.mutation.Description(desc)
	return d
}

func (d *delete[E]) SqlLogLevel(level Level) *delete[E] {
	d.mutation.SqlLogLevel(level)
	return d
}

func (d *delete[E]) Cond(handler func(c *Cond)) *delete[E] {
	if handler != nil {
		d.cond = &Cond{}
		handler(d.cond)
	}
	return d
}

func (d *delete[E]) SkipSafety() *delete[E] {
	d.skipSafety = true
	return d
}

func (d *delete[E]) Do() (int64, error) {
	return d.mutation.BuildSql(func(b *SQLBuilder) {
		ei, err := d.mutation.db.getEntityInfo(reflect.TypeFor[E]())
		if err != nil {
			b.Error(err)
			return
		}
		b.Text("DELETE FROM ").Quoted(ei.table).accept(where{
			cond:   d.cond,
			safety: !d.skipSafety,
		})
	}).Do()
}

type deleteSoftly[E any] struct {
	mutation   *mutation
	cond       *Cond
	skipSafety bool
}

func (d *deleteSoftly[E]) Must() *deleteSoftly[E] {
	d.mutation.Must()
	return d
}

func (d *deleteSoftly[E]) Description(desc string) *deleteSoftly[E] {
	d.mutation.Description(desc)
	return d
}

func (d *deleteSoftly[E]) SqlLogLevel(level Level) *deleteSoftly[E] {
	d.mutation.SqlLogLevel(level)
	return d
}

func (d *deleteSoftly[E]) Cond(handler func(c *Cond)) *deleteSoftly[E] {
	if handler != nil {
		d.cond = &Cond{}
		handler(d.cond)
	}
	return d
}

func (d *deleteSoftly[E]) SkipSafety() *deleteSoftly[E] {
	d.skipSafety = true
	return d
}

func (d *deleteSoftly[E]) Do() (int64, error) {
	return d.mutation.BuildSql(func(b *SQLBuilder) {
		ei, err := d.mutation.db.getEntityInfo(reflect.TypeFor[E]())
		if err != nil {
			b.Error(err)
			return
		}
		b.Text("UPDATE ").Quoted(ei.table).Text(" SET ").Quoted(ei.deleteSoftlyPolicy.deletedColumn).Text(" = ")
		switch ei.deleteSoftlyPolicy.mode.ID {
		case deleteSoftlyMode_.assignedPk.ID:
			b.Quoted(ei.deleteSoftlyPolicy.pkColumn)
		case deleteSoftlyMode_.assignedNull.ID:
			b.Text("NULL")
		default:
			b.Error(errors.New("delete softly mode is undefined"))
			return
		}
		b.accept(where{
			cond:   d.cond,
			policy: ei.deleteSoftlyPolicy,
			safety: !d.skipSafety,
		})
	}).Do()
}

type count[E any] struct {
	query          *query[Tuple[int64]]
	cond           *Cond
	includeDeleted bool
}

func (c *count[E]) Must() *count[E] {
	c.query.Must()
	return c
}

func (c *count[E]) Description(desc string) *count[E] {
	c.query.Description(desc)
	return c
}

func (c *count[E]) SqlLogLevel(level Level) *count[E] {
	c.query.SqlLogLevel(level)
	return c
}

func (c *count[E]) Cond(handler func(c *Cond)) *count[E] {
	if handler != nil {
		c.cond = &Cond{}
		handler(c.cond)
	}
	return c
}

func (c *count[E]) IncludeDeleted() *count[E] {
	c.includeDeleted = true
	return c
}

func (c *count[E]) Do() (i int64, err error) {
	es, err := c.query.BuildSql(func(b *SQLBuilder) {
		ei, err := c.query.db.getEntityInfo(reflect.TypeFor[E]())
		if err != nil {
			b.Error(err)
			return
		}
		b.Text("SELECT COUNT(*) FROM ").Quoted(ei.table).accept(where{
			cond:           c.cond,
			policy:         ei.deleteSoftlyPolicy,
			includeDeleted: c.includeDeleted,
		})
	}).Do()
	if err == nil {
		i = es[0].Field
	}
	return
}

func newAssignedManager[E any](ctx context.Context, policy assignedPolicy, entitiesValueMap []map[string]any, setColumns *setColumns, pk string) *assignedManager[E] {
	return &assignedManager[E]{
		ctx:                    ctx,
		policy:                 policy,
		entitiesValueMaps:      entitiesValueMap,
		setColumns:             setColumns,
		pkColumn:               pk,
		reusePolicyValueWriter: make(map[string]sqlWriter, len(policy.reusedColumnSet)),
	}
}

type assignedManager[E any] struct {
	ctx                    context.Context
	policy                 assignedPolicy
	entitiesValueMaps      []map[string]any
	setColumns             *setColumns
	pkColumn               string
	reusePolicyValueWriter map[string]sqlWriter
}

func (a *assignedManager[E]) getValueWriter(column string, entityIndex int) sqlWriter {
	if a.policy.forceColumnSet.contain(column) {
		return a._getPolicyValueWriter(column)
	}
	if a.setColumns != nil {
		if vw, ok := a.setColumns.valueWriterMap[column]; ok {
			return vw
		}
	}
	if a.pkColumn != "" {
		if len(a.entitiesValueMaps) == 1 {
			entityIndex = 0
		} else {
			allNil := true
			caseItems := make([]assignedCaseItem, 0, len(a.entitiesValueMaps))
			for _, entityValueMap := range a.entitiesValueMaps {
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
			return assignedCases{pkColumn: a.pkColumn, caseItems: caseItems}
		}
	}
	if value, ok := a.entitiesValueMaps[entityIndex][column]; ok {
		return assignedValue{value: value}
	}
	if a.policy.defaultColumnSet.contain(column) {
		return a._getPolicyValueWriter(column)
	}
	return assignedValue{value: nil}
}

func (a *assignedManager[E]) _getPolicyValueWriter(column string) sqlWriter {
	if vw, ok := a.reusePolicyValueWriter[column]; ok {
		return vw
	}
	var vm sqlWriter
	p := a.policy.assignedValueMap[column]
	if p.trueExprFalseValue {
		vm = assignedExpr{expr: p.expr(a.ctx)}
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
	valueWriterMap map[string]sqlWriter
}

func (s *setColumns) add(column string, valueWriter sqlWriter) {
	if s.columnSet.contain(column) {
		s.valueWriterMap[column] = valueWriter
	} else {
		if s.columnSet == nil {
			s.columnSet = newSet[string]()
			s.valueWriterMap = make(map[string]sqlWriter)
		}
		s.columnSet.add(column)
		s.valueWriterMap[column] = valueWriter
	}
}

type assignedValue struct {
	value any
}

func (a assignedValue) writeSQL(b *SQLBuilder) {
	if a.value == nil {
		b.Text("NULL")
	} else {
		b.Arg(a.value)
	}
}

type assignedExpr struct {
	expr string
}

func (a assignedExpr) writeSQL(b *SQLBuilder) {
	b.Text(a.expr)
}

type assignedCases struct {
	pkColumn  string
	caseItems []assignedCaseItem
}

func (a assignedCases) writeSQL(b *SQLBuilder) {
	b.Text("CASE ").Quoted(a.pkColumn)
	for _, item := range a.caseItems {
		b.accept(item)
	}
	b.Text(" END")
}

type assignedCaseItem struct {
	caseValue any
	thenValue assignedValue
}

func (a assignedCaseItem) writeSQL(b *SQLBuilder) {
	b.Text(" WHEN ").Arg(a.caseValue).Text(" THEN ").accept(a.thenValue)
}

type where struct {
	policy         deleteSoftlyPolicy
	cond           *Cond
	includeDeleted bool
	safety         bool
}

func (w where) writeSQL(b *SQLBuilder) {
	c := w.cond
	if c.isEmpty() {
		if w.safety {
			b.Error(errors.New("full table modification blocked"))
		}
		return
	}
	if w.policy.mode.IsPresent() && !w.includeDeleted {
		c = (&Cond{}).add(w.cond).Eq(w.policy.deletedColumn, w.policy.normalValue)
	}
	b.Text(" WHERE ").accept(c)
}

type OrderBy struct {
	items []orderByItem
}

func (o *OrderBy) Asc(column string) *OrderBy {
	o.items = append(o.items, orderByItem{column: column, seq: "ASC"})
	return o
}

func (o *OrderBy) Desc(column string) *OrderBy {
	o.items = append(o.items, orderByItem{column: column, seq: "DESC"})
	return o
}

func (o *OrderBy) writeSQL(b *SQLBuilder) {
	if o != nil && len(o.items) > 0 {
		b.Text(" ORDER BY ").ForEach(b.Sep(", "), o.items, func(_ int, item orderByItem) {
			b.accept(item)
		})
	}
}

type orderByItem struct {
	column string
	seq    string
}

func (o orderByItem) writeSQL(b *SQLBuilder) {
	b.Quoted(o.column).Text(" ").Text(o.seq)
}

type page struct {
	pageMode         PageMode
	offset, rowCount int
}

func (p page) writeSQL(b *SQLBuilder) {
	if p.pageMode.IsPresent() {
		p.pageMode.writeSQL(b, p.offset, p.rowCount)
	}
}
