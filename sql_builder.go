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

type SqlBuilder struct {
	b      strings.Builder
	args   []any
	cancel bool
	err    error
	ph     SqlWriter
	qit    QuotedIdentifier
}

func (b *SqlBuilder) Write(str string, args ...any) *SqlBuilder {
	b.b.WriteString(str)
	b.Args(args...)
	return b
}

func (b *SqlBuilder) WriteIf(check bool, str string, args ...any) *SqlBuilder {
	if check {
		b.b.WriteString(str)
		b.Args(args...)
	}
	return b
}

func (b *SqlBuilder) WritePh() *SqlBuilder {
	b.Accept(b.ph)
	return b
}

func (b *SqlBuilder) WriteColumn(column string) *SqlBuilder {
	if b.qit.addQuotes != nil {
		column = b.qit.addQuotes(column)
	}
	b.Write(column)
	return b
}

func (b *SqlBuilder) Accept(w SqlWriter) *SqlBuilder {
	w.WriteSQL(b)
	return b
}

func (b *SqlBuilder) ForEach[T any](sep separate, elems []T, handle func(i int, t T)) *SqlBuilder {
	total := len(elems)
	if sep.prefix != "" && (total > 0 || !sep.omitempty) {
		b.Write(sep.prefix)
	}
	if total > 0 {
		handle(0, elems[0])
	}
	for i := 1; i < total; i++ {
		b.Write(sep.separator)
		handle(i, elems[i])
	}
	if sep.suffix != "" && (total > 0 || !sep.omitempty) {
		b.Write(sep.suffix)
	}
	return b
}

func (b *SqlBuilder) Args(args ...any) *SqlBuilder {
	b.args = append(b.args, args...)
	return b
}

func (b *SqlBuilder) Cancel() {
	b.cancel = true
}

func (b *SqlBuilder) Error(err error) {
	if b.err == nil {
		b.err = err
	}
}

func (b *SqlBuilder) Sep(separator string) separate {
	return separate{separator: separator}
}

func (b *SqlBuilder) SepFix(prefix, separator, suffix string) separate {
	return separate{prefix: prefix, separator: separator, suffix: suffix, omitempty: false}
}

func (b *SqlBuilder) SepFixOpt(prefix, separator, suffix string) separate {
	return separate{prefix: prefix, separator: separator, suffix: suffix, omitempty: true}
}

func (b *SqlBuilder) sqlAndArgs() (sql string, args []any) {
	if b != nil {
		sql = b.b.String()
		args = b.args
	}
	return
}

type separate struct {
	prefix, separator, suffix string
	omitempty                 bool
}

type SqlWriter interface {
	WriteSQL(b *SqlBuilder)
}

type defaultPh struct{}

func (d defaultPh) WriteSQL(b *SqlBuilder) {
	b.Write("?")
}

type prefixPh struct {
	prefix string
	argNum int
}

func (p *prefixPh) WriteSQL(b *SqlBuilder) {
	p.argNum++
	ph := p.prefix + strconv.Itoa(p.argNum)
	b.Write(ph)
}

func newSqlBuilder(paramPrefix string, qit QuotedIdentifier) *SqlBuilder {
	var ph SqlWriter
	if paramPrefix != "" {
		ph = &prefixPh{prefix: paramPrefix}
	} else {
		ph = defaultPh{}
	}
	return &SqlBuilder{ph: ph, qit: qit}
}
