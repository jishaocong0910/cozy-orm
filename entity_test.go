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
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetEntity(t *testing.T) {
	r := require.New(t)
	{
		ei, err := newEntityInfo(reflect.TypeFor[int](), defaultNameMapper, defaultNameMapper, nil)
		r.Nil(ei)
		r.EqualError(err, "not a valid entity type")
	}
	{
		ei, err := newEntityInfo(reflect.TypeFor[User](), defaultNameMapper, defaultNameMapper, ColumnPolicyConfigs{
			NewColumnPolicyConfig("create_at").UseCreateTime(),
			NewColumnPolicyConfig("update_at").UseUpdateTime(),
			NewColumnPolicyConfig("version").UseRowVersion(),
			NewColumnPolicyConfig("deleted").OnDeleteSoftly().AssignedPkMode(0),
		})
		r.NoError(err)
		r.Equal("user", ei.table)
		r.Equal([]string{"id", "name", "phone", "email", "avatar_url", "status", "level", "properties", "category", "tags", "attributes",
			"uuid", "create_at", "update_at", "version", "deleted"}, ei.columns)
		r.Equal(map[string]int{"id": 0, "name": 1, "phone": 4, "email": 5, "avatar_url": 6, "status": 7, "level": 8,
			"properties": 9, "category": 10, "tags": 11, "attributes": 12, "uuid": 13, "create_at": 14, "update_at": 15,
			"version": 16, "deleted": 17},
			ei.columnToFieldIndexMap)
		r.Equal(map[string]string{"id": "Id", "name": "Name", "phone": "Phone",
			"email": "Email", "avatar_url": "AvatarUrl", "status": "Status", "level": "Level",
			"properties": "Properties", "category": "Category", "tags": "Tags", "attributes": "Attributes",
			"uuid": "Uuid", "create_at": "CreateAt", "update_at": "UpdateAt", "version": "Version", "deleted": "Deleted"},
			ei.columnToFieldNameMap)
		r.Equal([]string{"id"}, ei.pkColumns)
		r.Equal([]string{"id"}, ei.autoColumns)
		r.Equal(int64(1), ei.lastInsertIdStep)
		r.NotNil(ei.lastInsertIdConversion)

		r.Nil(ei.insertPolicy.ignoredColumnSet)
		r.Equal(newSet[string]("create_at", "update_at", "version"), ei.insertPolicy.reusedColumnSet)
		r.Nil(ei.insertPolicy.forceColumnSet)
		r.Equal(newSet[string]("create_at", "update_at", "version"), ei.insertPolicy.defaultColumnSet)
		checkMapKeys(r, []string{"create_at", "update_at", "version"}, ei.insertPolicy.assignedValueMap)
		r.False(ei.insertPolicy.assignedValueMap["create_at"].trueRawSqlFalseValue)
		r.NotNil(ei.insertPolicy.assignedValueMap["create_at"].value)
		r.Nil(ei.insertPolicy.assignedValueMap["create_at"].rawSql)
		r.False(ei.insertPolicy.assignedValueMap["update_at"].trueRawSqlFalseValue)
		r.NotNil(ei.insertPolicy.assignedValueMap["update_at"].value)
		r.Nil(ei.insertPolicy.assignedValueMap["update_at"].rawSql)

		r.Equal(newSet[string]("create_at", "id"), ei.updatePolicy.ignoredColumnSet)
		r.Equal(newSet[string]("update_at", "version"), ei.updatePolicy.reusedColumnSet)
		r.Equal(newSet[string]("update_at", "version"), ei.updatePolicy.forceColumnSet)
		r.Nil(ei.updatePolicy.defaultColumnSet)
		checkMapKeys(r, []string{"update_at", "version"}, ei.updatePolicy.assignedValueMap)
		r.False(ei.updatePolicy.assignedValueMap["update_at"].trueRawSqlFalseValue)
		r.NotNil(ei.updatePolicy.assignedValueMap["update_at"].value)
		r.Nil(ei.updatePolicy.assignedValueMap["update_at"].rawSql)
		r.True(ei.updatePolicy.assignedValueMap["version"].trueRawSqlFalseValue)
		r.Nil(ei.updatePolicy.assignedValueMap["version"].value)
		r.NotNil(ei.updatePolicy.assignedValueMap["version"].rawSql)

		r.Equal(deleteSoftlyMode_.assignedPk, ei.deleteSoftlyPolicy.mode)
		r.Equal("deleted", ei.deleteSoftlyPolicy.deletedColumn)
		r.Equal("id", ei.deleteSoftlyPolicy.pkColumn)
		r.Equal(0, ei.deleteSoftlyPolicy.normalValue)
	}
	{
		ei, err := newEntityInfo(reflect.TypeFor[DemoIgnoreField](), defaultNameMapper, defaultNameMapper, nil)
		r.NoError(err)
		r.Equal("demo_ignore_field", ei.table)
		r.Equal([]string{"id"}, ei.columns)
		r.Len(ei.columnToFieldIndexMap, 1)
		r.Equal(map[string]int{"id": 0}, ei.columnToFieldIndexMap)
		r.Equal(map[string]string{"id": "Id"}, ei.columnToFieldNameMap)
		r.Empty(ei.pkColumns)
		r.Equal([]string{"id"}, ei.autoColumns)
		r.Equal(int64(2), ei.lastInsertIdStep)
		r.NotNil(ei.lastInsertIdConversion)
	}
	{
		ei, err := newEntityInfo(reflect.TypeFor[DemoEntityTag](), defaultNameMapper, defaultNameMapper, nil)
		r.NoError(err)
		r.Equal("demo", ei.table)
	}
	{
		ei, err := newEntityInfo(reflect.TypeFor[DemoMulPk](), defaultNameMapper, defaultNameMapper, nil)
		r.NoError(err)
		r.Equal("demo_mul_pk", ei.table)
		r.Equal([]string{"id1", "id2"}, ei.columns)
		r.Equal(map[string]int{"id1": 0, "id2": 1}, ei.columnToFieldIndexMap)
		r.Equal(map[string]string{"id1": "Id1", "id2": "Id2"}, ei.columnToFieldNameMap)
		r.Equal([]string{"id1", "id2"}, ei.pkColumns)
		r.Equal([]string{"id1", "id2"}, ei.autoColumns)
		r.Equal(int64(0), ei.lastInsertIdStep)
		r.Nil(ei.lastInsertIdConversion)
	}
	{
		c1 := NewColumnPolicyConfig("name").ForTable("product").OnInsert().Value(false, false, func(context.Context) any { return nil })
		c2 := NewColumnPolicyConfig("field").OnInsert().Value(false, false, func(context.Context) any { return nil })
		c3 := NewColumnPolicyConfig("name").ForTable("user").OnInsert().Value(false, false, func(context.Context) any { return nil }).
			OnUpdate().Value(false, false, func(context.Context) any { return nil })
		c4 := NewColumnPolicyConfig("name").OnInsert().Value(false, false, func(context.Context) any { return nil }).
			OnUpdate().Value(false, false, func(context.Context) any { return nil })
		c5 := NewColumnPolicyConfig("level").ForTable("user").OnInsert().Value(false, false, func(context.Context) any { return nil }).
			OnUpdate().Value(false, false, func(context.Context) any { return nil })
		c6 := NewColumnPolicyConfig("level").ForTable("user").OnInsert().Value(false, false, func(context.Context) any { return nil }).
			OnUpdate().Value(false, false, func(context.Context) any { return nil })
		c7 := NewColumnPolicyConfig("create_at").OnInsert().Value(false, false, func(context.Context) any { return nil }).
			OnUpdate().Value(false, false, func(context.Context) any { return nil })
		c8 := NewColumnPolicyConfig("update").IgnoreTable("user").OnInsert().Value(false, false, func(context.Context) any { return nil }).
			OnUpdate().Value(false, false, func(context.Context) any { return nil })
		ei, err := newEntityInfo(reflect.TypeFor[User](), defaultNameMapper, defaultNameMapper, ColumnPolicyConfigs{c1, c2, c3, c4, c5, c6, c7, c8})
		r.NoError(err)
		checkEqualFunc(r, c3.onInsert.value, ei.insertPolicy.assignedValueMap["name"].value)
		checkEqualFunc(r, c3.onUpdate.value, ei.updatePolicy.assignedValueMap["name"].value)
		checkEqualFunc(r, c6.onInsert.value, ei.insertPolicy.assignedValueMap["level"].value)
		checkEqualFunc(r, c6.onUpdate.value, ei.updatePolicy.assignedValueMap["level"].value)
		checkEqualFunc(r, c7.onInsert.value, ei.insertPolicy.assignedValueMap["create_at"].value)
		checkEqualFunc(r, c7.onUpdate.value, ei.updatePolicy.assignedValueMap["create_at"].value)
		checkNoKey(r, "update_at", ei.insertPolicy.assignedValueMap)
		checkNoKey(r, "update_at", ei.updatePolicy.assignedValueMap)
	}
	{
		c1 := NewColumnPolicyConfig("deleted").ForTable("user").OnDeleteSoftly().AssignedNullMode(0)
		c2 := NewColumnPolicyConfig("deleted").OnDeleteSoftly().AssignedPkMode(0)
		ei, err := newEntityInfo(reflect.TypeFor[User](), defaultNameMapper, defaultNameMapper, ColumnPolicyConfigs{c1, c2})
		r.NoError(err)
		r.Equal(deleteSoftlyMode_.assignedNull, ei.deleteSoftlyPolicy.mode)

		c1 = NewColumnPolicyConfig("deleted").ForTable("user").OnDeleteSoftly().AssignedNullMode(0)
		c2 = NewColumnPolicyConfig("deleted").ForTable("user").OnDeleteSoftly().AssignedPkMode(0)
		ei, err = newEntityInfo(reflect.TypeFor[User](), defaultNameMapper, defaultNameMapper, ColumnPolicyConfigs{c1, c2})
		r.NoError(err)
		r.Equal(deleteSoftlyMode_.assignedPk, ei.deleteSoftlyPolicy.mode)
	}
}

func TestEntity_getOnDemandColumnSet(t *testing.T) {
	r := require.New(t)
	ei, err := newEntityInfo(reflect.TypeFor[User](), defaultNameMapper, defaultNameMapper, nil)
	r.NoError(err)
	{
		r.Nil(ei._getOnDemandColumnSet(nil))
		r.Nil(ei._getOnDemandColumnSet(&OnDemand{}))
		r.Nil(ei._getOnDemandColumnSet(DemandFor[string]()))
		r.Equal(newSet("name", "phone", "email", "level"), ei._getOnDemandColumnSet(DemandFor[UserSimple]()))
		r.NotNil(ei._getOnDemandColumnSet(DemandFor[DemoDemand]()))
		r.Equal(newSet("name", "phone", "email"), ei._getOnDemandColumnSet(DemandFor[DemoDemand]()))
	}
}

func TestEntity_getColumns(t *testing.T) {
	r := require.New(t)
	ei, err := newEntityInfo(reflect.TypeFor[User](), defaultNameMapper, defaultNameMapper, nil)
	r.NoError(err)
	{
		columns := ei.getColumns(&User{
			Id:      new(int64(1)),
			Name:    new("name"),
			Phone:   new("phone"),
			Email:   new("email"),
			Deleted: new(int64(1)),
		}, nil, newSet("status", "level"), newSet("name", "level"))
		r.Equal([]string{"id", "phone", "email", "status", "deleted"}, columns)
	}
	{
		columns := ei.getColumns(nil, DemandFor[UserSimple](), newSet("status", "tags"), newSet("tags", "level"))
		r.Equal([]string{"name", "phone", "email", "status"}, columns)
	}
	{
		columns := ei.getColumns(nil, DemandFor[int](), newSet("status", "tags"), newSet("tags", "level"))
		r.Equal([]string{"status"}, columns)
	}
}

func TestEntity_getValueMap(t *testing.T) {
	r := require.New(t)
	ei, err := newEntityInfo(reflect.TypeFor[User](), defaultNameMapper, defaultNameMapper, nil)
	r.NoError(err)
	r.Equal(map[string]any{
		"name":   new("name"),
		"status": new(UserStatus(1)),
	}, ei.getValueMap(&User{Name: new("name"), Status: new(UserStatus(1))}, []string{"name", "status", "level"}))
	r.Nil(nil, ei.getValueMap(nil, []string{"name", "status", "level"}))
	r.Nil((*User)(nil), ei.getValueMap(nil, []string{"name", "status", "level"}))
}

func TestLastInsertIdConversion(t *testing.T) {
	r := require.New(t)
	id := int64(123)
	r.Equal(123, convertId[*int](id))
	r.Equal(int8(123), convertId[*int8](id))
	r.Equal(int16(123), convertId[*int16](id))
	r.Equal(int32(123), convertId[*int32](id))
	r.Equal(int64(123), convertId[*int64](id))
	r.Equal(uint(123), convertId[*uint](id))
	r.Equal(uint8(123), convertId[*uint8](id))
	r.Equal(uint16(123), convertId[*uint16](id))
	r.Equal(uint32(123), convertId[*uint32](id))
	r.Equal(uint64(123), convertId[*uint64](id))
	r.Equal(float32(123), convertId[*float32](id))
	r.Equal(float64(123), convertId[*float64](id))
	r.Equal("123", convertId[*string](id))
}
