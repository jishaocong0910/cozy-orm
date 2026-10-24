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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSubPrefix(t *testing.T) {
	r := require.New(t)
	m := NewNameMapper().SubPrefix("Tb")
	r.Equal("", m.Convert(""))
	r.Equal("Product", m.Convert("TbProduct"))
}

func TestSubSuffix(t *testing.T) {
	r := require.New(t)
	m := NewNameMapper().SubSuffix("Po")
	r.Equal("", m.Convert(""))
	r.Equal("Product", m.Convert("ProductPo"))
}

func TestAddPrefix(t *testing.T) {
	r := require.New(t)
	m := NewNameMapper().AddPrefix("Tb")
	r.Equal("", m.Convert(""))
	r.Equal("TbProduct", m.Convert("Product"))
}

func TestAddSuffix(t *testing.T) {
	r := require.New(t)
	m := NewNameMapper().AddSuffix("Po")
	r.Equal("", m.Convert(""))
	r.Equal("ProductPo", m.Convert("Product"))
}

func TestLowerCamelCase(t *testing.T) {
	r := require.New(t)
	m := NewNameMapper().LowerCamelCase()
	r.Equal("", m.Convert(""))
	r.Equal("tbProductPo", m.Convert("tb_product_Po"))
	r.Equal("tbProductPo", m.Convert("Tb Product Po"))
	r.Equal("tbProductPo", m.Convert("Tb-Product-Po"))
	r.Equal("tbProductPo", m.Convert("TB_PRODUCT_PO"))
	r.Equal("tbProductPo", m.Convert("TbProductPo"))
	r.Equal("tbProductPo", m.Convert("tbProductPo"))
}

func TestLowerSnakeCase(t *testing.T) {
	r := require.New(t)
	m := NewNameMapper().LowerSnakeCase()
	r.Equal("", m.Convert(""))
	r.Equal("tb_product_po", m.Convert("TbProductPo"))
	r.Equal("tb_product_po", m.Convert("Tb Product Po"))
	r.Equal("tb_product_po", m.Convert("Tb-Product-Po"))
	r.Equal("tb_product_po", m.Convert("TB_PRODUCT_PO"))
}

func TestLowerFirstLiteral(t *testing.T) {
	r := require.New(t)
	m := NewNameMapper().LowerFirstLiteral()
	r.Equal("", m.Convert(""))
	r.Equal("tbProductPo", m.Convert("TbProductPo"))
}

func TestUpperCamelCase(t *testing.T) {
	r := require.New(t)
	m := NewNameMapper().UpperCamelCase()
	r.Equal("", m.Convert(""))
	r.Equal("TbProductPo", m.Convert("tb_product_Po"))
	r.Equal("TbProductPo", m.Convert("Tb Product Po"))
	r.Equal("TbProductPo", m.Convert("Tb-Product-Po"))
	r.Equal("TbProductPo", m.Convert("TB_PRODUCT_PO"))
	r.Equal("TbProductPo", m.Convert("TbProductPo"))
	r.Equal("TbProductPo", m.Convert("tbProductPo"))
}

func TestUpperSnakeCase(t *testing.T) {
	r := require.New(t)
	m := NewNameMapper().UpperSnakeCase()
	r.Equal("", m.Convert(""))
	r.Equal("TB_PRODUCT_PO", m.Convert("TbProductPo"))
	r.Equal("TB_PRODUCT_PO", m.Convert("Tb Product Po"))
	r.Equal("TB_PRODUCT_PO", m.Convert("Tb-Product-Po"))
	r.Equal("TB_PRODUCT_PO", m.Convert("tb_product_po"))
}

func TestUpperFirstLiteral(t *testing.T) {
	r := require.New(t)
	m := NewNameMapper().UpperFirstLiteral()
	r.Equal("", m.Convert(""))
	r.Equal("TbProductPo", m.Convert("tbProductPo"))
}
