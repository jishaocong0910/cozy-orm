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
	"errors"
	"maps"
	"reflect"
	"strconv"
	"sync"
)

func newEntityInfo(t reflect.Type, tableNameMapper, columnNameMapper *NameMapper, columnPolicyConfig_ []*columnPolicyConfig) (*entityInfo, error) {
	if !isEntityType(t) {
		return nil, errors.New("not a valid entity type")
	}

	ei := &entityInfo{
		typ:                   t,
		columnToFieldIndexMap: make(map[string]int, t.NumField()),
		columnToFieldNameMap:  make(map[string]string, t.NumField()),
		fieldToColumnMap:      make(map[string]string, t.NumField()),
	}

	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		if tf.Anonymous {
			continue
		}
		if tf.Name == "_" {
			if i == 0 && tf.Type.Kind() == reflect.Struct && tf.Type.NumField() == 0 {
				et := parseEntityTag(tf)
				ei.table = et.table
			}
			continue
		}
		if !tf.IsExported() {
			continue
		}
		tag := parseFieldTag(tf)
		if tag.ignore {
			continue
		}
		if !isValidFieldType(tf.Type) {
			continue
		}
		ei._registerField(tf, tag, columnNameMapper)
	}
	if ei.table == "" {
		ei.table = tableNameMapper.Convert(t.Name())
	}

	ei._registerPolicy(columnPolicyConfig_)
	return ei, nil
}

type entityInfo struct {
	typ                    reflect.Type
	table                  string
	column_                []string
	columnToFieldIndexMap  map[string]int
	columnToFieldNameMap   map[string]string
	fieldToColumnMap       map[string]string
	pkColumn_              []string
	autoColumn_            []string
	lastInsertIdStep       int64
	lastInsertIdConversion func(id int64) reflect.Value
	insertPolicy           assignedPolicy
	updatePolicy           assignedPolicy
	deleteSoftlyPolicy     deleteSoftlyPolicy

	onDemandColumns             sync.Map
	registerOnDemandColumnsLock sync.Mutex
}

func (e *entityInfo) getColumns(entity any, onDemand *OnDemand, requiredSet set[string], ignoredSet set[string]) []string {
	var column_ []string
	var ev reflect.Value
	onDemandColumnSet := e._getOnDemandColumnSet(onDemand)

	if onDemandColumnSet == nil {
		column_ = make([]string, 0, len(e.column_))
		if entity != nil {
			ev = reflect.ValueOf(entity).Elem()
		}
	} else {
		column_ = make([]string, 0, len(onDemandColumnSet))
	}

	for _, column := range e.column_ {
		if ignoredSet.contain(column) {
			continue
		}
		if requiredSet.contain(column) {
			column_ = append(column_, column)
			continue
		}
		if onDemandColumnSet == nil {
			if !ev.IsValid() {
				continue
			}
			if !ev.Field(e.columnToFieldIndexMap[column]).IsNil() {
				column_ = append(column_, column)
				continue
			}
		} else if onDemandColumnSet.contain(column) {
			column_ = append(column_, column)
			continue
		}
	}
	return column_
}

func (e *entityInfo) _getOnDemandColumnSet(onDemand *OnDemand) set[string] {
	if onDemand == nil || onDemand.t == nil {
		return nil
	}
	if val, ok := e.onDemandColumns.Load(onDemand.t); ok {
		columnSet, _ := val.(set[string])
		return columnSet
	}

	e.registerOnDemandColumnsLock.Lock()
	defer e.registerOnDemandColumnsLock.Unlock()

	if val, ok := e.onDemandColumns.Load(onDemand.t); ok { // coverage-ignore
		columnSet, _ := val.(set[string])
		return columnSet
	}

	if onDemand.t.Kind() != reflect.Struct {
		e.onDemandColumns.Store(onDemand.t, nil)
		return nil
	}

	columns := make([]string, 0, onDemand.t.NumField())
	for tf := range onDemand.t.Fields() {
		if tf.Anonymous {
			continue
		}
		if c, ok := e.fieldToColumnMap[tf.Name]; ok {
			columns = append(columns, c)
		}
	}

	columnSet := newSet(columns...)
	e.onDemandColumns.Store(onDemand.t, columnSet)
	return columnSet
}

func (e *entityInfo) getValueMap(entity any, column__ ...[]string) map[string]any {
	var v reflect.Value
	if entity != nil {
		v = reflect.ValueOf(entity).Elem()
	}
	if !v.IsValid() {
		return nil
	}
	var valueMap map[string]any
	size := 0
	for _, column_ := range column__ {
		size += len(column_)
	}
	valueMap = make(map[string]any, size)
	for _, column_ := range column__ {
		for _, column := range column_ {
			if i, ok := e.columnToFieldIndexMap[column]; ok {
				vf := v.Field(i)
				if !vf.IsNil() {
					valueMap[column] = vf.Interface()
				}
			}
		}
	}
	return valueMap
}

func (e *entityInfo) _registerField(tf reflect.StructField, tag fieldTag, columnNameMapper *NameMapper) {
	column := tag.column
	if column == "" {
		column = columnNameMapper.Convert(tf.Name)
	}
	e.column_ = append(e.column_, column)
	e.columnToFieldIndexMap[column] = tf.Index[0]
	e.columnToFieldNameMap[column] = tf.Name
	e.fieldToColumnMap[tf.Name] = column
	if tag.pk {
		e.pkColumn_ = append(e.pkColumn_, column)
	}
	if tag.auto > 0 {
		e.autoColumn_ = append(e.autoColumn_, column)
		if len(e.autoColumn_) > 1 {
			e.lastInsertIdStep = 0
			e.lastInsertIdConversion = nil
		} else if a, ok := lastInsertIdConversionMap[tf.Type]; ok {
			e.lastInsertIdStep = tag.auto
			e.lastInsertIdConversion = a
		}
	}
}

func (e *entityInfo) _registerPolicy(columnPolicyConfig_ []*columnPolicyConfig) {
	for _, pk := range e.pkColumn_ {
		columnPolicyConfig_ = append(columnPolicyConfig_, NewColumnPolicyConfig(pk, e.table).OnUpdate().Never())
	}
	columnSet := newSet(e.column_...)
	columnOnInsertMap := map[string]*assignedPolicyConfig{}
	columnOnUpdateMap := map[string]*assignedPolicyConfig{}
	var finalDeleteSoftlyPolicyConfig *deleteSoftlyPolicyConfig
	for _, config := range columnPolicyConfig_ {
		if !config.isDefault() && !config.tableSet.contain(e.table) {
			continue
		}
		if !columnSet.contain(config.column) {
			continue
		}
		if config.onInsert != nil {
			if exists, ok := columnOnInsertMap[config.column]; !ok || exists.parent.isDefault() || !config.isDefault() {
				columnOnInsertMap[config.column] = config.onInsert
			}
		}
		if config.onUpdate != nil {
			if exists, ok := columnOnUpdateMap[config.column]; !ok || exists.parent.isDefault() || !config.isDefault() {
				columnOnUpdateMap[config.column] = config.onUpdate
			}
		}
		if config.onDeleteSoftly != nil {
			if finalDeleteSoftlyPolicyConfig == nil || finalDeleteSoftlyPolicyConfig.parent.isDefault() || !config.isDefault() {
				finalDeleteSoftlyPolicyConfig = config.onDeleteSoftly
			}
		}
	}
	e.insertPolicy.loadConfig(columnOnInsertMap)
	e.updatePolicy.loadConfig(columnOnUpdateMap)
	e.deleteSoftlyPolicy.loadConfig(finalDeleteSoftlyPolicyConfig, e.pkColumn_)
}

type assignedPolicy struct {
	ignoredColumnSet set[string]
	reusedColumnSet  set[string]
	forceColumnSet   set[string]
	defaultColumnSet set[string]
	assignedValueMap map[string]assignedValuePolicy
}

func (p *assignedPolicy) loadConfig(columnOnInsertMap map[string]*assignedPolicyConfig) {
	if len(columnOnInsertMap) == 0 {
		return
	}
	ignoredColumnSet := newSetWithCap[string](len(columnOnInsertMap))
	reusedColumnSet := newSetWithCap[string](len(columnOnInsertMap))
	forceColumnSet := newSetWithCap[string](len(columnOnInsertMap))
	defaultColumnSet := newSetWithCap[string](len(columnOnInsertMap))
	assignedValueMap := make(map[string]assignedValuePolicy, len(columnOnInsertMap))
	for column, config := range columnOnInsertMap {
		if config.never {
			ignoredColumnSet.add(column)
			continue
		}
		if config.batchReuse {
			reusedColumnSet.add(column)
		}
		if config.force {
			forceColumnSet.add(column)
		} else {
			defaultColumnSet.add(column)
		}
		assignedValueMap[column] = assignedValuePolicy{
			trueRawSqlFalseValue: config.trueRawSqlFalseValue,
			value:                config.value,
			rawSql:               config.rawSql,
		}
	}
	if len(ignoredColumnSet) > 0 {
		p.ignoredColumnSet = newSetWithCap[string](len(ignoredColumnSet))
		maps.Copy(p.ignoredColumnSet, ignoredColumnSet)
	}
	if len(reusedColumnSet) > 0 {
		p.reusedColumnSet = newSetWithCap[string](len(reusedColumnSet))
		maps.Copy(p.reusedColumnSet, reusedColumnSet)
	}
	if len(forceColumnSet) > 0 {
		p.forceColumnSet = newSetWithCap[string](len(forceColumnSet))
		maps.Copy(p.forceColumnSet, forceColumnSet)
	}
	if len(defaultColumnSet) > 0 {
		p.defaultColumnSet = newSetWithCap[string](len(defaultColumnSet))
		maps.Copy(p.defaultColumnSet, defaultColumnSet)
	}
	if len(assignedValueMap) > 0 {
		p.assignedValueMap = make(map[string]assignedValuePolicy, len(assignedValueMap))
		maps.Copy(p.assignedValueMap, assignedValueMap)
	}
}

type assignedValuePolicy struct {
	trueRawSqlFalseValue bool
	value                func() any
	rawSql               func() string
}

type deleteSoftlyPolicy struct {
	mod           deleteSoftlyMode
	deletedColumn string
	pkColumn      string
	normalValue   any
}

func (p *deleteSoftlyPolicy) loadConfig(config *deleteSoftlyPolicyConfig, pkColumn_ []string) {
	if config == nil || config.mode.IsUndefined() || config.normalValue == nil || config.mode.Is(deleteSoftlyMode_.pk) && len(pkColumn_) != 1 {
		return
	}
	p.mod = config.mode
	p.deletedColumn = config.parent.column
	p.pkColumn = pkColumn_[0]
	p.normalValue = config.normalValue
}

var (
	lastInsertIdConversionMap = map[reflect.Type]func(id int64) reflect.Value{
		reflect.TypeFor[*int]():     func(id int64) reflect.Value { return reflect.ValueOf(new(int(id))) },
		reflect.TypeFor[*int8]():    func(id int64) reflect.Value { return reflect.ValueOf(new(int8(id))) },
		reflect.TypeFor[*int16]():   func(id int64) reflect.Value { return reflect.ValueOf(new(int16(id))) },
		reflect.TypeFor[*int32]():   func(id int64) reflect.Value { return reflect.ValueOf(new(int32(id))) },
		reflect.TypeFor[*int64]():   func(id int64) reflect.Value { return reflect.ValueOf(new(id)) },
		reflect.TypeFor[*uint]():    func(id int64) reflect.Value { return reflect.ValueOf(new(uint(id))) },
		reflect.TypeFor[*uint8]():   func(id int64) reflect.Value { return reflect.ValueOf(new(uint8(id))) },
		reflect.TypeFor[*uint16]():  func(id int64) reflect.Value { return reflect.ValueOf(new(uint16(id))) },
		reflect.TypeFor[*uint32]():  func(id int64) reflect.Value { return reflect.ValueOf(new(uint32(id))) },
		reflect.TypeFor[*uint64]():  func(id int64) reflect.Value { return reflect.ValueOf(new(uint64(id))) },
		reflect.TypeFor[*float32](): func(id int64) reflect.Value { return reflect.ValueOf(new(float32(id))) },
		reflect.TypeFor[*float64](): func(id int64) reflect.Value { return reflect.ValueOf(new(float64(id))) },
		reflect.TypeFor[*string]():  func(id int64) reflect.Value { return reflect.ValueOf(new(strconv.FormatInt(id, 10))) },
	}
)
