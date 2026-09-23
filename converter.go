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
	valueConverters            sync.Map
	registerValueConverterLock sync.Mutex
	fieldConverters            sync.Map
	registerFieldConverterLock sync.Mutex
)

func getValueConverter(t reflect.Type) valueConverter {
	if val, ok := valueConverters.Load(t); ok {
		vc, _ := val.(valueConverter)
		return vc
	}
	return registerValueConverter(t)
}

func registerValueConverter(t reflect.Type) valueConverter {
	registerValueConverterLock.Lock()
	defer registerValueConverterLock.Unlock()

	if val, ok := valueConverters.Load(t); ok { // coverage-ignore
		vc, _ := val.(valueConverter)
		return vc
	}

	if !isImplementConverter(t) {
		valueConverters.Store(t, nil)
		return nil
	}

	vt := t
	if vt.Kind() == reflect.Pointer {
		vt = t.Elem()
	}
	_, valueReceiver := vt.MethodByName(toArgMethodName)

	var vc valueConverter
	if valueReceiver {
		vc = vrValueConverter{}
	} else {
		vc = prValueConverter{}
	}
	valueConverters.Store(t, vc)
	return vc
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
		indirectType := method.Type.In(1)
		switch t.Kind() {
		case reflect.Pointer:
			fc = ptrFieldConverter{ptrType: reflect.PointerTo(indirectType)}
		case reflect.Slice:
			fc = sliceFieldConverter{ptrType: reflect.PointerTo(indirectType)}
		case reflect.Map:
			fc = mapFieldConverter{ptrType: reflect.PointerTo(indirectType)}
		default:
			fc = valueFieldConverter{ptrType: reflect.PointerTo(indirectType)}
		}
	} else if isBaseValueType(t) {
		fc = zeroFieldConverter{ptrType: reflect.PointerTo(t)}
	}

	fieldConverters.Store(t, fc)
	return fc
}

type valueConverter interface {
	toValue(v reflect.Value) any
}

type fieldConverter interface {
	newScanDest() scanDest
	toField(field reflect.Value, value any)
}

type vrValueConverter struct{}

func (c vrValueConverter) toValue(v reflect.Value) any {
	return v.MethodByName(toArgMethodName).Call(nil)[0].Interface()
}

type prValueConverter struct{}

func (c prValueConverter) toValue(v reflect.Value) any {
	if v.Kind() != reflect.Pointer {
		pv := reflect.New(v.Type())
		pv.Elem().Set(v)
		v = pv
	}
	return v.MethodByName(toArgMethodName).Call(nil)[0].Interface()
}

type ptrFieldConverter struct {
	ptrType reflect.Type
}

func (c ptrFieldConverter) newScanDest() scanDest {
	return scanDest{p2pValue: reflect.New(c.ptrType)}
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

func (c sliceFieldConverter) newScanDest() scanDest {
	return scanDest{p2pValue: reflect.New(c.ptrType)}
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

func (c mapFieldConverter) newScanDest() scanDest {
	return scanDest{p2pValue: reflect.New(c.ptrType)}
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

func (c valueFieldConverter) newScanDest() scanDest {
	return scanDest{p2pValue: reflect.New(c.ptrType)}
}

func (c valueFieldConverter) toField(field reflect.Value, value any) {
	if value != nil {
		field.Addr().MethodByName(toFieldMethodName).Call([]reflect.Value{reflect.ValueOf(value)})
	}
}

type zeroFieldConverter struct {
	ptrType reflect.Type
}

func (c zeroFieldConverter) newScanDest() scanDest {
	return scanDest{p2pValue: reflect.New(c.ptrType)}
}

func (c zeroFieldConverter) toField(field reflect.Value, value any) {
	if value != nil {
		field.Addr().Elem().Set(reflect.ValueOf(value))
	}
}

type scanDest struct {
	p2pValue reflect.Value
}

func (d scanDest) dest() any {
	return d.p2pValue.Interface()
}

func (d scanDest) value() any {
	if ptr := d.p2pValue.Elem(); !ptr.IsNil() {
		return ptr.Elem().Interface()
	}
	return nil
}

func convertArgs(arg_ []any) []any {
	for i, a := range arg_ {
		if a != nil {
			v := reflect.ValueOf(a)
			if isBaseKind(v.Kind()) && v.IsNil() {
				continue
			}
			if c := getValueConverter(v.Type()); c != nil {
				arg_[i] = c.toValue(v)
			}
		}
	}
	return arg_
}
