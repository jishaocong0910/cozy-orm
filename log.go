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
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Logger interface {
	Debug(ctx context.Context, msg string)
	Info(ctx context.Context, msg string)
	Warn(ctx context.Context, msg string)
	Error(ctx context.Context, msg string)
}

func printWarn(ctx context.Context, logger Logger, err error) {
	if err != nil {
		printLog(ctx, logger, Level_.Warn, fmt.Sprintf("%v", err))
	}
}

func printSql(ctx context.Context, logger Logger, level Level, tx bool, desc string, sql string, args []any,
	rowCount int64, affected int64, cost time.Duration, err error) {
	if err != nil {
		level = Level_.Error
	} else if level.IsUndefined() {
		return
	}

	var builder strings.Builder
	builder.WriteString("SQL: ")
	builder.WriteString(sql)
	builder.WriteString("; args:")
	builder.WriteString(formatArg(args))
	builder.WriteString(", tx: ")
	builder.WriteString(strconv.FormatBool(tx))
	if desc != "" {
		builder.WriteString(", desc: ")
		builder.WriteString(desc)
	}
	if rowCount >= 0 {
		builder.WriteString(", rows: ")
		builder.WriteString(strconv.FormatInt(rowCount, 10))
	}
	if affected >= 0 {
		builder.WriteString(", affected: ")
		builder.WriteString(strconv.FormatInt(affected, 10))
	}
	if cost >= 0 {
		builder.WriteString(", cost: ")
		builder.WriteString(strconv.FormatInt(int64(cost/time.Millisecond), 10))
		builder.WriteString("ms")
	}
	if err != nil {
		builder.WriteString(", error: ")
		builder.WriteString(err.Error())
	}
	printLog(ctx, logger, level, builder.String())
}

func printLog(ctx context.Context, logger Logger, level Level, msg string, args ...any) {
	if logger != nil {
		if len(args) > 0 {
			msg = fmt.Sprintf(msg, args...)
		}
		switch level.ID {
		case Level_.Debug.ID:
			logger.Debug(ctx, msg)
		case Level_.Info.ID:
			logger.Info(ctx, msg)
		case Level_.Warn.ID:
			logger.Warn(ctx, msg)
		case Level_.Error.ID:
			logger.Error(ctx, msg)
		}
	}
}

func formatArg(args []any) string {
	var builder strings.Builder
	sep := " "
	for _, a := range args {
		builder.WriteString(sep)

		if a == nil {
			builder.WriteString("<nil>")
			continue
		}

		v := reflect.ValueOf(a)
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				builder.WriteString("<nil>")
				continue
			}
			v = v.Elem()
		}

		builder.WriteString(fmt.Sprintf("%v", v.Interface()))
		builder.WriteString("(")
		builder.WriteString(v.Type().String())
		builder.WriteString(")")
	}
	str := builder.String()
	if str == "" {
		str = " "
	}
	return str
}
