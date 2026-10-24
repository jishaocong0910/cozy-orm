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
	"strconv"
	"strings"
)

type entityTag struct {
	table string
}

type fieldTag struct {
	column string
	pk     bool
	auto   int64
	ignore bool
}

func parseEntityTag(tf reflect.StructField) entityTag {
	t := entityTag{}
	if ormTag, ok := tf.Tag.Lookup("orm"); ok {
		params := strings.SplitSeq(ormTag, ";")
		for p := range params {
			kv := strings.Split(p, "=")
			if len(kv) == 2 {
				k := strings.TrimSpace(kv[0])
				v := strings.TrimSpace(kv[1])
				switch k {
				case "table":
					t.table = v
				}
			}
		}
	}
	return t
}

func parseFieldTag(tf reflect.StructField) fieldTag {
	t := fieldTag{}
	if ormTag, ok := tf.Tag.Lookup("orm"); ok {
		params := strings.SplitSeq(ormTag, ";")
		for p := range params {
			kv := strings.Split(p, "=")
			if len(kv) == 1 {
				p = strings.TrimSpace(p)
				switch p {
				case "pk":
					t.pk = true
				case "auto":
					t.auto = 1
				case "ignore":
					t.ignore = true
				}
			} else if len(kv) == 2 {
				k := strings.TrimSpace(kv[0])
				v := strings.TrimSpace(kv[1])
				switch k {
				case "auto":
					i, err := strconv.ParseInt(v, 10, 64)
					if err == nil {
						t.auto = i
					}
				case "column":
					t.column = v
				}
			}
		}
	}
	return t
}
