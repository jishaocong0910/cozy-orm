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
	mapping(columns []string, target any) (dest []any, afterScan func())
}

type entityMapper struct {
	ei *entityInfo
}

func (m entityMapper) mapping(columns []string, target any) (dest []any, afterScan func()) {
	v := reflect.ValueOf(target).Elem()
	dest = make([]any, 0, len(columns))
	conv := make([]func(), 0, len(columns))
	for _, column := range columns {
		if index, ok := m.ei.columnToFieldIndexMap[column]; ok {
			field := v.Field(index)
			if c := getFieldConverter(field.Type()); c != nil {
				medium := newMappingMedium(c.ptrMediumType)
				dest = append(dest, medium.dest())
				conv = append(conv, func() { c.convert(field, medium.value()) })
				continue
			}
			dest = append(dest, field.Addr().Interface())
			continue
		}
		dest = append(dest, new(any))
	}
	afterScan = func() {
		for _, f := range conv {
			f()
		}
	}
	return
}

type tupleMapper struct{}

func (m tupleMapper) mapping(columns []string, target any) (dest []any, afterScan func()) {
	v := reflect.ValueOf(target).Elem()
	dest = make([]any, 0, len(columns))
	conv := make([]func(), 0, len(columns))
	for i := range columns {
		if v.NumField() > i {
			field := v.Field(i)
			fieldType := field.Type()
			if c := getFieldConverter(fieldType); c != nil {
				medium := newMappingMedium(c.ptrMediumType)
				dest = append(dest, medium.dest())
				conv = append(conv, func() { c.convert(field, medium.value()) })
				continue
			}
			if isBaseValueType(fieldType) {
				medium := newMappingMedium(reflect.PointerTo(fieldType))
				dest = append(dest, medium.dest())
				conv = append(conv, func() {
					if value := medium.value(); value != nil {
						field.Addr().Elem().Set(reflect.ValueOf(value))
					}
				})
				continue
			}
			if isValidFieldType(fieldType) || isImplementScannerValuer(fieldType) {
				dest = append(dest, field.Addr().Interface())
				continue
			}
		}
		dest = append(dest, new(any))
	}
	afterScan = func() {
		for _, f := range conv {
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

type mappingMedium struct {
	ptrToPtr reflect.Value
}

func (d mappingMedium) dest() any {
	return d.ptrToPtr.Interface()
}

func (d mappingMedium) value() any {
	if ptr := d.ptrToPtr.Elem(); !ptr.IsNil() {
		return ptr.Elem().Interface()
	}
	return nil
}

func newMappingMedium(t reflect.Type) mappingMedium {
	return mappingMedium{ptrToPtr: reflect.New(t)}
}
