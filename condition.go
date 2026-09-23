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

func Cond() *Condition {
	return &Condition{}
}

type Condition struct {
	condBase
	isNot   bool
	hasOr   bool
	nextNot bool
	nextOr  bool
	items   []cond
}

func (c *Condition) WriteSQL(b *SqlBuilder) {
	if c != nil && len(c.items) > 0 {
		c.doWrite(b, func() {
			for i, item := range c.items {
				if i != 0 {
					if item.isOr() {
						b.Write(" OR ")
					} else {
						b.Write(" AND ")
					}
				}
				b.Accept(item)
			}
		})
	}
}

func (c *Condition) notEmpty() bool {
	if c != nil {
		for _, item := range c.items {
			if item.notEmpty() {
				return true
			}
		}
	}
	return false
}

func (c *Condition) add(c2 cond) *Condition {
	if c.nextNot {
		c2.setNot()
		c.nextNot = false
	}
	if c.nextOr {
		c2.setOr()
		c.nextOr = false
		c.hasOr = true
	}
	if len(c.items) > 0 {
		if c2.canParen() {
			c2.setParen()
		}
		if len(c.items) == 1 {
			if c.items[0].canParen() {
				c.items[0].setParen()
			}
		}
	}
	c.items = append(c.items, c2)
	return c
}

func (c *Condition) setNot() {
	c.not = true
	if len(c.items) > 1 {
		c.condBase.setParen()
	}
}

func (c *Condition) canParen() bool {
	return len(c.items) > 1 && c.hasOr
}

func (c *Condition) setParen() {
	c.condBase.setParen()
}

func (c *Condition) Not() *Condition {
	c.nextNot = true
	return c
}

func (c *Condition) Or() *Condition {
	if len(c.items) > 0 {
		c.nextOr = true
	}
	return c
}

func (c *Condition) Sub(sub *Condition) *Condition {
	if sub.notEmpty() {
		c.add(sub)
	}
	return c
}

func (c *Condition) If(check func() bool, do func(c *Condition)) *Condition {
	if check() {
		do(c)
	}
	return c
}

func (c *Condition) Raw(sql string) *Condition {
	return c.add(&condRaw{sql: sql})
}

func (c *Condition) Eq(column string, arg any) *Condition {
	return c.add(&condBinOp{column: column, op: "=", arg: arg})
}

func (c *Condition) Ne(column string, arg any) *Condition {
	return c.add(&condBinOp{column: column, op: "<>", arg: arg})
}

func (c *Condition) Gt(column string, arg any) *Condition {
	return c.add(&condBinOp{column: column, op: ">", arg: arg})
}

func (c *Condition) Lt(column string, arg any) *Condition {
	return c.add(&condBinOp{column: column, op: "<", arg: arg})
}

func (c *Condition) Ge(column string, arg any) *Condition {
	return c.add(&condBinOp{column: column, op: ">=", arg: arg})
}

func (c *Condition) Le(column string, arg any) *Condition {
	return c.add(&condBinOp{column: column, op: "<=", arg: arg})
}

func (c *Condition) Like(column string, arg string) *Condition {
	return c.add(&condBinOp{column: column, op: "LIKE", arg: "%" + arg + "%"})
}

func (c *Condition) LikeLeft(column string, arg string) *Condition {
	return c.add(&condBinOp{column: column, op: "LIKE", arg: arg + "%"})
}

func (c *Condition) LikeRight(column string, arg string) *Condition {
	return c.add(&condBinOp{column: column, op: "LIKE", arg: "%" + arg})
}

func (c *Condition) In(column string, args []any) *Condition {
	return c.add(&condIn{column: column, args: args})
}

func (c *Condition) Between(column string, min, max any) *Condition {
	return c.add(&condBetween{column: column, min: min, max: max})
}

func (c *Condition) IsNull(column string) *Condition {
	return c.add(&condIsNull{column: column})
}

func (c *Condition) IsNotNull(column string) *Condition {
	return c.add(&condIsNotNull{column: column})
}

type cond interface {
	SqlWriter
	notEmpty() bool
	isOr() bool
	setNot()
	setOr()
	setParen()
	canParen() bool
}

type condBase struct {
	or    bool
	not   bool
	paren bool
}

func (c *condBase) notEmpty() bool {
	return true
}

func (c *condBase) isOr() bool {
	return c.or
}

func (c *condBase) setNot() {
	c.not = true
}

func (c *condBase) setOr() {
	c.or = true
}

func (c *condBase) canParen() bool { return false }

func (c *condBase) setParen() { c.paren = true }

func (c *condBase) doWrite(b *SqlBuilder, append func()) {
	if c.not {
		b.Write("NOT ")
	}
	if c.paren {
		b.Write("(")
	}
	append()
	if c.paren {
		b.Write(")")
	}
}

type condRaw struct {
	condBase
	sql string
}

func (c condRaw) WriteSQL(b *SqlBuilder) {
	c.doWrite(b, func() {
		b.Write(c.sql)
	})
}

type condBinOp struct {
	condBase
	column string
	op     string
	arg    any
}

func (c condBinOp) WriteSQL(b *SqlBuilder) {
	c.doWrite(b, func() {
		b.WriteColumn(c.column).Write(" ").Write(c.op).Write(" ").WritePh().Args(c.arg)
	})
}

type condIn struct {
	condBase
	column string
	args   []any
}

func (c condIn) WriteSQL(b *SqlBuilder) {
	c.doWrite(b, func() {
		b.WriteColumn(c.column).Write(" IN(")
		for i := 0; i < len(c.args); i++ {
			b.WriteIf(i != 0, ", ").WritePh()
		}
		b.Write(")", c.args...)
	})
}

type condBetween struct {
	condBase
	column   string
	min, max any
}

func (c condBetween) WriteSQL(b *SqlBuilder) {
	c.doWrite(b, func() {
		b.WriteColumn(c.column).Write(" BETWEEN ").WritePh().Write(" AND ").WritePh().Args(c.min, c.max)
	})
}

type condIsNull struct {
	condBase
	column string
}

func (c condIsNull) WriteSQL(b *SqlBuilder) {
	c.doWrite(b, func() {
		b.WriteColumn(c.column).Write(" IS NULL")
	})
}

type condIsNotNull struct {
	condBase
	column string
}

func (c condIsNotNull) WriteSQL(b *SqlBuilder) {
	c.doWrite(b, func() {
		b.WriteColumn(c.column).Write(" IS NOT NULL")
	})
}
