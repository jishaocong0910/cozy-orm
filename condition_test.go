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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCond(t *testing.T) {
	r := require.New(t)
	{
		a := &argFetcher{}
		c := (&Cond{}).Eq("Eq", a.next("Eq")).Ne("Ne", a.next("Ne")).
			Gt("Gt", a.next("Gt")).Lt("Lt", a.next("Lt")).
			Ge("Ge", a.next("Ge")).Le("Le", a.next("Le")).
			Like("Like", a.next2("Like", func(next string) any { return "%" + next + "%" })).
			LikeLeft("LikeLeft", a.next2("LikeLeft", func(next string) any { return next + "%" })).
			LikeRight("LikeRight", a.next2("LikeRight", func(next string) any { return "%" + next })).
			LikePattern("LikePattern", a.next("LikePattern")).
			In("In", []any{a.next("In"), a.next("In")}).
			Between("Between", a.next("Between"), a.next("Between")).
			IsNull("IsNull").IsNotNull("IsNull").
			Not().Eq("Not", a.next("Not")).
			Or().Eq("OrEq1", a.next("OrEq1")).Eq("OrEq2", a.next("OrEq2"))

		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.Accept(c)
		r.Equal("Eq = ? AND Ne <> ? AND Gt > ? AND Lt < ? AND Ge >= ? AND Le <= ? AND Like LIKE ? AND LikeLeft LIKE ?"+
			" AND LikeRight LIKE ? AND LikePattern LIKE ? AND In IN(?, ?) AND Between BETWEEN ? AND ? AND IsNull IS NULL"+
			" AND IsNull IS NOT NULL AND NOT Not = ? OR OrEq1 = ? AND OrEq2 = ?", b.b.String())
		r.Equal(a.args, b.args)
	}
}

func TestCond_Expr(t *testing.T) {
	r := require.New(t)
	{
		a := &argFetcher{}
		c := (&Cond{}).Expr(func(c *CondExpr) {
			c.Str("c1 = 'c1'").Str(" AND c2 = ").Arg(a.next("c2"))
		})

		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.Accept(c)
		r.Equal("c1 = 'c1' AND c2 = ?", b.b.String())
		r.Equal(a.args, b.args)
	}
	{
		a := &argFetcher{}
		c := (&Cond{}).Expr(func(c *CondExpr) {
			c.Str("c1 = 'c1'").Str(" AND c2 = ").Arg(a.next("c2"))
		}).Eq("c3", a.next("c3"))

		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.Accept(c)
		r.Equal("(c1 = 'c1' AND c2 = ?) AND c3 = ?", b.b.String())
		r.Equal(a.args, b.args)
	}
	{
		a := &argFetcher{}
		c := (&Cond{}).Sub(func(c *Cond) {

		}).Expr(func(c *CondExpr) {
			c.Str("c1 = 'c1'").Str(" AND c2 = ").Arg(a.next("c2"))
		})

		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.Accept(c)
		r.Equal("c1 = 'c1' AND c2 = ?", b.b.String())
		r.Equal(a.args, b.args)
	}
	{
		a := &argFetcher{}
		c := (&Cond{}).Sub(func(c *Cond) {
			c.Eq("c1", a.next("c1"))
		}).Expr(func(c *CondExpr) {
			c.Str("c2 = 'c2'").Str(" AND c3 = ").Arg(a.next("c3"))
		})

		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.Accept(c)
		r.Equal("c1 = ? AND (c2 = 'c2' AND c3 = ?)", b.b.String())
		r.Equal(a.args, b.args)
	}
}

func TestCond_Sub(t *testing.T) {
	r := require.New(t)
	{
		a := &argFetcher{}
		c := (&Cond{}).Sub(nil).Sub(func(c *Cond) {
			c.Eq("c1", a.next("c1")).Or().Eq("c2", a.next("c2"))
		})
		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.Accept(c)
		r.Equal("c1 = ? OR c2 = ?", b.b.String())
		r.Equal([]any{"c1_1", "c2_2"}, b.args)
	}
	{
		a := &argFetcher{}
		c := (&Cond{}).Sub(func(c *Cond) {
			c.Eq("c1", a.next("c1")).Or().Eq("c2", a.next("c2"))
		}).Eq("c3", a.next("c3"))
		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.Accept(c)
		r.Equal("(c1 = ? OR c2 = ?) AND c3 = ?", b.b.String())
		r.Equal([]any{"c1_1", "c2_2", "c3_3"}, b.args)
	}
	{
		a := &argFetcher{}
		c := (&Cond{}).Eq("c1", a.next("c1")).Sub(func(c *Cond) {
			c.Eq("c2", a.next("c2")).Or().Eq("c3", a.next("c3"))
		})
		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.Accept(c)
		r.Equal("c1 = ? AND (c2 = ? OR c3 = ?)", b.b.String())
		r.Equal([]any{"c1_1", "c2_2", "c3_3"}, b.args)
	}
	{
		a := &argFetcher{}
		c := (&Cond{}).Not().Sub(func(c *Cond) {
			c.Eq("c1", a.next("c1")).Eq("c2", a.next("c2"))
		}).Not().Sub(func(c *Cond) {
			c.Eq("c3", a.next("c3")).Eq("c4", a.next("c4"))
		})
		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.Accept(c)
		r.Equal("NOT (c1 = ? AND c2 = ?) AND NOT (c3 = ? AND c4 = ?)", b.b.String())
		r.Equal([]any{"c1_1", "c2_2", "c3_3", "c4_4"}, b.args)
	}
}

type argFetcher struct {
	args []any
	i    int
}

func (a *argFetcher) next(prefix string) string {
	a.i++
	arg := prefix + "_" + strconv.Itoa(a.i)
	a.args = append(a.args, arg)
	return arg
}

func (a *argFetcher) next2(prefix string, actualHandler func(next string) any) string {
	a.i++
	arg := prefix + "_" + strconv.Itoa(a.i)
	if actualHandler != nil {
		a.args = append(a.args, actualHandler(arg))
	} else {
		a.args = append(a.args, arg)
	}
	return arg
}
