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
	"time"

	"github.com/stretchr/testify/require"
)

func TestPrintSql(t *testing.T) {
	r := require.New(t)
	{
		log := &MockLogger{}
		printSql(nil, log, Level_.UNDEFINED, true, "update a user", `  
SELECT *
  FROM
user`,
			[]any{2, nil, (*int)(nil), new("abc"), time.UnixMilli(1703659380000).In(time.UTC)},
			100, 20, 15*time.Millisecond, errors.New("error"))
		r.True(Level_.Error.Is(log.Messages[0].Level))
		r.Equal(`SQL:   
SELECT *
  FROM
user; args: 2(int) <nil> <nil> abc(string) 2023-12-27 06:43:00 +0000 UTC(time.Time), tx: true, desc: update a user, rows: 100, affected: 20, cost: 15ms, error: error`, log.Messages[0].Msg)
	}
	{
		log := &MockLogger{}
		printSql(nil, log, Level_.UNDEFINED, false, "test", "test", nil, -1, -1, -1, nil)
		r.Empty(log.Messages)
	}
	{
		log := &MockLogger{}
		printSql(nil, log, Level_.Debug, false, "test", "", nil, -1, -1, -1, nil)
		r.True(log.Messages[0].Level.Is(Level_.Debug))
		r.Equal("SQL: ; args: , tx: false, desc: test", log.Messages[0].Msg)
	}
	{
		log := &MockLogger{}
		printSql(nil, log, Level_.Info, false, "test", "", nil, -1, -1, -1, nil)
		r.True(log.Messages[0].Level.Is(Level_.Info))
		r.Equal("SQL: ; args: , tx: false, desc: test", log.Messages[0].Msg)
	}
}

func TestPrintWarn(t *testing.T) {
	r := require.New(t)
	log := &MockLogger{}
	printWarn(nil, log, errors.New("warn"))
	r.True(log.Messages[0].Level.Is(Level_.Warn))
	r.Equal("warn", log.Messages[0].Msg)
}
