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
	"reflect"
	"sync"
	"time"
)

type DB struct {
	sqlDB               *sql.DB
	logger              Logger
	sqlLogLevel         Level
	tabNameMapper       *NameMapper
	colNameMapper       *NameMapper
	paramPrefix         string
	GetGeneratedKeyMode GetGeneratedKeyMode
	pageMode            PageMode
	quotedIdentifier    QuotedIdentifier
	columnPolicyConfig_ []*columnPolicyConfig

	entities           sync.Map
	mappers            sync.Map
	registerEntityLock sync.Mutex
	registerMapperLock sync.Mutex
}

// Raw returns the underlying *sql.DB
func (d *DB) Raw() *sql.DB {
	return d.sqlDB
}

func (d *DB) Query[E any](ctx context.Context) *query[E] {
	return newQuery[E](newExecutor(ctx, d))
}

func (d *DB) Mutation(ctx context.Context) *mutation {
	return newMutation(newExecutor(ctx, d))
}

func (d *DB) Find[E any](ctx context.Context) *find[E] {
	return &find[E]{query: d.Query[E](ctx)}
}

func (d *DB) FindOne[E any](ctx context.Context) *findOne[E] {
	return &findOne[E]{find: d.Find[E](ctx)}
}

func (d *DB) Insert[E any](ctx context.Context) *insert[E] {
	return &insert[E]{executor: newExecutor(ctx, d)}
}

func (d *DB) Update[E any](ctx context.Context) *update[E] {
	return &update[E]{mutation: d.Mutation(ctx)}
}

func (d *DB) UpdateRow[E any](ctx context.Context) *updateRow[E] {
	return &updateRow[E]{mutation: d.Mutation(ctx)}
}

func (d *DB) Delete[E any](ctx context.Context) *delete[E] {
	return &delete[E]{mutation: d.Mutation(ctx)}
}

func (d *DB) DeleteSoftly[E any](ctx context.Context) *deleteSoftly[E] {
	return &deleteSoftly[E]{mutation: d.Mutation(ctx)}
}

func (d *DB) Count[E any](ctx context.Context) *count[E] {
	return &count[E]{query: d.Query[Tuple[int64]](ctx)}
}

func (d *DB) Tx(ctx context.Context) *tx {
	return newTx(ctx, d)
}

func (d *DB) getEntityInfo(t reflect.Type) (*entityInfo, error) {
	if val, ok := d.entities.Load(t); ok {
		return val.(*entityInfo), nil
	}

	d.registerEntityLock.Lock()
	defer d.registerEntityLock.Unlock()

	if val, ok := d.entities.Load(t); ok { // coverage-ignore
		return val.(*entityInfo), nil
	}

	ei, err := newEntityInfo(t, d.tabNameMapper, d.colNameMapper, d.columnPolicyConfig_)
	if ei != nil {
		d.entities.Store(t, ei)
	}
	return ei, err
}

func (d *DB) getMapper(t reflect.Type) (mapper, error) {
	if val, ok := d.mappers.Load(t); ok {
		return val.(mapper), nil
	}

	d.registerMapperLock.Lock()
	defer d.registerMapperLock.Unlock()

	if m, ok := d.mappers.Load(t); ok { // coverage-ignore
		return m.(mapper), nil
	}

	ei, _ := d.getEntityInfo(t)
	m, err := newMapper(ei, t)
	if m != nil {
		d.mappers.Store(t, m)
	}
	return m, err
}

type DBConfig struct {
	SqlDB       *sql.DB
	Logger      Logger
	SqlLogLevel Level
	// TabNameMapper setting name mapping from entity to table
	TabNameMapper *NameMapper
	// setting field name mapping from entity to table
	ColNameMapper *NameMapper
	DBType        DBType
	// setting the SQL parameter placeholder prefix; if empty, parameter placeholder uses the default "?",
	// otherwise, use the specified prefix appended to the auto increment number started at 1. This effect is manifested
	// in method SqlBuilder.WritePh()
	ParamPrefix string
	// indicates how to get id when executing INSERT SQL
	GetGeneratedKeyMode GetGeneratedKeyMode
	// indicates the paging clause
	PageMode            PageMode
	QuotedIdentifier    QuotedIdentifier
	ColumnPolicyConfigs ColumnPolicyConfigs
}

func (c DBConfig) Build() *DB {
	switch c.DBType.ID {
	case DBType_.MySQL.ID:
		c.QuotedIdentifier = QuotedIdentifier_.Backtick
		c.GetGeneratedKeyMode = GetGeneratedKeyMode_.FirstInsertId
		c.PageMode = PageMode_.LimitOffset
	case DBType_.Oracle.ID:
		c.ParamPrefix = ":"
		c.QuotedIdentifier = QuotedIdentifier_.DoubleQuote
		c.GetGeneratedKeyMode = GetGeneratedKeyMode_.Oracle
		c.PageMode = PageMode_.OffsetFetch
	case DBType_.Postgres.ID:
		c.ParamPrefix = "$"
		c.QuotedIdentifier = QuotedIdentifier_.DoubleQuote
		c.GetGeneratedKeyMode = GetGeneratedKeyMode_.InsertReturning
		c.PageMode = PageMode_.LimitOffset
	case DBType_.SQLServer.ID:
		c.QuotedIdentifier = QuotedIdentifier_.Bracket
		c.ParamPrefix = ":"
		c.GetGeneratedKeyMode = GetGeneratedKeyMode_.SQLServer
		c.PageMode = PageMode_.OffsetFetch
	case DBType_.SQLite.ID:
		c.QuotedIdentifier = QuotedIdentifier_.Backtick
		c.GetGeneratedKeyMode = GetGeneratedKeyMode_.LastInsertId
		c.PageMode = PageMode_.LimitOffset
	}
	if c.TabNameMapper == nil {
		c.TabNameMapper = defaultNameMapper
	}
	if c.ColNameMapper == nil {
		c.ColNameMapper = defaultNameMapper
	}
	return &DB{
		sqlDB:               c.SqlDB,
		logger:              c.Logger,
		sqlLogLevel:         c.SqlLogLevel,
		tabNameMapper:       c.TabNameMapper,
		colNameMapper:       c.ColNameMapper,
		paramPrefix:         c.ParamPrefix,
		GetGeneratedKeyMode: c.GetGeneratedKeyMode,
		pageMode:            c.PageMode,
		quotedIdentifier:    c.QuotedIdentifier,
		columnPolicyConfig_: c.ColumnPolicyConfigs,
	}
}

type ColumnPolicyConfigs []*columnPolicyConfig

func NewColumnPolicyConfig(column string, table_ ...string) *columnPolicyConfig {
	return &columnPolicyConfig{column: column, tableSet: newSet(table_...)}
}

type columnPolicyConfig struct {
	column         string
	tableSet       set[string]
	onInsert       *assignedPolicyConfig
	onUpdate       *assignedPolicyConfig
	onDeleteSoftly *deleteSoftlyPolicyConfig
}

func (c *columnPolicyConfig) OnInsert() *assignedPolicyConfig {
	c.onInsert = &assignedPolicyConfig{parent: c}
	return c.onInsert
}

func (c *columnPolicyConfig) OnUpdate() *assignedPolicyConfig {
	c.onUpdate = &assignedPolicyConfig{parent: c}
	return c.onUpdate
}

func (c *columnPolicyConfig) OnDeleteSoftly() *deleteSoftlyPolicyConfig {
	c.onDeleteSoftly = &deleteSoftlyPolicyConfig{parent: c}
	return c.onDeleteSoftly
}

func (c *columnPolicyConfig) UseCreateTime() *columnPolicyConfig {
	c.OnInsert().Value(false, true, func() any {
		return time.Now()
	}).OnUpdate().Never()
	return c
}

func (c *columnPolicyConfig) UseUpdateTime() *columnPolicyConfig {
	c.OnInsert().Value(false, true, func() any {
		return time.Now()
	}).OnUpdate().Value(true, true, func() any {
		return time.Now()
	})
	return c
}

func (c *columnPolicyConfig) UseRowVersion() *columnPolicyConfig {
	c.OnUpdate().RawSql(true, true, func() string {
		return c.column + " + 1"
	})
	return c
}

func (c *columnPolicyConfig) isDefault() bool {
	return len(c.tableSet) == 0
}

type assignedPolicyConfig struct {
	parent               *columnPolicyConfig
	force                bool
	reuseInBatch         bool
	never                bool
	trueRawSqlFalseValue bool
	value                func() any
	rawSql               func() string
}

func (c *assignedPolicyConfig) Never() *columnPolicyConfig {
	c.never = true
	return c.parent
}

func (c *assignedPolicyConfig) Value(force bool, reuseInBatch bool, value func() any) *columnPolicyConfig {
	c.force = force
	c.reuseInBatch = reuseInBatch
	c.value = value
	return c.parent
}

func (c *assignedPolicyConfig) RawSql(force bool, reuseInBatch bool, rawSql func() string) *columnPolicyConfig {
	c.force = force
	c.reuseInBatch = reuseInBatch
	c.trueRawSqlFalseValue = true
	c.rawSql = rawSql
	return c.parent
}

type deleteSoftlyPolicyConfig struct {
	parent      *columnPolicyConfig
	mode        deleteSoftlyMode
	normalValue any
}

func (c *deleteSoftlyPolicyConfig) PkMode[T string | int](normalValue T) *columnPolicyConfig {
	c.normalValue = normalValue
	c.mode = deleteSoftlyMode_.pk
	return c.parent
}

func (c *deleteSoftlyPolicyConfig) NullMode[T string | int](normalValue T) *columnPolicyConfig {
	c.normalValue = normalValue
	c.mode = deleteSoftlyMode_.null
	return c.parent
}

var defaultNameMapper = NewNameMapper().LowerSnakeCase()
