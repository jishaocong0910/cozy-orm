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
	"strings"
	"unicode"
)

type NameMapper struct {
	mapping_     []nameMapping
	hasCamelCase bool
}

func (n *NameMapper) Convert(str string) string {
	for _, m := range n.mapping_ {
		str = m(str)
	}
	return str
}
func NewNameMapper() *NameMapper {
	return &NameMapper{}
}

type nameMapping func(str string) string

func (n *NameMapper) LowerCamelCase() *NameMapper {
	n.hasCamelCase = true
	n.mapping_ = append(n.mapping_, lowerCamelCase)
	return n
}

func (n *NameMapper) LowerSnakeCase() *NameMapper {
	if !n.hasCamelCase {
		n.mapping_ = append(n.mapping_, lowerCamelCase)
	}
	n.mapping_ = append(n.mapping_, lowerSnakeCase)
	return n
}

func (n *NameMapper) LowerFirstLiteral() *NameMapper {
	n.mapping_ = append(n.mapping_, lowerFirstLiteral)
	return n
}

func (n *NameMapper) UpperCamelCase() *NameMapper {
	n.hasCamelCase = true
	n.mapping_ = append(n.mapping_, upperCamelCase)
	return n
}

func (n *NameMapper) UpperSnakeCase() *NameMapper {
	if !n.hasCamelCase {
		n.mapping_ = append(n.mapping_, upperCamelCase)
	}
	n.mapping_ = append(n.mapping_, upperSnakeCase)
	return n
}

func (n *NameMapper) UpperFirstLiteral() *NameMapper {
	n.mapping_ = append(n.mapping_, upperFirstLiteral)
	return n
}

func (n *NameMapper) AddPrefix(prefix string) *NameMapper {
	n.mapping_ = append(n.mapping_, addPrefix(prefix))
	return n
}

func (n *NameMapper) AddSuffix(suffix string) *NameMapper {
	n.mapping_ = append(n.mapping_, addSuffix(suffix))
	return n
}

func (n *NameMapper) SubPrefix(prefix string) *NameMapper {
	n.mapping_ = append(n.mapping_, subPrefix(prefix))
	return n
}

func (n *NameMapper) SubSuffix(suffix string) *NameMapper {
	n.mapping_ = append(n.mapping_, subSuffix(suffix))
	return n
}

var lowerCamelCase = func(str string) string {
	if str == "" {
		return str
	}
	var builder = strings.Builder{}
	char_ := []rune(str)
	builder.WriteRune(unicode.ToLower(char_[0]))
	up := false
	for i := 1; i < len(char_); i++ {
		c := char_[i]
		switch c {
		case '_', '-', ' ':
			up = true
			continue
		default:
			if up {
				up = false
				builder.WriteRune(unicode.ToUpper(c))
			} else if unicode.IsUpper(c) && unicode.IsLower(char_[i-1]) {
				builder.WriteRune(c)
			} else {
				builder.WriteRune(unicode.ToLower(c))
			}
		}
	}
	return builder.String()
}

var lowerSnakeCase = func(str string) string {
	if str == "" {
		return str
	}
	var builder = strings.Builder{}
	char_ := []rune(str)
	builder.WriteRune(unicode.ToLower(char_[0]))
	for i := 1; i < len(char_); i++ {
		c := char_[i]
		if unicode.IsUpper(c) && char_[i-1] != '-' && char_[i-1] != ' ' {
			builder.WriteRune('_')
			builder.WriteRune(unicode.ToLower(c))
		} else {
			builder.WriteRune(unicode.ToLower(c))
		}
	}
	return builder.String()
}

var lowerFirstLiteral = func(str string) string {
	if str == "" {
		return str
	}
	rune_ := []rune(str)
	return strings.ToLower(string(rune_[:1])) + string(rune_[1:])
}

var upperCamelCase = func(str string) string {
	if str == "" {
		return str
	}
	var builder = strings.Builder{}
	char_ := []rune(str)
	builder.WriteRune(unicode.ToUpper(char_[0]))
	up := false
	for i := 1; i < len(char_); i++ {
		c := char_[i]
		switch c {
		case '_', '-', ' ':
			up = true
			continue
		default:
			if up {
				up = false
				builder.WriteRune(unicode.ToUpper(c))
			} else if unicode.IsUpper(c) && unicode.IsLower(char_[i-1]) {
				builder.WriteRune(c)
			} else {
				builder.WriteRune(unicode.ToLower(c))
			}
		}
	}
	return builder.String()
}

var upperSnakeCase = func(str string) string {
	if str == "" {
		return str
	}
	var builder = strings.Builder{}
	char_ := []rune(str)
	builder.WriteRune(unicode.ToUpper(char_[0]))
	for i := 1; i < len(char_); i++ {
		c := char_[i]
		if unicode.IsUpper(c) && char_[i-1] != '-' && char_[i-1] != ' ' {
			builder.WriteRune('_')
			builder.WriteRune(unicode.ToUpper(c))
		} else {
			builder.WriteRune(unicode.ToUpper(c))
		}
	}
	return builder.String()
}

var upperFirstLiteral = func(str string) string {
	if str == "" {
		return str
	}
	rune_ := []rune(str)
	return strings.ToUpper(string(rune_[:1])) + string(rune_[1:])
}

var addPrefix = func(prefix string) nameMapping {
	return func(str string) string {
		if str == "" {
			return str
		}
		return prefix + str
	}
}

var addSuffix = func(suffix string) nameMapping {
	return func(str string) string {
		if str == "" {
			return str
		}
		return str + suffix
	}
}

var subPrefix = func(prefix string) nameMapping {
	return func(str string) string {
		if str == "" {
			return str
		}
		return strings.TrimPrefix(str, prefix)
	}
}

var subSuffix = func(suffix string) nameMapping {
	return func(str string) string {
		if str == "" {
			return str
		}
		return strings.TrimSuffix(str, suffix)
	}
}
