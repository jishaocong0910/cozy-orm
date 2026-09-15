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
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
	"uuid"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
)

func MockDB(r *require.Assertions) (db *DB, mock sqlmock.Sqlmock) {
	d, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	r.NoError(err)
	return DbConfig{SqlDB: d}.Build(), mock
}

func MockSqlDB(r *require.Assertions) (d *sql.DB, mock sqlmock.Sqlmock) {
	d, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
	r.NoError(err)
	return
}

func MockLogger(db *DB) *mockLogger {
	log := &mockLogger{}
	if db != nil {
		db.logger = log
	}
	return log
}

func convertId[T any](id int64) any {
	return lastInsertIdConversionMap[reflect.TypeFor[T]()](id).Elem().Interface()
}

func checkMapKeys[K comparable, V any](r *require.Assertions, keys []K, actual map[K]V) {
	r.Len(keys, len(actual))
	for _, key := range keys {
		r.Contains(actual, key)
	}
}

func equalFunc(r *require.Assertions, f1, f2 any) {
	r.Equal(reflect.ValueOf(f1).Pointer(), reflect.ValueOf(f2).Pointer())
}

type logMsg struct {
	Ctx   context.Context
	Level Level
	Msg   string
}

type mockLogger struct {
	Msgs []logMsg
}

func (d *mockLogger) Debug(ctx context.Context, msg string) {
	d.Msgs = append(d.Msgs, logMsg{Ctx: ctx, Level: Level_.Debug, Msg: msg})
}

func (d *mockLogger) Info(ctx context.Context, msg string) {
	d.Msgs = append(d.Msgs, logMsg{Ctx: ctx, Level: Level_.Info, Msg: msg})
}

func (d *mockLogger) Warn(ctx context.Context, msg string) {
	d.Msgs = append(d.Msgs, logMsg{Ctx: ctx, Level: Level_.Warn, Msg: msg})
}

func (d *mockLogger) Error(ctx context.Context, msg string) {
	d.Msgs = append(d.Msgs, logMsg{Ctx: ctx, Level: Level_.Error, Msg: msg})
}

type UnsupportedLastInsertIdResult struct {
}

func (u UnsupportedLastInsertIdResult) LastInsertId() (int64, error) {
	return 0, errors.New("lastInsertId is not supported")
}

func (u UnsupportedLastInsertIdResult) RowsAffected() (int64, error) {
	return 0, nil
}

type UserSimple struct {
	Name  string
	Phone string
	Email string
	Level string
}

type User struct {
	Id         *int64  `orm:"column=id,pk,auto"`
	Name       *string `orm:"column=name"`
	Password   *string `orm:"column=password,ignore"`
	Address    *string `orm:"ignore"`
	Phone      *string `orm:"column=phone"`
	Email      *string
	AvatarUrl  []byte `orm:"column=avatar_url"`
	Status     *UserStatus
	Level      *UserLevel
	Properties *UserProperties
	Category   *UserCategory
	Tags       UserTags
	Attributes UserAttributes
	Uuid       *uuid.UUID
	CreateAt   *time.Time
	UpdateAt   *time.Time
	Version    *int64
	Deleted    *int64
}

type UserStatus int

type UserLevel string

func (u UserLevel) ToValue() int8 {
	i, _ := strconv.ParseInt(string(u), 10, 8)
	return int8(i)
}

func (u *UserLevel) ToField(value int8) {
	str := strconv.FormatInt(int64(value), 10)
	*u = UserLevel(str)
}

type UserProperties struct {
	Source  string `json:"source"`
	Country string `json:"country"`
}

func (p *UserProperties) ToValue() string {
	bs, _ := json.Marshal(p)
	return string(bs)
}

func (p *UserProperties) ToField(value string) {
	json.Unmarshal([]byte(value), &p)
}

type UserTags []string

func (s UserTags) ToValue() string {
	return strings.Join(s, ",")
}

func (s *UserTags) ToField(value string) {
	arr := strings.Split(value, ",")
	for _, a := range arr {
		*s = append(*s, a)
	}
}

type UserAttributes map[string]string

func (a UserAttributes) ToValue() string {
	bs, _ := json.Marshal(a)
	return string(bs)
}

func (a *UserAttributes) ToField(value string) {
	json.Unmarshal([]byte(value), &a)
}

type UserCategory struct {
	Organization string `json:"organization"`
	Class        int    `json:"class"`
}

func (c UserCategory) Value() (driver.Value, error) {
	bs, _ := json.Marshal(c)
	return string(bs), nil
}

func (c *UserCategory) Scan(src any) error {
	json.Unmarshal([]byte(src.(string)), &c)
	return nil
}

type DemoEntityTag struct {
	_ struct{} `orm:"table=demo"`
}

type DemoIgnoreField struct {
	Id *string `orm:"column=id,auto=2"`

	DemoIgnoreField2                                // ignore
	ignoreUnexported              *string           `orm:"column=ignore1"`
	IgnoreNotPointer              string            `orm:"column=ignore2"`
	IgnoreNotBaseType             *any              `orm:"column=ignore3"`
	IgnoreNotBaseTypeElem         []any             `orm:"column=ignore4"`
	IgnoreInvalidImplementConvert *DemoIgnoreField2 `orm:"column=ignore5"`
}

type DemoIgnoreField2 struct {
	Field *string
}

type DemoMulPk struct {
	Id1 *string `orm:"pk,auto"`
	Id2 *string `orm:"pk,auto"`
}

type DemoDemand struct {
	DemoDemandAnon
	Name  string
	Phone string
	Email string
	level string
	Field string
}

type DemoDemandAnon struct {
	CreateAt *time.Time
	UpdateAt *time.Time
	Version  *int64
	Deleted  *int64
}

type TxInfoInner struct {
	Creator *tx
	SqlTx   *sql.Tx
	TxHook  []*txHook
}

func GetTxInfoInner(ctx context.Context) *TxInfoInner {
	ti := cvTx.get(ctx)
	if ti == nil || ti.txInfoInner == nil {
		return nil
	}
	return &TxInfoInner{
		Creator: ti.creator,
		SqlTx:   ti.sqlTx,
		TxHook:  ti.txHook_,
	}
}

type AnyTime struct{}

func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}
