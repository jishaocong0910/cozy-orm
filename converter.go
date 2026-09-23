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

type argConverter struct {
	isPtrReceiver bool
}

func (c *argConverter) toArg(v reflect.Value) any {
	if c.isPtrReceiver && v.Kind() != reflect.Pointer {
		pv := reflect.New(v.Type())
		pv.Elem().Set(v)
		v = pv
	}
	return v.MethodByName(toArgMethodName).Call(nil)[0].Interface()
}

func convertArgs(arg_ []any) []any {
	for i, a := range arg_ {
		if a != nil {
			v := reflect.ValueOf(a)
			if isBaseKind(v.Kind()) && v.IsNil() {
				continue
			}
			if c := getArgConverter(v.Type()); c != nil {
				arg_[i] = c.toArg(v)
			}
		}
	}
	return arg_
}

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

type fieldConverter struct {
	ptrType reflect.Type
}

func (c *fieldConverter) toField(field reflect.Value, value any) {

}

func getFieldConverter(t reflect.Type) fieldConverter {
	if val, ok := fieldConverters.Load(t); ok {
		fc, _ := val.(fieldConverter)
		return fc
	}
	return registerFieldConverter(t)
}

func registerFieldConverter(t reflect.Type) fieldConverter {
	registerFieldConverterLock.Lock()
	defer registerFieldConverterLock.Unlock()

	if val, ok := fieldConverters.Load(t); ok { // coverage-ignore
		fc, _ := val.(fieldConverter)
		return fc
	}

	var fc fieldConverter
	if isImplementConverter(t) {
		pt := t
		if pt.Kind() != reflect.Pointer {
			pt = reflect.PointerTo(t)
		}
		method, _ := pt.MethodByName(toFieldMethodName)
		mediumType := method.Type.In(1)
		switch t.Kind() {
		case reflect.Pointer:
			fc = ptrFieldConverter{ptrMediumType: reflect.PointerTo(mediumType)}
		case reflect.Slice:
			fc = sliceFieldConverter{ptrType: reflect.PointerTo(mediumType)}
		case reflect.Map:
			fc = mapFieldConverter{ptrType: reflect.PointerTo(mediumType)}
		default:
			fc = valueFieldConverter{ptrType: reflect.PointerTo(mediumType)}
		}
	} else if isBaseValueType(t) {
		fc = zeroFieldConverter{ptrType: reflect.PointerTo(t)}
	}

	fieldConverters.Store(t, fc)
	return fc
}

type ptrFieldConverter struct {
	ptrMediumType reflect.Type
}

func (c ptrFieldConverter) toField(field reflect.Value, value any) {
	if value != nil {
		field.Set(reflect.New(field.Type().Elem()))
		field.MethodByName(toFieldMethodName).Call([]reflect.Value{reflect.ValueOf(value)})
	}
}

type sliceFieldConverter struct {
	ptrType reflect.Type
}

func (c sliceFieldConverter) newScanDest() mappingMedium {
	return mappingMedium{p2pValue: reflect.New(c.ptrType)}
}

func (c sliceFieldConverter) toField(field reflect.Value, value any) {
	if value != nil {
		field.Set(reflect.New(field.Type()).Elem())
		field.Addr().MethodByName(toFieldMethodName).Call([]reflect.Value{reflect.ValueOf(value)})
	}
}

type mapFieldConverter struct {
	ptrType reflect.Type
}

func (c mapFieldConverter) newScanDest() mappingMedium {
	return mappingMedium{p2pValue: reflect.New(c.ptrType)}
}

func (c mapFieldConverter) toField(field reflect.Value, value any) {
	if value != nil {
		field.Set(reflect.New(field.Type()).Elem())
		field.Addr().MethodByName(toFieldMethodName).Call([]reflect.Value{reflect.ValueOf(value)})
	}
}

type valueFieldConverter struct {
	ptrType reflect.Type
}

func (c valueFieldConverter) newScanDest() mappingMedium {
	return mappingMedium{p2pValue: reflect.New(c.ptrType)}
}

func (c valueFieldConverter) toField(field reflect.Value, value any) {
	if value != nil {
		field.Addr().MethodByName(toFieldMethodName).Call([]reflect.Value{reflect.ValueOf(value)})
	}
}

type zeroFieldConverter struct {
	ptrType reflect.Type
}

func (c zeroFieldConverter) newScanDest() mappingMedium {
	return mappingMedium{p2pValue: reflect.New(c.ptrType)}
}

func (c zeroFieldConverter) toField(field reflect.Value, value any) {
	if value != nil {
		field.Addr().Elem().Set(reflect.ValueOf(value))
	}
}
