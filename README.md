<a id="english"></a>

# CozyORM

[English](#english) | [中文](#chinese)

Lightweight Go ORM with method chaining, minimalist API, and multi-database support.

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/jishaocong0910/cozy-orm)
[![GoDoc](https://pkg.go.dev/badge/github.com/gookit/validate.svg)](https://pkg.go.dev/github.com/jishaocong0910/cozy-orm)
![coverage](https://raw.githubusercontent.com/jishaocong0910/cozy-orm/badges/.badges/main/coverage.svg)

## Features

- **Method Chaining**: Adopting Go 1.27's revolutionary generic method feature.
- **Minimalist API**: Only two base execution APIs (Query & Mutation), along with common CRUD APIs.
- **Dynamic SQL**: Supports dynamic SQL construction, compatible with all SQL dialects and driver placeholders.
- **Getting Generated Keys**: Provides various strategies for getting generated keys after inserting.
- **Custom Mapping**: Provides enhanced custom type mapping.
- **Transaction Propagation**: Propagates transactions via Context, making database operations reusable.
- **Predefined Entity**: For dynamic mapping when selecting only a few fields.
- **Code Generation**: Provides code generation for supported databases.

## Installation

```shell
go get github.com/jishaocong0910/cozy-orm
```

## Documentation

https://jishaocong0910.github.io/cozy-orm-doc/

---

<a id="chinese"></a>

# CozyORM (中文)

[English](#english) | [中文](#chinese)

轻量级Go语言ORM框架，链式调用模式，极少的API，灵活兼容各种数据库。

## 特色

- **链式调用**：依托 Go 1.27 泛型方法的革命性特性，实现流畅的链式调用。
- **轻量 API**：仅查询、变更两个基础执行 API，同时封装了 CRUD 等通用 API。
- **SQL 构建**：支持动态 SQL 构建，兼容各种 SQL 方言与参数占位符。
- **获取生成 Key**：兼容各主流数据库插入记录后获取自动生成 Key 的不同机制。
- **自定义映射**：提供更友好的自定义类型映射功能。
- **事务传播**：基于 Context 实现事务传播，使数据库操作代码得到复用。
- **预定义实体**：用于查询少量字段场景下进行动态映射。
- **代码生成**：提供多种常用数据库的实体代码生成器。

## 安装

```shell
go get github.com/jishaocong0910/cozy-orm
```

## 文档

https://jishaocong0910.github.io/cozy-orm-doc/zh/

