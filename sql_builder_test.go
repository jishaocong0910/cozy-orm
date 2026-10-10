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

func TestSQLBuilder_Write(t *testing.T) {
	r := require.New(t)
	{
		b := newSQLBuilder("", IdentifierDelimiter_.UNDEFINED)
		b.Write("a", 1, 2)
		b.Write(" ").WritePh()
		b.Write(" ").Accept(WriterTest{})
		b.Write(" ").ForEach(b.Sep(","), []string{"s1", "s2", "s3"}, func(_ int, t string) {
			b.Write(t)
		})
		b.ForEach(b.SepWrap(" (", ",", ")"), []string{"s1", "s2", "s3"}, func(_ int, t string) {
			b.Write(t)
		})
		b.ForEach(b.SepWrap(" (", ",", ")"), nil, func(_ int, t string) {
			b.Write(t)
		})
		b.ForEach(b.SepWrapOpt(" (", ",", ")"), []string{"s1", "s2", "s3"}, func(_ int, t string) {
			b.Write(t)
		})
		b.ForEach(b.SepWrapOpt(" (", ",", ")"), nil, func(_ int, t string) {
			b.Write(t)
		})
		b.AddArgs(3, 4)
		b.Write(" ").WriteColumn("col")
		b.Error(errors.New("test error"))

		r.Equal("a ? test writer s1,s2,s3 (s1,s2,s3) () (s1,s2,s3) col", b.b.String())
		r.Equal([]any{1, 2, 3, 4}, b.args)
		r.EqualError(b.err, "test error")
	}
	{
		b := newSQLBuilder("", IdentifierDelimiter_.Backtick)
		b.WriteColumn("col")
		r.Equal("`col`", b.b.String())
	}
	{
		b := newSQLBuilder("", IdentifierDelimiter_.DoubleQuote)
		b.WriteColumn("col")
		r.Equal("\"col\"", b.b.String())
	}
	{
		b := newSQLBuilder("", IdentifierDelimiter_.Bracket)
		b.WriteColumn("col")
		r.Equal("[col]", b.b.String())
	}
	{
		b := newSQLBuilder("$", IdentifierDelimiter_.UNDEFINED)
		b.WritePh().Write(" ").WritePh().Write(" ").WritePh()
		r.Equal("$1 $2 $3", b.b.String())
	}
}

type WriterTest struct{}

func (w WriterTest) WriteSQL(b *SQLBuilder) {
	b.Write("test writer")
}
