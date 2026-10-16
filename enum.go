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

type DBType struct {
	e.EnumElem
}

type _DBType struct {
	e.Enum[DBType]
	MySQL,
	Oracle,
	PostgreSQL,
	SQLServer,
	SQLite DBType
}

var DBType_ = e.NewEnum(_DBType{})

type LogLevel struct {
	e.EnumElem
}

type _LogLevel struct {
	e.Enum[LogLevel]
	Off,
	Debug,
	Info,
	Warn,
	Error LogLevel
}

var LogLevel_ = e.NewEnum(_LogLevel{})

type IdentifierDelimiter struct {
	e.EnumElem
	quoted func(string) string
}

type _IdentifierDelimiter struct {
	e.Enum[IdentifierDelimiter]
	Backtick,
	DoubleQuote,
	Bracket IdentifierDelimiter
}

var IdentifierDelimiter_ = e.NewEnum(_IdentifierDelimiter{
	Backtick: IdentifierDelimiter{quoted: func(s string) string {
		return "`" + s + "`"
	}},
	DoubleQuote: IdentifierDelimiter{quoted: func(s string) string {
		return "\"" + s + "\""
	}},
	Bracket: IdentifierDelimiter{quoted: func(s string) string {
		return "[" + s + "]"
	}},
})

type GetGeneratedKeyMode struct {
	e.EnumElem
}

type _GetGeneratedKeyMode struct {
	e.Enum[GetGeneratedKeyMode]
	FirstInsertId,
	LastInsertId,
	InsertReturning,
	SQLServer,
	Oracle GetGeneratedKeyMode
}

var GetGeneratedKeyMode_ = e.NewEnum(_GetGeneratedKeyMode{})

type PageMode struct {
	e.EnumElem
	writeSQL func(b *SQLBuilder, offset, count int)
}

type _PageMode struct {
	e.Enum[PageMode]
	LimitOffset,
	OffsetFetch PageMode
}

var PageMode_ = e.NewEnum(_PageMode{
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
	e.EnumElem
}

type _deleteSoftlyMode struct {
	e.Enum[deleteSoftlyMode]
	assignedPk, assignedNull deleteSoftlyMode
}

var deleteSoftlyMode_ = e.NewEnum(_deleteSoftlyMode{})
