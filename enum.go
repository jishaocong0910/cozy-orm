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

	"github.com/jishaocong0910/enum"
)

type DBType struct {
	enum.Val
}

type _DBType struct {
	enum.Typ[DBType]
	MySQL,
	Oracle,
	PostgreSQL,
	SQLServer,
	SQLite DBType
}

var DBType_ = enum.NewType(_DBType{})

type LogLevel struct {
	enum.Val
}

type _LogLevel struct {
	enum.Typ[LogLevel]
	Off,
	Debug,
	Info,
	Warn,
	Error LogLevel
}

var LogLevel_ = enum.NewType(_LogLevel{})

type IdentifierDelimiter struct {
	enum.Val
	quote func(string) string
}

type _IdentifierDelimiter struct {
	enum.Typ[IdentifierDelimiter]
	Backtick,
	DoubleQuote,
	Bracket IdentifierDelimiter
}

var IdentifierDelimiter_ = enum.NewType(_IdentifierDelimiter{
	Backtick: IdentifierDelimiter{quote: func(s string) string {
		return "`" + s + "`"
	}},
	DoubleQuote: IdentifierDelimiter{quote: func(s string) string {
		return "\"" + s + "\""
	}},
	Bracket: IdentifierDelimiter{quote: func(s string) string {
		return "[" + s + "]"
	}},
})

type GetGeneratedKeyMode struct {
	enum.Val
}

type _GetGeneratedKeyMode struct {
	enum.Typ[GetGeneratedKeyMode]
	FirstInsertId,
	LastInsertId,
	InsertReturning,
	SQLServer,
	Oracle GetGeneratedKeyMode
}

var GetGeneratedKeyMode_ = enum.NewType(_GetGeneratedKeyMode{})

type PageMode struct {
	enum.Val
	writeSQL func(b *SQLBuilder, offset, count int)
}

type _PageMode struct {
	enum.Typ[PageMode]
	LimitOffset,
	OffsetFetch PageMode
}

var PageMode_ = enum.NewType(_PageMode{
	LimitOffset: PageMode{
		writeSQL: func(b *SQLBuilder, offset, count int) {
			b.Text(" LIMIT ").Text(strconv.FormatInt(int64(count), 10))
			if offset > 0 {
				b.Text(" OFFSET ").Text(strconv.FormatInt(int64(offset), 10))
			}
		},
	},
	OffsetFetch: PageMode{
		writeSQL: func(b *SQLBuilder, offset, count int) {
			b.Text(" OFFSET ").Text(strconv.FormatInt(int64(offset), 10)).Text(" ROWS")
			b.Text(" FETCH NEXT ").Text(strconv.FormatInt(int64(count), 10)).Text(" ROWS ONLY")
		},
	},
})

type deleteSoftlyMode struct {
	enum.Val
}

type _deleteSoftlyMode struct {
	enum.Typ[deleteSoftlyMode]
	assignedPk, assignedNull deleteSoftlyMode
}

var deleteSoftlyMode_ = enum.NewType(_deleteSoftlyMode{})
