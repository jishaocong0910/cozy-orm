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

type Cond struct {
	condBase
	isNot   bool
	hasOr   bool
	nextNot bool
	nextOr  bool
	items   []condItem
	err     error
}

func (c *Cond) writeSQL(b *SQLBuilder) {
	if c != nil && len(c.items) > 0 {
		c.writeWrap(b, func() {
			for i, item := range c.items {
				if i != 0 {
					if item.isOr() {
						b.Text(" OR ")
					} else {
						b.Text(" AND ")
					}
				}
				b.accept(item)
			}
		})
	}
}

func (c *Cond) isEmpty() bool {
	if c != nil {
		for _, item := range c.items {
			if !item.isEmpty() {
				return false
			}
		}
	}
	return true
}

func (c *Cond) setNot() {
	c.not = true
	if len(c.items) > 1 {
		c.paren = true
	}
}

func (c *Cond) setParen() {
	if len(c.items) > 1 && c.hasOr {
		c.paren = true
	}
}

func (c *Cond) add(item condItem) *Cond {
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

func (c *Cond) Not() *Cond {
	c.nextNot = true
	return c
}

func (c *Cond) Or() *Cond {
	if len(c.items) > 0 {
		c.nextOr = true
	}
	return c
}

func (c *Cond) Sub(handler func(c *Cond)) *Cond {
	if handler != nil {
		c2 := &Cond{}
		handler(c2)
		if c2.err != nil {
			c.err = c2.err
		} else {
			c.add(c2)
		}
	}
	return c
}

func (c *Cond) Expr(handler func(c *CondExpr)) *Cond {
	if handler != nil {
		c2 := &CondExpr{}
		handler(c2)
		if c2.err != nil {
			c.err = c2.err
		} else {
			c.add(c2)
		}
	}
	return c
}

func (c *Cond) Eq(column string, arg any) *Cond {
	return c.add(&condBinOp{column: column, op: "=", arg: arg})
}

func (c *Cond) Ne(column string, arg any) *Cond {
	return c.add(&condBinOp{column: column, op: "<>", arg: arg})
}

func (c *Cond) Gt(column string, arg any) *Cond {
	return c.add(&condBinOp{column: column, op: ">", arg: arg})
}

func (c *Cond) Lt(column string, arg any) *Cond {
	return c.add(&condBinOp{column: column, op: "<", arg: arg})
}

func (c *Cond) Ge(column string, arg any) *Cond {
	return c.add(&condBinOp{column: column, op: ">=", arg: arg})
}

func (c *Cond) Le(column string, arg any) *Cond {
	return c.add(&condBinOp{column: column, op: "<=", arg: arg})
}

func (c *Cond) Like(column string, str string) *Cond {
	return c.add(&condBinOp{column: column, op: "LIKE", arg: "%" + str + "%"})
}

func (c *Cond) LikeLeft(column string, str string) *Cond {
	return c.add(&condBinOp{column: column, op: "LIKE", arg: str + "%"})
}

func (c *Cond) LikeRight(column string, str string) *Cond {
	return c.add(&condBinOp{column: column, op: "LIKE", arg: "%" + str})
}

func (c *Cond) LikePattern(column string, pattern string) *Cond {
	return c.add(&condBinOp{column: column, op: "LIKE", arg: pattern})
}

func (c *Cond) In[T any](column string, args []T) *Cond {
	return c.add(&condIn[T]{column: column, args: args})
}

func (c *Cond) Between(column string, min, max any) *Cond {
	return c.add(&condBetween{column: column, min: min, max: max})
}

func (c *Cond) IsNull(column string) *Cond {
	return c.add(&condIsNull{column: column})
}

func (c *Cond) IsNotNull(column string) *Cond {
	return c.add(&condIsNotNull{column: column})
}

func (c *Cond) Error(err error) {
	c.err = err
}

type condItem interface {
	sqlWriter
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
		b.Text("NOT ")
	}
	if c.paren {
		b.Text("(")
	}
	write()
	if c.paren {
		b.Text(")")
	}
}

type condBinOp struct {
	condBase
	column string
	op     string
	arg    any
}

func (c condBinOp) writeSQL(b *SQLBuilder) {
	c.writeWrap(b, func() {
		b.Quoted(c.column).Text(" ").Text(c.op).Text(" ").Arg(c.arg)
	})
}

type condIn[T any] struct {
	condBase
	column string
	args   []T
}

func (c condIn[T]) writeSQL(b *SQLBuilder) {
	c.writeWrap(b, func() {
		b.Quoted(c.column).Text(" IN(")
		for i := 0; i < len(c.args); i++ {
			if i != 0 {
				b.Text(", ")
			}
			b.Arg(c.args[i])
		}
		b.Text(")")
	})
}

type condBetween struct {
	condBase
	column   string
	min, max any
}

func (c condBetween) writeSQL(b *SQLBuilder) {
	c.writeWrap(b, func() {
		b.Quoted(c.column).Text(" BETWEEN ").Arg(c.min).Text(" AND ").Arg(c.max)
	})
}

type condIsNull struct {
	condBase
	column string
}

func (c condIsNull) writeSQL(b *SQLBuilder) {
	c.writeWrap(b, func() {
		b.Quoted(c.column).Text(" IS NULL")
	})
}

type condIsNotNull struct {
	condBase
	column string
}

func (c condIsNotNull) writeSQL(b *SQLBuilder) {
	c.writeWrap(b, func() {
		b.Quoted(c.column).Text(" IS NOT NULL")
	})
}

type CondExpr struct {
	condBase
	items []any
	args  []any
	err   error
}

func (c *CondExpr) isEmpty() bool {
	return len(c.items) == 0
}

func (c *CondExpr) setParen() {
	c.paren = true
}

func (c *CondExpr) writeSQL(b *SQLBuilder) {
	c.writeWrap(b, func() {
		for _, item := range c.items {
			if i, ok := item.(int); ok {
				b.Arg(c.args[i])
			} else {
				b.Text(item.(string))
			}
		}
	})
}

func (c *CondExpr) Text(str string) *CondExpr {
	c.items = append(c.items, str)
	return c
}

func (c *CondExpr) Arg(a any) *CondExpr {
	c.args = append(c.args, a)
	c.items = append(c.items, len(c.args)-1)
	return c
}

func (c *CondExpr) Error(err error) {
	c.err = err
}
