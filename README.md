# CozyORM

Lightweight Go ORM with method chaining, minimalist API, and multi-database support.

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/jishaocong0910/cozy-orm)
[![GoDoc](https://pkg.go.dev/badge/github.com/gookit/validate.svg)](https://pkg.go.dev/github.com/jishaocong0910/cozy-orm)
![coverage](https://raw.githubusercontent.com/jishaocong0910/cozy-orm/badges/.badges/main/coverage.svg)

👉 [🇨🇳 切换至中文说明 ↓](#chinese-version)

# Features

- **Method Chaining**: Adopting Go 1.27's revolutionary generic method feature.
- **Minimalist API**: Provides only two core execution methods, along with some methods for CRUD.
- **Dynamic SQL**: Supports dynamic SQL construction, compatible with all SQL dialects and driver placeholders.
- **Getting Generated Keys**: Provides various strategies for getting generated keys after inserting.
- **Custom Mapping**: Provides enhanced custom type mapping.
- **Transaction Propagation**: Propagates transactions via Context, making database operations reusable.
- **Predefined Entity**: For dynamic mapping when selecting only a few fields.
- **Code Generation**: Provides code generation for supported databases.

# Installation

```shell
go get github.com/jishaocong0910/cozy-orm
```

# Documentation

Please refer to the [Documentation]() for the full guide.

---

<div id="chinese-version"></div>

# CozyORM (中文说明)

轻量级Go语言ORM框架，链式调用模式，极少的API，灵活兼容各种数据库。

👉 [Back to English ↑](#cozyorm)

# 特色

- **链式调用**：依托Go 1.27泛型方法的革命性特性，实现流畅的链式调用。
- **轻量API**：核心仅2个基础执行方法，同时封装了CRUD等常用方法。
- **SQL构建**：支持动态SQL构建，兼容各种SQL方言与参数占位符。
- **获取生成Key**：兼容各主流数据库插入记录后获取自动生成Key的不同机制。
- **自定义映射**：提供更友好的自定义类型映射功能。
- **事务传播**：基于Context实现事务传播，使数据库操作代码的得到复用。
- **预定义实体**：用于查询少量字段场景下进行动态映射。
- **代码生成**：提供多种常用数据库的实体代码生成器。

# 安装

```shell
go get github.com/jishaocong0910/cozy-orm
```

# 文档

