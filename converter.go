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
	"reflect"
	"sync"
)

type Converter[T any] interface {
	ToArg() T
	ToField(src T)
}

const (
	toArgMethodName   = "ToArg"
	toFieldMethodName = "ToField"
)

var (
	argConverters              sync.Map
	registerArgConverterLock   sync.Mutex
	fieldConverters            sync.Map
	registerFieldConverterLock sync.Mutex
)

func getArgConverter(t reflect.Type) *argConverter {
	if val, ok := argConverters.Load(t); ok {
		ac, _ := val.(*argConverter)
		return ac
	}
	return registerArgConverter(t)
}

func registerArgConverter(t reflect.Type) *argConverter {
	registerArgConverterLock.Lock()
	defer registerArgConverterLock.Unlock()

	if val, ok := argConverters.Load(t); ok { // coverage-ignore
		ac, _ := val.(*argConverter)
		return ac
	}

	if !isImplementConverter(t) {
		argConverters.Store(t, nil)
		return nil
	}

	vt := t
	if vt.Kind() == reflect.Pointer {
		vt = t.Elem()
	}
	_, isValueReceiver := vt.MethodByName(toArgMethodName)
	ac := &argConverter{isPtrReceiver: !isValueReceiver}

	argConverters.Store(t, ac)
	return ac
}

type argConverter struct {
	isPtrReceiver bool
}

func (c *argConverter) convert(v reflect.Value) any {
	if c.isPtrReceiver && v.Kind() != reflect.Pointer {
		pv := reflect.New(v.Type())
		pv.Elem().Set(v)
		v = pv
	}
	return v.MethodByName(toArgMethodName).Call(nil)[0].Interface()
}

func getFieldConverter(t reflect.Type) *fieldConverter {
	if val, ok := fieldConverters.Load(t); ok {
		fc, _ := val.(*fieldConverter)
		return fc
	}
	return registerFieldConverter(t)
}

func registerFieldConverter(t reflect.Type) *fieldConverter {
	registerFieldConverterLock.Lock()
	defer registerFieldConverterLock.Unlock()

	if val, ok := fieldConverters.Load(t); ok { // coverage-ignore
		fc, _ := val.(*fieldConverter)
		return fc
	}

	var fc *fieldConverter
	if isImplementConverter(t) {
		pt := t
		if pt.Kind() != reflect.Pointer {
			pt = reflect.PointerTo(t)
		}
		method, _ := pt.MethodByName(toFieldMethodName)
		mediumType := method.Type.In(1)
		fc = &fieldConverter{
			ptrMediumType: reflect.PointerTo(mediumType),
			fieldKind:     t.Kind(),
		}
	}

	fieldConverters.Store(t, fc)
	return fc
}

type fieldConverter struct {
	ptrMediumType reflect.Type
	fieldKind     reflect.Kind
}

func (c *fieldConverter) convert(field reflect.Value, value any) {
	if value == nil {
		return
	}
	switch c.fieldKind {
	case reflect.Pointer:
		field.Set(reflect.New(field.Type().Elem()))
		field.MethodByName(toFieldMethodName).Call([]reflect.Value{reflect.ValueOf(value)})
	case reflect.Slice, reflect.Map:
		field.Set(reflect.New(field.Type()).Elem())
		fallthrough
	default:
		field.Addr().MethodByName(toFieldMethodName).Call([]reflect.Value{reflect.ValueOf(value)})
	}
}

func convertArgs(args []any) []any {
	for i, a := range args {
		if a != nil {
			v := reflect.ValueOf(a)
			if isBaseKind(v.Kind()) && v.IsNil() {
				continue
			}
			if c := getArgConverter(v.Type()); c != nil {
				args[i] = c.convert(v)
			}
		}
	}
	return args
}
