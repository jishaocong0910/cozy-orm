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

	e "github.com/jishaocong0910/enum"
)

type DbType struct {
	e.EnumElem
}

type _DbType struct {
	e.Enum[DbType]
	MySQL,
	Oracle,
	Postgres,
	SQLServer,
	SQLite DbType
}

var DbType_ = e.NewEnum(_DbType{})

type Level struct {
	e.EnumElem
}

type _Level struct {
	e.Enum[Level]
	Off,
	Debug,
	Info,
	Warn,
	Error Level
}

var Level_ = e.NewEnum(_Level{})

type QuotedIdentifier struct {
	e.EnumElem
	addQuotes func(string) string
}

type _QuotedIdentifier struct {
	e.Enum[QuotedIdentifier]
	Backtick,
	DoubleQuotes,
	Brackets QuotedIdentifier
}

var QuotedIdentifier_ = e.NewEnum(_QuotedIdentifier{
	Backtick: QuotedIdentifier{addQuotes: func(s string) string {
		return "`" + s + "`"
	}},
	DoubleQuotes: QuotedIdentifier{addQuotes: func(s string) string {
		return "\"" + s + "\""
	}},
	Brackets: QuotedIdentifier{addQuotes: func(s string) string {
		return "[" + s + "]"
	}},
})

type GetGeneratedKeyType struct {
	e.EnumElem
	writeSql func(b *SqlBuilder, autoColumn_ []string)
}

type _GetGeneratedKeyType struct {
	e.Enum[GetGeneratedKeyType]
	FirstInsertId,
	LastInsertId,
	Returning,
	Output GetGeneratedKeyType
}

var GetGeneratedKeyType_ = e.NewEnum(_GetGeneratedKeyType{
	Returning: GetGeneratedKeyType{
		writeSql: func(b *SqlBuilder, autoColumn_ []string) {
			b.ForEach(b.SepFixOpt(" RETURNING ", ", ", ""), autoColumn_, func(_ int, column string) {
				b.WriteColumn(column)
			})
		},
	},
	Output: GetGeneratedKeyType{
		writeSql: func(b *SqlBuilder, autoColumn_ []string) {
			b.ForEach(b.SepFixOpt(" OUTPUT ", ", ", ""), autoColumn_, func(_ int, column string) {
				b.Write("INSERTED.").WriteColumn(column)
			})
		},
	},
})

type PageType struct {
	e.EnumElem
	writeSql func(b *SqlBuilder, offset, count int)
}

type _PageType struct {
	e.Enum[PageType]
	LimitOffset,
	FetchNext PageType
}

var PageType_ = e.NewEnum(_PageType{
	LimitOffset: PageType{
		writeSql: func(b *SqlBuilder, offset, count int) {
			b.Write(" LIMIT ").Write(strconv.FormatInt(int64(count), 10))
			if offset > 0 {
				b.Write(" OFFSET ").Write(strconv.FormatInt(int64(offset), 10))
			}
		},
	},
	FetchNext: PageType{
		writeSql: func(b *SqlBuilder, offset, count int) {
			b.Write(" OFFSET ").Write(strconv.FormatInt(int64(offset), 10)).Write(" ROWS")
			b.Write(" FETCH NEXT ").Write(strconv.FormatInt(int64(count), 10)).Write(" ROWS ONLY")
		},
	},
})

type deleteSoftlyMode struct {
	e.EnumElem
}

type _deleteSoftlyMode struct {
	e.Enum[deleteSoftlyMode]
	pk, null deleteSoftlyMode
}

var deleteSoftlyMode_ = e.NewEnum(_deleteSoftlyMode{})
