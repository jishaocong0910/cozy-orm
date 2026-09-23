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
		arg := argFetcher()
		c := Cond().If(func() bool {
			return false
		}, func(c *Condition) {
			c.Eq("c1", arg("c1"))
		}).If(func() bool {
			return true
		}, func(c *Condition) {
			c.Eq("c1", arg("c1"))
		}).Raw("c2 = 'c2'").
			Eq("c3", arg("c3")).Ne("c4", arg("c4")).
			Gt("c5", arg("c5")).Lt("c6", arg("c6")).
			Ge("c7", arg("c7")).Le("c8", arg("c8")).
			Like("c9", arg("c9")).LikeLeft("c10", arg("c10")).LikeRight("c11", arg("c11")).
			In("c12", []any{arg("c12"), arg("c12")}).Between("c13", arg("c13"), arg("c13")).
			IsNull("c14").IsNotNull("c15").
			Not().Eq("c16", arg("c16")).
			Or().Eq("c17", arg("c17")).Eq("c18", arg("c18"))

		b := newSqlBuilder("", QuotedIdentifier_.UNDEFINED)
		b.Accept(c)
		r.Equal("c1 = ? AND c2 = 'c2' AND c3 = ? AND c4 <> ? AND c5 > ? AND c6 < ? AND c7 >= ? AND c8 <= ? AND c9 LIKE ? "+
			"AND c10 LIKE ? AND c11 LIKE ? AND c12 IN(?, ?) AND c13 BETWEEN ? AND ? AND c14 IS NULL AND c15 IS NOT NULL "+
			"AND NOT c16 = ? OR c17 = ? AND c18 = ?", b.b.String())
		r.Equal([]any{"c1_1", "c3_2", "c4_3", "c5_4", "c6_5", "c7_6", "c8_7", "%c9_8%",
			"c10_9%", "%c11_10", "c12_11", "c12_12", "c13_13", "c13_14", "c16_15", "c17_16", "c18_17"}, b.args)
	}
}

func TestCondSub(t *testing.T) {
	r := require.New(t)
	{
		arg := argFetcher()
		c := Cond().Sub(Cond()).Sub(Cond().Eq("c1", arg("c1")).Or().Eq("c2", arg("c2")))
		b := newSqlBuilder("", QuotedIdentifier_.UNDEFINED)
		b.Accept(c)
		r.Equal("c1 = ? OR c2 = ?", b.b.String())
		r.Equal([]any{"c1_1", "c2_2"}, b.args)
	}
	{
		arg := argFetcher()
		c := Cond().Sub(Cond().Eq("c1", arg("c1")).Or().Eq("c2", arg("c2"))).Eq("c3", arg("c3"))
		b := newSqlBuilder("", QuotedIdentifier_.UNDEFINED)
		b.Accept(c)
		r.Equal("(c1 = ? OR c2 = ?) AND c3 = ?", b.b.String())
		r.Equal([]any{"c1_1", "c2_2", "c3_3"}, b.args)
	}
	{
		arg := argFetcher()
		c := Cond().Eq("c1", arg("c1")).Sub(Cond().Eq("c2", arg("c2")).Or().Eq("c3", arg("c3")))
		b := newSqlBuilder("", QuotedIdentifier_.UNDEFINED)
		b.Accept(c)
		r.Equal("c1 = ? AND (c2 = ? OR c3 = ?)", b.b.String())
		r.Equal([]any{"c1_1", "c2_2", "c3_3"}, b.args)
	}
	{
		arg := argFetcher()
		c := Cond().Not().Sub(Cond().Eq("c1", arg("c1")).Eq("c2", arg("c2"))).
			Not().Sub(Cond().Eq("c3", arg("c3")).Eq("c4", arg("c4")))
		b := newSqlBuilder("", QuotedIdentifier_.UNDEFINED)
		b.Accept(c)
		r.Equal("NOT (c1 = ? AND c2 = ?) AND NOT (c3 = ? AND c4 = ?)", b.b.String())
		r.Equal([]any{"c1_1", "c2_2", "c3_3", "c4_4"}, b.args)
	}
}

var argFetcher = func() func(string) string {
	var i int
	return func(str string) string {
		i++
		return str + "_" + strconv.Itoa(i)
	}
}
