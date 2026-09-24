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
	"database/sql"
	"database/sql/driver"
	"reflect"
	"strings"
	"time"
	_ "unsafe"
)

var isValidKind = func() func(t reflect.Kind) bool {
	s := newSet(
		reflect.Pointer,
		reflect.Slice,
		reflect.Map,
	)
	return func(k reflect.Kind) bool {
		return s.contain(k)
	}
}()

func isValidBuiltinType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Pointer:
		if isBaseValueType(t.Elem()) {
			return true
		}
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return true
		}
	}
	return false
}

//go:linkname isValidFieldType github.com/jishaocong0910/cozy-orm/orm.isValidFieldType
func isValidFieldType(t reflect.Type) bool {
	if !isValidKind(t.Kind()) {
		return false
	}
	return isValidBuiltinType(t) || isImplementConverter(t) || isImplementScannerValuer(t)
}

func isBaseValueType(t reflect.Type) bool {
	if isBaseScalarKind(t) || isBaseStructType(t) || isBaseArrayType(t) {
		return true
	}
	return false
}

var isBaseScalarKind = func() func(t reflect.Type) bool {
	var baseScalarTypes = newSet(
		reflect.TypeFor[int](),
		reflect.TypeFor[int8](),
		reflect.TypeFor[int16](),
		reflect.TypeFor[int32](),
		reflect.TypeFor[int64](),
		reflect.TypeFor[uint](),
		reflect.TypeFor[uint8](),
		reflect.TypeFor[uint16](),
		reflect.TypeFor[uint32](),
		reflect.TypeFor[uint64](),
		reflect.TypeFor[float32](),
		reflect.TypeFor[float64](),
		reflect.TypeFor[bool](),
		reflect.TypeFor[string](),
	)

	var baseScalarKinds = newSetWithCap[reflect.Kind](len(baseScalarTypes))
	for t := range baseScalarTypes {
		baseScalarKinds.add(t.Kind())
	}

	return func(t reflect.Type) bool {
		return baseScalarKinds.contain(t.Kind())
	}
}()

var isBaseStructType = func() func(t reflect.Type) bool {
	s := newSet(reflect.TypeFor[time.Time]())
	return func(t reflect.Type) bool {
		return s.contain(t)
	}
}()

func isBaseArrayType(t reflect.Type) bool {
	return t.Kind() == reflect.Array && t.Elem().Kind() == reflect.Uint8
}

func isImplementConverter(t reflect.Type) bool {
	if t.Kind() != reflect.Pointer {
		t = reflect.PointerTo(t)
	}
	var vt reflect.Type
	if method, ok := t.MethodByName(toArgMethodName); ok {
		mt := method.Type
		// 方法参数个数必须为1（接收者）
		if mt.NumIn() != 1 {
			return false
		}
		// 返回值个数必须为1
		if mt.NumOut() != 1 {
			return false
		}
		// 返回值基本标量类型
		vt = mt.Out(0)
		if !isBaseValueType(vt) && !isValidBuiltinType(vt) {
			return false
		}
	} else {
		return false
	}
	if method, ok := t.MethodByName(toFieldMethodName); ok {
		mt := method.Type
		// 方法参数个数必须为2（接收者，值）
		if mt.NumIn() != 2 {
			return false
		}
		// 方法第二个参数（值）类型必须与vt相同
		if mt.In(1) != vt {
			return false
		}
		// 必须无返回值
		if mt.NumOut() != 0 {
			return false
		}
		// 接受者必须为指针
		if _, ok = t.Elem().MethodByName(toFieldMethodName); ok {
			return false
		}
	} else {
		return false
	}
	return true
}

var isImplementScannerValuer = func() func(t reflect.Type) bool {
	valuerType := reflect.TypeFor[driver.Valuer]()
	scannerType := reflect.TypeFor[sql.Scanner]()
	return func(t reflect.Type) bool {
		if t.Kind() != reflect.Pointer {
			t = reflect.PointerTo(t)
		}
		return t.Implements(valuerType) && t.Implements(scannerType)
	}
}()

func isEntityType(t reflect.Type) bool {
	return t.Kind() == reflect.Struct && !isTupleType(t)
}

func isTupleType(t reflect.Type) bool {
	return t.PkgPath() == "github.com/jishaocong0910/cozy-orm" && strings.HasPrefix(t.String(), "orm.Tuple")
}
