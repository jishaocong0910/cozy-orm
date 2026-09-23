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
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSqlBuilder_Write(t *testing.T) {
	r := require.New(t)
	{
		b := newSqlBuilder("", QuotedIdentifier_.UNDEFINED)
		b.Write("a", 1, 2)
		b.WriteIf(true, " b", 3, 4)
		b.WriteIf(false, " c", 5)
		b.Write(" ").WritePh()
		b.Write(" ").Accept(WriterTest{})
		b.Write(" ").ForEach(b.Sep(","), []string{"s1", "s2", "s3"}, func(_ int, t string) {
			b.Write(t)
		})
		b.ForEach(b.SepFix(" (", ",", ")"), []string{"s1", "s2", "s3"}, func(_ int, t string) {
			b.Write(t)
		})
		b.ForEach(b.SepFix(" (", ",", ")"), nil, func(_ int, t string) {
			b.Write(t)
		})
		b.ForEach(b.SepFixOpt(" (", ",", ")"), []string{"s1", "s2", "s3"}, func(_ int, t string) {
			b.Write(t)
		})
		b.ForEach(b.SepFixOpt(" (", ",", ")"), nil, func(_ int, t string) {
			b.Write(t)
		})
		b.Args(5, 6)
		b.Write(" ").WriteColumn("col")
		b.Error(errors.New("test error"))

		r.Equal("a b ? test writer s1,s2,s3 (s1,s2,s3) () (s1,s2,s3) col", b.b.String())
		r.Equal([]any{1, 2, 3, 4, 5, 6}, b.args)
		r.EqualError(b.err, "test error")
	}
	{
		b := newSqlBuilder("", QuotedIdentifier_.Backtick)
		b.WriteColumn("col")
		r.Equal("`col`", b.b.String())
	}
	{
		b := newSqlBuilder("", QuotedIdentifier_.DoubleQuote)
		b.WriteColumn("col")
		r.Equal("\"col\"", b.b.String())
	}
	{
		b := newSqlBuilder("", QuotedIdentifier_.Bracket)
		b.WriteColumn("col")
		r.Equal("[col]", b.b.String())
	}
	{
		b := newSqlBuilder("$", QuotedIdentifier_.UNDEFINED)
		b.WritePh().Write(" ").WritePh().Write(" ").WritePh()
		r.Equal("$1 $2 $3", b.b.String())
	}
}

type WriterTest struct{}

func (w WriterTest) WriteSQL(b *SqlBuilder) {
	b.Write("test writer")
}
