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
	"reflect"
)

type mapper interface {
	mapping(column_ []string, target any) (dest_ []any, after func())
}

type entityMapper struct {
	ei *entityInfo
}

func (m entityMapper) mapping(columns []string, target any) (dest_ []any, after func()) {
	v := reflect.ValueOf(target).Elem()
	dest_ = make([]any, 0, len(columns))
	toFields := make([]func(), 0, len(columns))
	for _, column := range columns {
		if index, ok := m.ei.columnToFieldIndexMap[column]; ok {
			field := v.Field(index)
			if c := getFieldConverter(field.Type()); c != nil {
				sd := c.newScanDest()
				dest_ = append(dest_, sd.dest())
				toFields = append(toFields, func() { c.toField(field, sd.value()) })
				continue
			}
			dest_ = append(dest_, field.Addr().Interface())
			continue
		}
		dest_ = append(dest_, new(any))
	}
	after = func() {
		for _, f := range toFields {
			f()
		}
	}
	return
}

type tupleMapper struct{}

func (m tupleMapper) mapping(column_ []string, target any) (dest_ []any, after func()) {
	v := reflect.ValueOf(target).Elem()
	dest_ = make([]any, 0, len(column_))
	toFields := make([]func(), 0, len(column_))
	for i := range column_ {
		if v.NumField() > i {
			field := v.Field(i)
			fieldType := field.Type()
			if c := getFieldConverter(fieldType); c != nil {
				sd := c.newScanDest()
				dest_ = append(dest_, sd.dest())
				toFields = append(toFields, func() { c.toField(field, sd.value()) })
				continue
			}
			if isValidFieldType(fieldType) || isImplementScannerValuer(fieldType) {
				dest_ = append(dest_, field.Addr().Interface())
				continue
			}
		}
		dest_ = append(dest_, new(any))
	}
	after = func() {
		for _, f := range toFields {
			f()
		}
	}
	return
}

func newMapper(ei *entityInfo, t reflect.Type) (mapper, error) {
	if ei != nil {
		return entityMapper{ei: ei}, nil
	} else if isTupleType(t) {
		return tupleMapper{}, nil
	}
	return nil, errors.New("unsupported mapping type")
}
