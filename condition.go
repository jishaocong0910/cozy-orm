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

type Condition struct {
	condBase
	isNot   bool
	hasOr   bool
	nextNot bool
	nextOr  bool
	items   []condItem
}

func (c *Condition) WriteSQL(b *SQLBuilder) {
	if c != nil && len(c.items) > 0 {
		c.writeWrap(b, func() {
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

func (c *Condition) isEmpty() bool {
	if c != nil {
		for _, item := range c.items {
			if !item.isEmpty() {
				return false
			}
		}
	}
	return true
}

func (c *Condition) setNot() {
	c.not = true
	if len(c.items) > 1 {
		c.paren = true
	}
}

func (c *Condition) setParen() {
	if len(c.items) > 1 && c.hasOr {
		c.paren = true
	}
}

func (c *Condition) add(item condItem) *Condition {
	if item.isEmpty() {
		return c
	}
	if c.nextNot {
		item.setNot()
		c.nextNot = false
	}
	if c.nextOr {
		item.setOr()
		c.nextOr = false
		c.hasOr = true
	}
	if len(c.items) > 0 {
		item.setParen()
		if len(c.items) == 1 {
			c.items[0].setParen()
		}
	}
	c.items = append(c.items, item)
	return c
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

func (c *Condition) Sub(handler func(c *Condition)) *Condition {
	if handler != nil {
		c2 := &Condition{}
		handler(c2)
		c.add(c2)
	}
	return c
}

func (c *Condition) Expr(handler func(c *CondExpr)) *Condition {
	if handler != nil {
		e := &CondExpr{}
		handler(e)
		c.add(e)
	}
	return c
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

func (c *Condition) Like(column string, str string) *Condition {
	return c.add(&condBinOp{column: column, op: "LIKE", arg: "%" + str + "%"})
}

func (c *Condition) LikeLeft(column string, str string) *Condition {
	return c.add(&condBinOp{column: column, op: "LIKE", arg: str + "%"})
}

func (c *Condition) LikeRight(column string, str string) *Condition {
	return c.add(&condBinOp{column: column, op: "LIKE", arg: "%" + str})
}

func (c *Condition) LikePattern(column string, pattern string) *Condition {
	return c.add(&condBinOp{column: column, op: "LIKE", arg: pattern})
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

type condItem interface {
	SQLWriter
	isEmpty() bool
	isOr() bool
	setNot()
	setOr()
	setParen()
}

type condBase struct {
	or    bool
	not   bool
	paren bool
}

func (c *condBase) isEmpty() bool {
	return false
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

func (c *condBase) setParen() {}

func (c *condBase) writeWrap(b *SQLBuilder, write func()) {
	if c.not {
		b.Write("NOT ")
	}
	if c.paren {
		b.Write("(")
	}
	write()
	if c.paren {
		b.Write(")")
	}
}

type condBinOp struct {
	condBase
	column string
	op     string
	arg    any
}

func (c condBinOp) WriteSQL(b *SQLBuilder) {
	c.writeWrap(b, func() {
		b.WriteColumn(c.column).Write(" ").Write(c.op).Write(" ").WritePh().AddArgs(c.arg)
	})
}

type condIn struct {
	condBase
	column string
	args   []any
}

func (c condIn) WriteSQL(b *SQLBuilder) {
	c.writeWrap(b, func() {
		b.WriteColumn(c.column).Write(" IN(")
		for i := 0; i < len(c.args); i++ {
			if i != 0 {
				b.Write(", ")
			}
			b.WritePh()
		}
		b.Write(")", c.args...)
	})
}

type condBetween struct {
	condBase
	column   string
	min, max any
}

func (c condBetween) WriteSQL(b *SQLBuilder) {
	c.writeWrap(b, func() {
		b.WriteColumn(c.column).Write(" BETWEEN ").WritePh().Write(" AND ").WritePh().AddArgs(c.min, c.max)
	})
}

type condIsNull struct {
	condBase
	column string
}

func (c condIsNull) WriteSQL(b *SQLBuilder) {
	c.writeWrap(b, func() {
		b.WriteColumn(c.column).Write(" IS NULL")
	})
}

type condIsNotNull struct {
	condBase
	column string
}

func (c condIsNotNull) WriteSQL(b *SQLBuilder) {
	c.writeWrap(b, func() {
		b.WriteColumn(c.column).Write(" IS NOT NULL")
	})
}

type CondExpr struct {
	condBase
	items []any
	args  []any
}

func (c *CondExpr) isEmpty() bool {
	return len(c.items) == 0
}

func (c *CondExpr) setParen() {
	c.paren = true
}

func (c *CondExpr) WriteSQL(b *SQLBuilder) {
	c.writeWrap(b, func() {
		for _, item := range c.items {
			if i, ok := item.(int); ok {
				b.WritePh().AddArgs(c.args[i])
			} else {
				b.Write(item.(string))
			}
		}
	})
}

func (c *CondExpr) Str(str string) *CondExpr {
	c.items = append(c.items, str)
	return c
}

func (c *CondExpr) Arg(a any) *CondExpr {
	c.args = append(c.args, a)
	c.items = append(c.items, len(c.args)-1)
	return c
}
