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
	"strconv"
	"strings"
	_ "unsafe"
)

type SQLBuilder struct {
	b      strings.Builder
	args   []any
	cancel bool
	err    error
	ph     SQLWriter
	delim  IdentifierDelimiter
}

func (b *SQLBuilder) Write(str string, args ...any) *SQLBuilder {
	b.b.WriteString(str)
	b.AddArgs(args...)
	return b
}

func (b *SQLBuilder) WritePh() *SQLBuilder {
	b.Accept(b.ph)
	return b
}

func (b *SQLBuilder) WriteColumn(column string) *SQLBuilder {
	if b.delim.addQuotes != nil {
		column = b.delim.addQuotes(column)
	}
	b.Write(column)
	return b
}

func (b *SQLBuilder) AddArgs(args ...any) *SQLBuilder {
	b.args = append(b.args, args...)
	return b
}

func (b *SQLBuilder) ForEach[T any](sep separate, items []T, handler func(i int, item T)) *SQLBuilder {
	total := len(items)
	if sep.open != "" && (total > 0 || !sep.optional) {
		b.Write(sep.open)
	}
	if total > 0 {
		handler(0, items[0])
	}
	for i := 1; i < total; i++ {
		b.Write(sep.separator)
		handler(i, items[i])
	}
	if sep.close != "" && (total > 0 || !sep.optional) {
		b.Write(sep.close)
	}
	return b
}

func (b *SQLBuilder) Sep(separator string) separate {
	return separate{separator: separator}
}

func (b *SQLBuilder) SepWrap(open, separator, close string) separate {
	return separate{open: open, separator: separator, close: close, optional: false}
}

func (b *SQLBuilder) SepWrapOpt(open, separator, close string) separate {
	return separate{open: open, separator: separator, close: close, optional: true}
}

func (b *SQLBuilder) Accept(w SQLWriter) *SQLBuilder {
	w.WriteSQL(b)
	return b
}

func (b *SQLBuilder) Cancel() {
	b.cancel = true
}

func (b *SQLBuilder) Error(err error) {
	if b.err == nil {
		b.err = err
	}
}

func (b *SQLBuilder) sqlAndArgs() (sql string, args []any) {
	if b != nil {
		sql = b.b.String()
		args = b.args
	}
	return
}

type separate struct {
	open, separator, close string
	optional               bool
}

type SQLWriter interface {
	WriteSQL(b *SQLBuilder)
}

type defaultPh struct{}

func (d defaultPh) WriteSQL(b *SQLBuilder) {
	b.Write("?")
}

type prefixPh struct {
	prefix string
	argNum int
}

func (p *prefixPh) WriteSQL(b *SQLBuilder) {
	p.argNum++
	ph := p.prefix + strconv.Itoa(p.argNum)
	b.Write(ph)
}

func newSQLBuilder(paramPrefix string, qit IdentifierDelimiter) *SQLBuilder {
	var ph SQLWriter
	if paramPrefix != "" {
		ph = &prefixPh{prefix: paramPrefix}
	} else {
		ph = defaultPh{}
	}
	return &SQLBuilder{ph: ph, delim: qit}
}
