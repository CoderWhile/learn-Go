# GORM 创建（Create）详解

> 官方原文：<https://gorm.io/zh_CN/docs/create.html>
> 本地环境对应版本：`gorm.io/gorm v1.31.2` + `gorm.io/driver/mysql v1.6.0`（见 `GORM/CRUD/go.mod`）
> 本文档在官方页面基础上做了三件事：**逐句拆解原理**、**标注坑点**、**给出可直接运行的示例**。

---

## 目录

- [0. 全局视角：Create 到底做了什么](#0-全局视角create-到底做了什么)
- [1. 准备工作：模型与连接](#1-准备工作模型与连接)
- [2. 创建记录（Create Record）](#2-创建记录create-record)
  - [2.1 Traditional API](#21-traditional-api)
  - [2.2 Generics API（1.30+ 新增）](#22-generics-api130-新增)
  - [2.3 `ctx context.Context` 到底是什么](#23-ctx-contextcontext-到底是什么)
  - [2.4 `gorm.WithResult()` 到底是什么](#24-gormwithresult-到底是什么)
  - [2.5 为什么"必须传指针"](#25-为什么必须传指针)
- [3. 用指定的字段创建记录（Select / Omit）](#3-用指定的字段创建记录select--omit)
- [4. 批量插入](#4-批量插入)
- [5. 创建钩子（Hooks）](#5-创建钩子hooks)
  - [5.1 钩子是什么](#51-钩子是什么)
  - [5.2 钩子有什么用（为什么要用它）](#52-钩子有什么用为什么要用它)
  - [5.3 全部钩子接口一览](#53-全部钩子接口一览)
  - [5.4 执行顺序（重点，极易记错）](#54-执行顺序重点极易记错)
  - [5.5 应用实例](#55-应用实例)
  - [5.6 使用钩子的注意事项](#56-使用钩子的注意事项)
  - [5.7 SkipHooks：跳过钩子](#57-skiphooks跳过钩子)
- [6. 根据 Map 创建](#6-根据-map-创建)
- [7. 使用 SQL 表达式 / Context Valuer 创建](#7-使用-sql-表达式--context-valuer-创建)
- [8. 高级选项](#8-高级选项)
  - [8.1 关联创建（Associations）](#81-关联创建associations)
  - [8.2 默认值（default tag）](#82-默认值default-tag)
  - [8.3 Upsert 及冲突（OnConflict）](#83-upsert-及冲突onconflict)
- [9. 坑点速查表](#9-坑点速查表)
- [10. 动手练习](#10-动手练习)

> 📌 **配套阅读**
> - 学习路线与整体概览 → [`GORM学习路线与整体概览.md`](./GORM学习路线与整体概览.md)
> - 疑问与知识点台账（事务 / 会话 / 连接的深入解释）→ [`GORM答疑记录.md`](./GORM答疑记录.md)
> - 可运行验证程序 → [`../事务与会话/main.go`](../事务与会话/main.go)

---

## 0. 全局视角：Create 到底做了什么

`Create` 是 CRUD 里的 "C"，最终落到一条 `INSERT` 语句上。但 GORM 在发 SQL 之前/之后还包了一整套流程：

```
db.Create(&user)
   │
   ├─ ① 解析模型 → 表名、字段、主键、默认值标签
   ├─ ② BeforeSave  → BeforeCreate   （钩子）
   ├─ ③ 处理关联（保存 belongs to / has one 等"前置关联"）
   ├─ ④ 生成并执行 INSERT SQL
   ├─ ⑤ 回写主键 / 回写默认值到结构体
   ├─ ⑥ 处理关联（保存 has many / many2many 等"后置关联"）
   ├─ ⑦ AfterCreate → AfterSave      （钩子，注意这个顺序）
   └─ ⑧ 提交事务
```

GORM 内部把这些步骤注册成一条**回调链（callbacks）**，顺序固定：

| 顺序 | 回调名 | 作用 |
|---|---|---|
| 1 | `gorm:begin_transaction` | 开启事务（可通过配置关掉） |
| 2 | `gorm:before_create` | 触发 `BeforeSave` → `BeforeCreate` |
| 3 | `gorm:save_before_associations` | 保存"前置关联"（先插主表前必须存在的关联） |
| 4 | `gorm:create` | 真正生成并执行 `INSERT` |
| 5 | `gorm:save_after_associations` | 保存"后置关联"（依赖主表主键的关联） |
| 6 | `gorm:after_create` | 触发 `AfterCreate` → `AfterSave`（⚠️ 顺序反直觉，见 5.2） |
| 7 | `gorm:commit_or_rollback_transaction` | 按有无错误提交或回滚 |

> ⚠️ **注意前后钩子的顺序不对称**（源码 `callbacks/create.go` 实证）：
> 前钩子是 `BeforeSave` → `BeforeCreate`，后钩子却是 `AfterCreate` → `AfterSave`。
> 很多人凭直觉写成 `AfterSave → AfterCreate`，这是错的。

> **重要认知**：GORM 的**写操作默认会开事务**。哪怕你只 `Create` 一条记录，它也会 `BEGIN ... COMMIT`。
> 如果追求极致性能且不需要回滚语义，可以全局关掉：`gorm.Config{SkipDefaultTransaction: true}`。
> 但注意：关了之后，批量插入跨多条 `INSERT` 时就不再是原子的了。

---

## 1. 准备工作：模型与连接

你现在的 `CRUD/main.go` 是这样的：

```go
package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 定义模型
type User struct {
	ID   int64
	Name string
	Age  int64
}

func main() {
	dsn := "root:bjy12345@(127.0.0.1:3306)/sql_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&User{})

	u := User{Name: "bai", Age: 19}
	db.Create(&u)
}
```

有几个可以立刻改进的点：

```go
package main

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	ID   int64  `gorm:"primaryKey"`
	Name string `gorm:"size:64;not null"`
	Age  int64
}

func main() {
	dsn := "root:bjy12345@(127.0.0.1:3306)/sql_test?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // 打印每条 SQL，学习阶段必备
		// SkipDefaultTransaction: true,             // 高级：关掉默认事务（先别开）
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal(err)
	}

	u := User{Name: "bai", Age: 19}
	result := db.Create(&u)

	fmt.Println("err          =", result.Error)        // 错误
	fmt.Println("rowsAffected =", result.RowsAffected) // 影响行数
	fmt.Println("u.ID         =", u.ID)                // 主键被回写
}
```

打开 `logger.Info` 后你会看到大致这样的日志，这是理解 GORM 最有效的方式：

```
2026/09/17 15:20:00 /path/main.go:39
[0.812ms] [rows:1] INSERT INTO `users` (`name`,`age`) VALUES ('bai',19)
```

注意它**没有插 `id`** —— 因为 `ID` 为零值且是自增主键，GORM 会自动跳过，交给数据库自增。

---

## 2. 创建记录（Create Record）

### 2.1 Traditional API

```go
user := User{Name: "Jinzhu", Age: 18, Birthday: time.Now()}

result := db.Create(&user) // 通过数据的指针来创建

user.ID             // 返回插入数据的主键
result.Error        // 返回 error
result.RowsAffected // 返回插入记录的条数
```

逐行拆解：

| 表达式 | 类型 | 含义 |
|---|---|---|
| `db.Create(&user)` | `*gorm.DB` | 返回的是**链式调用的 `*gorm.DB` 自身**（克隆版），不是新对象 |
| `user.ID` | `int64` | 插入后主键被**回写**到结构体字段中 |
| `result.Error` | `error` | nil 表示成功 |
| `result.RowsAffected` | `int64` | 单条插入时为 1 |

创建多条记录，直接传切片：

```go
users := []*User{
    {Name: "Jinzhu", Age: 18, Birthday: time.Now()},
    {Name: "Jackson", Age: 19, Birthday: time.Now()},
}

result := db.Create(users) // 传切片 → 插入多行

result.Error        // 返回 error
result.RowsAffected // 返回插入记录的条数
```

> **NOTE（官方原话）**：你无法向 `Create` 传递结构体，因此你应该传入数据的指针。

### 2.2 Generics API（1.30+ 新增）

你的 `go.mod` 是 `gorm.io/gorm v1.31.2`，所以**两种 API 都能用**。Generics API 是 1.30 引入的新写法，特点是类型安全、不用 `db.Model(&User{})` 这种"占位模型"的写法。

```go
user := User{Name: "Jinzhu", Age: 18, Birthday: time.Now()}

// 创建单条记录
ctx := context.Background()
err := gorm.G[User](db).Create(ctx, &user) // 传数据指针

// 需要拿到 result 时，显式声明
result := gorm.WithResult()
err = gorm.G[User](db, result).Create(ctx, &user)

user.ID             // 返回插入数据的主键
result.Error        // 返回 error
result.RowsAffected // 返回插入记录的条数
```

**两者对比：**

| 维度 | Traditional API | Generics API |
|---|---|---|
| 写法 | `db.Create(&user)` | `gorm.G[User](db).Create(ctx, &user)` |
| 类型安全 | 弱（靠反射推断模型） | 强（`G[User]` 编译期确定） |
| 链式查询 | `db.Model(&User{}).Where(...)` | `gorm.G[User](db).Where(...)` |
| `context` | 需 `db.WithContext(ctx)` | 第一个参数直接传 `ctx` |
| 获取 `RowsAffected` | 直接从返回值取 | 需 `gorm.WithResult()` |
| 生态成熟度 | 最成熟、示例最多 | 较新，部分插件/文档未覆盖 |

学习阶段建议**以 Traditional API 为主**（社区资料多、你现有的代码也是这个风格），理解透之后再考虑 Generics。

> Generics API 的 `gorm.G[T]` 是**函数式入口**，后续 `Where`/`Delete`/`Update` 等都挂在它返回的对象上，不能和 `db.Create` 混着链式调用。

---

### 2.3 `ctx context.Context` 到底是什么

#### 它不是 GORM 的东西

`ctx` 来自 **Go 标准库 `context`**，是 Go 里传递"取消信号、超时、请求范围数据"的标准机制。它跟数据库没关系 —— 但**任何可能阻塞的、跨越进程边界的调用都应该接受一个 ctx**，数据库操作正是典型。

`context.Context` 只有四个方法，理解它们就理解了 ctx：

```go
type Context interface {
	Deadline() (deadline time.Time, ok bool)  // 何时超时
	Done() <-chan struct{}                    // 关闭即表示"该停了"
	Err() error                               // 为什么停：Canceled 或 DeadlineExceeded
	Value(key any) any                        // 取透传的值
}
```

三种用途：

| 用途 | 场景 |
|---|---|
| **超时 / 取消** | 查询跑了 5 秒还没回来，主动中断，别把连接池占死 |
| **链路透传** | 把 traceID、当前登录用户、租户 ID 从 HTTP 请求带到 DAO 层 |
| **跨 goroutine 广播** | 一个 `Done()` 通道被关闭，所有派生 ctx 同时收到信号 |

#### GORM 拿它做了什么

Generics API 的每个终结方法都要求传 ctx，内部执行这段逻辑（源码 `generics.go`）：

```go
func (g *g[T]) apply(ctx context.Context) *DB {
	db := g.db
	if !db.DryRun {
		db = db.Session(&Session{NewDB: true, Context: ctx}).getInstance()
	}
	for _, op := range g.ops {
		db = op(db)
	}
	return db
}
```

关键就一行：`Session(&Session{Context: ctx})`。而传统 API 的 `WithContext` 源码是：

```go
func (db *DB) WithContext(ctx context.Context) *DB {
	return db.Session(&Session{Context: ctx})
}
```

**所以 `gorm.G[User](db).Create(ctx, &u)` 和 `db.WithContext(ctx).Create(&u)` 在 ctx 这一点上完全等价。** Generics API 只是把 ctx 从"链式的一个可选环节"提升成了"终结方法必填的第一个参数"，这符合 Go 官方对 ctx 的约定（ctx 应作为函数第一个参数），也顺手避免了"忘记调用 `WithContext`"这个高频 bug。

> `gorm.Open` 时 GORM 给 `Statement.Context` 的默认值是 `context.Background()`。所以传统 API 不调 `WithContext` 也能跑，但**也就失去了超时和取消能力**。

#### 四种典型写法

```go
// ① 最简：脚本/测试，不需要超时和取消
ctx := context.Background()
err := gorm.G[User](db).Create(ctx, &user)

// ② HTTP handler 中（最推荐）：客户端断开连接，数据库查询自动取消
func CreateMovie(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context() // ← 直接复用请求的 ctx
	var req MovieReq
	// ...解析 req...
	m := Movie{Title: req.Title, Duration: req.Duration}
	if err := gorm.G[Movie](db).Create(ctx, &m); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}

// ③ 显式超时：防止单条 SQL 拖垮服务
ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
defer cancel() // ← 必须写！否则定时器泄漏，go vet 会报 lostcancel
err := gorm.G[Movie](db).Where("id = ?", id).Update(ctx, "status", 1)

// ④ 透传值：把 traceID 一路带到 SQL 日志里
ctx = context.WithValue(ctx, traceKey{}, traceID)
```

传统 API 的等价写法：

```go
db.WithContext(ctx).Create(&user)
db.WithContext(ctx).Where("id = ?", id).Update("status", 1)
```

#### 踩坑清单

| 坑 | 说明 |
|---|---|
| 传 `nil` | **别传 nil**。没有上下文语义就用 `context.Background()`；拿不准用 `context.TODO()` |
| 忘记 `defer cancel()` | `WithTimeout`/`WithCancel` 返回的 cancel 不调用会泄漏定时器，`go vet` 报 `lostcancel` |
| 用 ctx 传业务参数 | ctx 的 `Value` 只适合放**请求范围**的元数据（traceID、用户身份），**不要拿它传函数参数** |
| 以为 ctx 会污染 `db` | 不会。ctx 只作用于本次操作，`db` 本身不变 |
| 忘了传统 API 的 `WithContext` | `db.Create(&u)` 默认是 `context.Background()`，超时控制会失效 |
| 超时错误的判断 | 超时是 `context.DeadlineExceeded`，主动取消是 `context.Canceled`，用 `errors.Is` 判断 |

```go
// 正确区分超时与取消
if err := db.WithContext(ctx).Create(&m).Error; err != nil {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		// 超时：可能是慢 SQL 或锁等待，要看日志
	case errors.Is(err, context.Canceled):
		// 用户主动断开，通常忽略即可
	}
}
```

> 补充：`gorm.Config` 里还有个 `DefaultContextTimeout` 字段，可以给所有操作设一个**保底超时**，避免某个忘记传 ctx 的调用无限阻塞。

---

### 2.4 `gorm.WithResult()` 到底是什么

#### 它解决什么问题

Generics API 的 `Create` 签名是：

```go
Create(ctx context.Context, r *T) error   // ← 只返回 error！
```

对比传统 API：`db.Create(&user)` 返回 `*gorm.DB`，上面有 `.Error`、`.RowsAffected`、`.Statement`。

所以 Generics API 会遇到一个矛盾：**方法只能返回一个 error，但你可能还想知道"影响了几行"。**

`WithResult()` 就是解法 —— 它给你一个**结果容器**，你把它一起传进去，执行完从容器里读。

#### 源码剖析（v1.31.x `generics.go`）

```go
// ① 结果容器：注意类型名 result 是【小写未导出】的
type result struct {
	Result       sql.Result  // 底层 database/sql 的原始结果
	RowsAffected int64       // 影响行数
	Error        error       // 错误
}

// ② 关键：把自己挂到 Statement 上
func (info *result) ModifyStatement(stmt *Statement) {
	info.Result = nil        // 先清空
	info.RowsAffected = 0
	info.Error = nil
	stmt.Result = info       // ← 记录的是同一个指针！
}

// ③ 实现 clause.Expression 接口（空实现，只为让类型能通过编译）
func (result) Build(clause.Builder) {}

// ④ 唯一的构造入口
func WithResult() *result {
	return &result{}
}
```

这一小段代码里有四个必须理解的细节：

**细节 1：`result` 是小写未导出的类型。**

你**无法**自己写 `&result{}` —— 编译器根本不认识这个类型。所以 `gorm.WithResult()` 是**唯一**的获取方式。这就是"为什么必须用它"的答案。

**细节 2：返回的是指针，所以是"同一个对象"。**

`WithResult()` 给你一个 `*result`，GORM 内部通过 `stmt.Result = info` 记住的是**这个指针**。操作执行完毕后，值被填进了你手里这个对象 —— 所以你能在调用**之后**读到结果。

**细节 3：它靠 `ModifyStatement` 生效，走的是 `Clauses` 通道。**

```go
func G[T any](db *DB, opts ...clause.Expression) Interface[T] {
	// ...
	v.ops = append(v.ops, func(db *DB) *DB {
		return db.Clauses(opts...)   // ← 你传的 WithResult() 从这里进去
	})
	// ...
}
```

而 `db.Clauses(...)` 遇到实现了这个接口的参数，就会调用它的 `ModifyStatement`：

```go
type StatementModifier interface {
	ModifyStatement(*Statement)
}
```

所以 `gorm.G[User](db, result)` 的效果就是：**在语句执行前，把 `result` 指针注册到 `Statement.Result` 上。**

**细节 4：三个字段分别由谁填写。** 查源码 `callbacks/create.go`：

```go
// MySQL 等走 ExecContext 分支
result, err := db.Statement.ConnPool.ExecContext(
	db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...,
)
if err != nil {
	db.AddError(err)
	return
}

db.RowsAffected, _ = result.RowsAffected()

if db.Statement.Result != nil {
	db.Statement.Result.Result = result              // ← 填 Result
	db.Statement.Result.RowsAffected = db.RowsAffected // ← 填 RowsAffected
}
```

错误则由 `gorm.go` 的 `AddError` 填写：

```go
func (db *DB) AddError(err error) error {
	if err != nil {
		// ...TranslateError 处理...
		if db.Statement != nil && db.Statement.Result != nil {
			db.Statement.Result.Error = db.Error     // ← 填 Error
		}
	}
	return db.Error
}
```

#### 三个字段的含义

| 字段 | 类型 | 含义 |
|---|---|---|
| `Error` | `error` | 本次操作的错误，`nil` 表示成功 |
| `RowsAffected` | `int64` | 影响行数，单条插入为 `1` |
| `Result` | `sql.Result` | 底层 `database/sql` 的原始结果，可调 `LastInsertId()` / `RowsAffected()` |

#### 完整用法

```go
result := gorm.WithResult()
err := gorm.G[User](db, result).Create(ctx, &user)
if err != nil {
	return err
}

fmt.Println("RowsAffected =", result.RowsAffected) // 1

// 想拿数据库层面的自增 ID（一般用不到，user.ID 已经有了）
if result.Result != nil {
	lastID, err := result.Result.LastInsertId()
	if err == nil {
		fmt.Println("LastInsertId =", lastID)      // 与 user.ID 一致
	}
}
```

#### 传统 API 等价写法对照

| 需求 | Traditional API | Generics API |
|---|---|---|
| 只要错误 | `err := db.Create(&u).Error` | `err := gorm.G[User](db).Create(ctx, &u)` |
| 错误 + 行数 | `r := db.Create(&u)` → `r.Error` / `r.RowsAffected` | `r := gorm.WithResult()` → `gorm.G[User](db, r).Create(ctx, &u)` |
| 底层 `sql.Result` | 返回值上没有直接暴露（`Statement.Result` 同样只由 `WithResult` 设置） | `r.Result` |

#### ⚠️ 四个坑

**坑 1：`result.Result` 在 PostgreSQL / SQLite 上可能是 `nil`。**

因为这两个库走 `RETURNING` 分支，那条路径**只填了 `RowsAffected`，没有填 `Result`**：

```go
if db.Statement.Result != nil {
	db.Statement.Result.RowsAffected = db.RowsAffected   // 注意：没有 .Result
}
```

所以调用 `result.Result.LastInsertId()` 之前**必须判空**。你用 MySQL，走的是 `ExecContext` 分支，`Result` 有值。

**坑 2：每次操作新建一个 `WithResult()`，不要复用同一个指针。**

`ModifyStatement` 会把它重置（`Result = nil; RowsAffected = 0; Error = nil`），如果多个 goroutine 共用一个，就是数据竞争。

```go
// ❌ 错误：跨并发复用
var sharedResult = gorm.WithResult()

// ✅ 正确：每次操作现场创建
r := gorm.WithResult()
```

**坑 3：批量插入时 `LastInsertId()` 只返回第一条。**

这是 MySQL 的语义（`sql.Result` 层面），不是 GORM 的问题。但批量插入**不需要**靠它拿主键 —— GORM 会把每条的主键逐条回写到切片元素里，`RowsAffected` 也是总条数。

**坑 4：只是想判断成功与否，别用 `WithResult()`。**

`Create` 已经返回 `error` 了，`err != nil` 就够了。多此一举只会让代码更啰嗦。

#### 自己看源码的位置

想看这些实现，不用上网，本机就有：

```
Windows:  %USERPROFILE%\go\pkg\mod\gorm.io\gorm@v1.31.2\generics.go
          %USERPROFILE%\go\pkg\mod\gorm.io\gorm@v1.31.2\callbacks\create.go
          %USERPROFILE%\go\pkg\mod\gorm.io\gorm@v1.31.2\gorm.go
macOS/Linux:  ~/go/pkg/mod/gorm.io/gorm@v1.31.2/
```

---

### 2.5 为什么"必须传指针"

这是新手最容易懵的一点，原理有三层：

1. **反射需要可寻址（addressable）的值**才能读写字段。传 `struct` 值进去，GORM 拿到的是副本，写不回去。
2. **主键回写**：`INSERT` 之后数据库生成了 `id`，GORM 需要把 `id` 塞回你的变量里 —— 只有指针才能改到你原来的变量。
3. **切片**：`db.Create(users)` 传的是 `[]*User` 或 `[]User` 都可以，因为切片底层有指向数组的指针；但推荐 `[]*User` 或对 `[]User` 传 `&users`，语义更明确。

```go
// ❌ 危险：值传递，主键不会被回写（编译能过，但行为不符合预期）
u := User{Name: "bai"}
db.Create(u)

// ✅ 正确
u := User{Name: "bai"}
db.Create(&u)
```

> 另外注意：`Create` 接收的如果是**单个结构体指针**，那它代表"一条记录"；如果是**切片**，代表"多条记录"。GORM 靠反射类型来判断，不要传 `*[]User` 之外的奇怪组合。

---

## 3. 用指定的字段创建记录（Select / Omit）

### Select：只插入指定字段

```go
db.Select("Name", "Age", "CreatedAt").Create(&user)
// INSERT INTO `users` (`name`,`age`,`created_at`) VALUES ("jinzhu", 18, "2020-07-04 11:05:21.775")
```

### Omit：插入时忽略指定字段

```go
db.Omit("Name", "Age", "CreatedAt").Create(&user)
// INSERT INTO `users` (`birthday`,`updated_at`) VALUES ("2020-01-01 00:00:00.000", "2020-07-04 11:05:21.775")
```

**典型使用场景：**

```go
// 场景：请求体里有 Role 字段，但不允许用户自己注册成 admin
db.Omit("Role").Create(&user)

// 场景：只想插入关键字段，让数据库默认值兜底
db.Select("Name", "Age").Create(&user)

// 场景：忽略所有关联
db.Omit(clause.Associations).Create(&user)
```

> ⚠️ 注意 `Select` / `Omit` 是**链式作用域**，只影响本次调用，不会污染全局 `db`。
> 因为 GORM 的 `db` 是"会话克隆"模式：`db.Select(...)` 返回的是新对象，原 `db` 不变。

---

## 4. 批量插入

### 4.1 基础用法

```go
var users = []User{{Name: "jinzhu1"}, {Name: "jinzhu2"}, {Name: "jinzhu3"}}
db.Create(&users)

for _, user := range users {
    user.ID // 1,2,3  ← 主键被逐条回写
}
```

生成的 SQL（MySQL）：

```sql
INSERT INTO `users` (`name`) VALUES ('jinzhu1'),('jinzhu2'),('jinzhu3')
```

官方说明的三个关键点：

1. **一条 SQL 插入所有数据** → 高效。
2. **返回所有主键值** → 主键会逐条回写到切片元素里。
3. **触发 Hook** → 每个元素都会走一遍钩子。
4. 当记录需要被拆成多个批次时，**GORM 会开启一个事务**来保证整体原子性。

### 4.2 CreateInBatches：指定批次大小

```go
var users = []User{{Name: "jinzhu_1"}, /* ... */, {Name: "jinzhu_10000"}}

// 每批 100 条
db.CreateInBatches(users, 100)
```

为什么要分批？因为 MySQL 对单条 SQL 有 `max_allowed_packet` 限制，一条 SQL 里塞 10 万行会直接报错。分批是**生产环境的常规操作**。

### 4.3 全局配置 CreateBatchSize

```go
db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{
	CreateBatchSize: 1000,
})

// 或者只在某次会话里生效
db = db.Session(&gorm.Session{CreateBatchSize: 1000})

users = [5000]User{{Name: "jinzhu", Pets: []Pet{pet1, pet2, pet3}} /* ... */}

db.Create(&users)
// INSERT INTO users xxx (5 batches)
// INSERT INTO pets  xxx (15 batches)
```

> **注意（官方原话）**：使用 `CreateBatchSize` 选项初始化 GORM 实例后，此后进行创建和关联操作时所有的 `INSERT` 行为都会遵循初始化时的配置。
> 也就是说，**关联的批量插入也会用同一个批次大小**（上面例子中 pets 被分成 15 批）。

### 4.4 关于主键是怎么批量回写的

| 数据库 | 机制 |
|---|---|
| PostgreSQL / SQLite（3.35+） | `INSERT ... RETURNING id`，一次拿回所有主键 |
| MySQL | 不支持 `RETURNING`，用 `LAST_INSERT_ID()` 拿到首个主键，再结合自增步长推导后续主键 |

这就解释了为什么 **MySQL 下批量插入依赖自增主键**。如果你的主键是 UUID（非自增），GORM 只能保证你**自己在代码里赋值**，无法回写。

> 正是这一点，解释了你在 GoFilm 项目里"影院/影厅用 UUID、影片/场次用自增 INT"的选择在批量插入场景下的差异：自增 ID 的批量插入能自动回写、UUID 的必须自己生成（通常放在 `BeforeCreate` 钩子里）。

---

## 5. 创建钩子（Hooks）

### 5.1 钩子是什么

**官方定义**：Hook 是在创建、查询、更新、删除等操作**之前、之后**调用的函数。

**本质上的理解**：钩子**不是你需要注册的东西，而是你给模型实现的方法**。GORM 在执行操作时，用反射检查你的模型有没有实现某个特定接口，实现了就在对应时机自动调用它。

签名是固定的：

```go
func (u *User) BeforeCreate(tx *gorm.DB) error
//       │              │            └── 返回 error，非 nil 则中止操作
//       │              └── 同一个事务会话，可在里面继续操作数据库
//       └── 接收者必须是指针
```

**放到回调链里看**（呼应第 0 节）：`gorm:before_create` 这个回调实际做的事，就是"找出你模型上实现了 `BeforeSave`/`BeforeCreate` 的方法并调用"。所以：

```
回调链:  begin_transaction → before_create → ... →  INSERT  → ... → after_create → commit
                                        ↑                              ↑
                                    你的 BeforeSave/               你的 AfterCreate/
                                    BeforeCreate                   AfterSave
```

**一句话：钩子就是 GORM 回调链留给你的"插入点"。** 这也是为什么钩子能、`Scopes` 也能 —— 它们都在回调层。

### 5.2 钩子有什么用（为什么要用它）

钩子解决的核心问题是：**"数据写入时必须发生的事，应该写在哪里？"**

如果全写在 service 层，你会在几十个地方重复写"生成 UUID、大写转小写、写审计日志……"，而且**总有一天会漏掉一处**。钩子把这类规则**收敛到模型自己身上** —— 模型是什么样的，它自己最清楚。

五类典型用途：

| 用途 | 例子 | 常用钩子 |
|---|---|---|
| **① 补全 / 规范化数据** | 生成 UUID、trim 空格、邮箱转小写、密码加密、算总价 | `BeforeCreate` / `BeforeSave` |
| **② 业务约束校验** | 角色合法性、金额非负、状态机流转校验 | `BeforeCreate` / `BeforeUpdate` / `BeforeDelete` |
| **③ 审计与日志** | 记录"谁在什么时候创建/修改了什么" | `AfterCreate` / `AfterUpdate` |
| **④ 同一事务内的连带写** | 评论入库后更新影片评论数；插入后写统计表 | `AfterCreate` / `AfterDelete` |
| **⑤ 读取时的默认值 / 脱敏** | 空头像补默认图、空标签补空数组 | `AfterFind` |

**钩子最大的价值是"隐式保证正确性"。** 因为它在事务内、且拦截一切写入路径（不管是 service、脚本还是后台任务触发的），所以不会有漏网之鱼。

> ⚠️ **但别滥用，这是双刃剑。**
> 钩子是**隐式**的 —— 别人读 `db.Create(&u)` 这行代码，看不到背后还写了审计日志、还改了统计表。
> **判断标准：写"数据本身的规则"，不要写"业务流程"。**
> - ✅ 适合：UUID 必须自动生成、密码必须加密、金额不能为负 —— 这是数据自身的性质
> - ❌ 不适合：下单后发短信、注册后发优惠券 —— 这是业务流程，应该在 service 层显式调用，让人一眼看到

### 5.3 全部钩子接口一览

签名统一为 `func(*gorm.DB) error`。按操作分类：

| 钩子 | 触发时机 | 说明 |
|---|---|---|
| `BeforeSave` | 保存前 | **Create 和 Update 都会触发** |
| `BeforeCreate` | 插入前 | 紧跟在 `BeforeSave` 之后 |
| `AfterCreate` | 插入后 | 此时**主键已经回写**，事务尚未提交 |
| `AfterSave` | 保存后 | **Create 和 Update 都会触发** |
| `BeforeUpdate` | 更新前 | |
| `AfterUpdate` | 更新后 | |
| `BeforeDelete` | 删除前 | 可以做"不允许删除"的拦截 |
| `AfterDelete` | 删除后 | |
| `AfterFind` | 查询后 | **只有 After，没有 `BeforeFind`** |

三个容易被忽略的点：

1. **`BeforeSave` / `AfterSave` 是"共享钩子"** —— 只要你实现了它们，插入和更新都会触发。所以如果你只想在创建时做某件事，用 `BeforeCreate` 而不是 `BeforeSave`。
2. **没有 `BeforeFind`**。想干预查询条件要用 `Scopes` 或 `Session`，不能靠钩子。
3. **`AfterFind` 对每条记录都会调用一次**，批量查 1000 条就跑 1000 次，里面绝不能写数据库查询。

### 5.4 执行顺序（重点，极易记错）

官方文档给出的完整顺序（<https://gorm.io/zh_CN/docs/hooks.html>）：

**创建对象：**

```
// 开始事务
BeforeSave
BeforeCreate
// 关联前的 save
// 插入记录至 db
// 关联后的 save
AfterCreate
AfterSave
// 提交或回滚事务
```

**更新对象：**

```
// 开始事务
BeforeSave
BeforeUpdate
// 关联前的 save
// 更新 db
// 关联后的 save
AfterUpdate
AfterSave
// 提交或回滚事务
```

**删除对象：**

```
// 开始事务
BeforeDelete
// 删除 db 中的数据
AfterDelete
// 提交或回滚事务
```

**查询对象：**

```
// 从 db 中加载数据
// Preloading (eager loading)
AfterFind
```

#### ⚠️ 一个必踩的坑：后钩子的顺序是反的

请注意对比：

| | 前钩子 | 后钩子 |
|---|---|---|
| **创建** | `BeforeSave` → `BeforeCreate` | `AfterCreate` → `AfterSave` |
| **更新** | `BeforeSave` → `BeforeUpdate` | `AfterUpdate` → `AfterSave` |

**前面是 `Save` 在前，后面却是 `Create`/`Update` 在前 —— 不对称！**

很多人（包括不少中文教程）会凭直觉写成 `AfterSave → AfterCreate`，**这是错的**。源码 `callbacks/create.go` 实证：

```go
// BeforeCreate 回调：先 Save 后 Create
func BeforeCreate(db *gorm.DB) {
	if db.Error == nil && db.Statement.Schema != nil && !db.Statement.SkipHooks &&
		(db.Statement.Schema.BeforeSave || db.Statement.Schema.BeforeCreate) {
		callMethod(db, func(value interface{}, tx *gorm.DB) (called bool) {
			if db.Statement.Schema.BeforeSave {
				if i, ok := value.(BeforeSaveInterface); ok {
					called = true
					db.AddError(i.BeforeSave(tx))
				}
			}
			if db.Statement.Schema.BeforeCreate {
				if i, ok := value.(BeforeCreateInterface); ok {
					called = true
					db.AddError(i.BeforeCreate(tx))
				}
			}
			return called
		})
	}
}

// AfterCreate 回调：先 Create 后 Save  ← 注意这里和上面相反
func AfterCreate(db *gorm.DB) {
	if db.Error == nil && db.Statement.Schema != nil && !db.Statement.SkipHooks &&
		(db.Statement.Schema.AfterSave || db.Statement.Schema.AfterCreate) {
		callMethod(db, func(value interface{}, tx *gorm.DB) (called bool) {
			if db.Statement.Schema.AfterCreate {          // ← Create 在前
				if i, ok := value.(AfterCreateInterface); ok {
					called = true
					db.AddError(i.AfterCreate(tx))
				}
			}
			if db.Statement.Schema.AfterSave {            // ← Save 在后
				if i, ok := value.(AfterSaveInterface); ok {
					called = true
					db.AddError(i.AfterSave(tx))
				}
			}
			return called
		})
	}
}
```

**记忆口诀**：前面是"Save 包住 Create"（Save 先），后面是"Create 先跑完再轮到 Save"。

> 实践建议：**如果两件事没有强依赖关系，就不要同时依赖这两个钩子**。真有顺序要求，用一个 `AfterSave` 里按操作类型分支判断，比依赖隐式顺序可靠。

### 5.5 应用实例

以下例子都用你的 GoFilm 场景。

#### 实例 1：自动生成 UUID（对应影院 / 影厅）

```go
import "github.com/google/uuid"

type Cinema struct {
	ID   string `gorm:"type:char(36);primaryKey"`
	Name string
}

func (c *Cinema) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {          // ← 判断空，避免覆盖手工指定的值
		c.ID = uuid.NewString()
	}
	return nil
}
```

**为什么要判空？** 因为有些场景（数据迁移、测试）你会想手工指定 ID。不判空就永远覆盖不了。

**为什么不用 `gorm:"default:uuid_generate_v3()"`？**
- 那是**数据库函数**，换了数据库（SQLite 测试环境）就不认了
- Go 侧生成的值在**插入前**就能拿到，比如你要拼 `"/cinemas/" + c.ID` 的返回体，或者要在日志里打出来
- 测试时可控（可以注入固定 UUID）

> 这正是你在 GoFilm 里"影院/影厅用 UUID"的落地方式 —— 而 UUID 主键的代价也在这里暴露：**批量插入时 GORM 无法回写主键**（不像自增 INT，可以靠 `LAST_INSERT_ID()` 推导），所以必须在 `BeforeCreate` 里自己生成。

#### 实例 2：数据规范化 + 业务校验

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// ① 规范化：这类"清洗"规则散落在各处最容易漏，放这里最合适
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))
	u.Phone = strings.TrimSpace(u.Phone)

	// ② 校验：不满足就 return error，GORM 会中止插入并回滚整个事务
	if u.Email == "" {
		return errors.New("邮箱不能为空")
	}
	if u.Age < 0 || u.Age > 150 {
		return fmt.Errorf("年龄 %d 不合法", u.Age)
	}

	// ③ 密码加密：保证"任何路径写入的密码都是加密的"
	if u.Password != "" && !strings.HasPrefix(u.Password, "$2a$") {
		hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		u.Password = string(hash)
	}
	return nil
}
```

**要点**：返回 `error` 不是"记个日志"，而是**真的中断**。官方原话：

> 在 GORM 中保存、删除操作会默认运行在事务上，因此在事务完成之前该事务中所作的更改是不可见的，如果您的钩子返回了任何错误，则修改将被回滚。

#### 实例 3：审计日志（同一事务内，不会写歪）

```go
type AuditLog struct {
	ID        uint
	Table     string
	RecordID  string
	Action    string
	CreatedAt time.Time
}

func (o *Order) AfterCreate(tx *gorm.DB) error {
	// 用 tx！这条日志和订单在同一个事务里 —— 订单回滚，日志一起回滚
	return tx.Create(&AuditLog{
		Table:     "orders",
		RecordID:  fmt.Sprint(o.ID),   // ← AfterCreate 时主键已经回写好了
		Action:    "create",
		CreatedAt: time.Now(),
	}).Error
}
```

**关键点**：`AfterCreate` 能安全地读 `o.ID`，因为回调链里 INSERT 和**主键回写**都发生在 `after_create` 之前（见第 0 节的回调顺序表）。换成 `BeforeCreate` 就读不到主键了。

#### 实例 4：连带写统计（评论数）

```go
func (r *Review) AfterCreate(tx *gorm.DB) error {
	return tx.Model(&Movie{}).
		Where("id = ?", r.MovieID).
		UpdateColumn("review_count", gorm.Expr("review_count + 1")).Error
}

func (r *Review) AfterDelete(tx *gorm.DB) error {
	return tx.Model(&Movie{}).
		Where("id = ?", r.MovieID).
		UpdateColumn("review_count", gorm.Expr("review_count - 1")).Error
}
```

**为什么用 `UpdateColumn` 而不是 `Update`？**

| | `Update` / `Updates` | `UpdateColumn` / `UpdateColumns` |
|---|---|---|
| 触发 Hook | ✅ 会触发 | ❌ 不触发 |
| 更新 `UpdatedAt` | ✅ 会 | ❌ 不会 |
| 适用 | 常规业务更新 | 计数、内部字段修正 |

在钩子里用 `Update` 有**递归风险**：更新 → 触发 `BeforeUpdate` → 钩子里又更新 → …… 而且会莫名其妙改掉 `UpdatedAt`（"只是加了个评论数，怎么影片的更新时间变了？"）。**用 `UpdateColumn` 规避。**

#### 实例 5：删除拦截（业务约束）

```go
// 还有未完成订单的影院不允许删除
func (c *Cinema) BeforeDelete(tx *gorm.DB) error {
	var count int64
	if err := tx.Model(&Order{}).
		Where("cinema_id = ? AND status = ?", c.ID, "pending").
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该影院还有未完成的订单，不能删除")
	}
	return nil
}
```

**价值**：这条约束作用于**所有**删除路径 —— admin 后台、脚本、命令行工具，全都拦得住。写在 service 层就做不到这么彻底。

（注意 `tx` 上的条件是干净的：官方原话"`tx` 是带有 `NewDB` 选项的新会话模式，基于 `tx` 的操作会在同一个事务中，但不会带上任何当前的条件"，所以这里 `Where("cinema_id = ?")` 不会和删除影院的条件纠缠。）

#### 实例 6：`AfterFind` 补默认值（让前端省心）

```go
func (m *Movie) AfterFind(tx *gorm.DB) error {
	if m.Poster == "" {
		m.Poster = "/static/img/default-poster.png"
	}
	if m.Tags == nil {
		m.Tags = []Tag{}          // ← 序列化成 JSON 是 []，不是 null
	}
	return nil
}
```

**好处**：前端不用到处写 `poster || 'default.png'`；JSON 序列化也不会出现 `"tags": null` 这种让前端崩溃的情况。

> ⚠️ **`AfterFind` 里绝对不要写数据库查询**。它**对每条记录都执行一次**，查 100 条电影就跑 100 次 —— 这就是教科书级的 **N+1**。

#### 实例 7（进阶）：在钩子里修改当前操作

官方还演示了一个能力 —— 钩子里可以改 `tx.Statement`，从而影响本次操作本身：

```go
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// 强制只插入 Name 和 Age，忽略其他字段
	tx.Statement.Select("Name", "Age")

	// 动态加一个"冲突就跳过"的子句
	tx.Statement.AddClause(clause.OnConflict{DoNothing: true})

	// tx 是带 NewDB 的新会话：在同一事务里，但不带当前累积的条件
	var role Role
	return tx.First(&role, "name = ?", u.Role).Error
}
```

这在"根据运行时条件动态调整本次写入行为"时很有用，但可读性差，**知道有这回事即可，别当常规手段用**。

### 5.6 使用钩子的注意事项

| # | 注意点 | 说明 |
|---|---|---|
| 1 | 接收者必须是**指针** | `func (u *User)` 而不是 `func (u User)`，否则拿不到也改不了字段 |
| 2 | 返回值命名可省 `return nil` | 写 `(err error)` 就用裸 `return`；否则结尾必须 `return nil` |
| 3 | **必须用参数 `tx`，不用外层 `db`** | 用 `db` 会脱离当前事务，回滚时数据不一致 |
| 4 | `tx` 是 `NewDB` 会话 | 在同一事务里，但**不带当前累积的查询条件** |
| 5 | `SkipHooks` 可整体跳过 | `DB.Session(&gorm.Session{SkipHooks: true}).Create(&u)`，适合迁移脚本 |
| 6 | **Map 创建不触发钩子** | `db.Model(&User{}).Create(map[string]any{...})` —— 没有模型方法可调 |
| 7 | 批量插入**每个元素**都触发 | `db.Create(&users)` 时钩子逐条执行 |
| 8 | 钩子里的重查询 = N+1 | 尤其 `AfterFind`，务必警惕 |
| 9 | 别用钩子写业务流程 | 隐式逻辑是债务。数据规则放钩子，业务流程放 service |
| 10 | 注意递归风险 | 钩子里对同表 `Update` 会再触发 `BeforeUpdate`；用 `UpdateColumn` 规避 |
| 11 | 钩子**执行在事务内** | 所以它能看到未提交的数据，也能被一起回滚 |
| 12 | 钩子抛 error 会中断整个操作 | 这是特性不是 bug；不想中断就别返回 error |

**第 3 条的完整对照：**

```go
// ✅ 正确：用 tx，参与同一事务
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Role == "" {
		u.Role = "user"
	}
	return tx.Model(&Counter{}).Where("id = ?", 1).
		Update("count", gorm.Expr("count + ?", 1)).Error
}

// ❌ 错误：用外层 db 会脱离当前事务
//    一旦后续步骤失败回滚，这里加过的 count 不会一起回滚，造成数据不一致
func (u *User) BadBeforeCreate(tx *gorm.DB) error {
	return db.Model(&Counter{}).Where("id = ?", 1).
		Update("count", gorm.Expr("count + ?", 1)).Error
}
```

**为什么"脱离事务"？** 一句话：`*gorm.DB` 只是个**会话句柄**，真正决定"走哪条连接"的是它内部 `Statement.ConnPool`。
事务里那个 `ConnPool` 是 `*sql.Tx`，而全局 `db` 的 `ConnPool` 是连接池 `*sql.DB` —— 所以钩子里的 `db.Xxx()` 会**另取一条连接、语句级自动提交**。

> 📌 完整拆解（含源码、连接 ID 实测、SQL 日志对照、修复方式）见
> **[`GORM答疑记录.md`](./GORM答疑记录.md) → Q1**。

### 5.7 SkipHooks：跳过钩子

```go
DB.Session(&gorm.Session{SkipHooks: true}).Create(&user)
DB.Session(&gorm.Session{SkipHooks: true}).Create(&users)
DB.Session(&gorm.Session{SkipHooks: true}).CreateInBatches(users, 100)
```

> `SkipHooks` 是**会话级**的，只影响这一条链式调用，之后 `DB` 本身不受影响。
> 典型场景：数据迁移 / 初始化脚本，不想让业务钩子介入（比如不想为 10 万条历史数据逐条写审计日志）。

---

## 6. 根据 Map 创建

GORM 支持 `map[string]interface{}` 和 `[]map[string]interface{}{}`：

```go
db.Model(&User{}).Create(map[string]interface{}{
	"Name": "jinzhu", "Age": 18,
})

// 从 []map[string]interface{}{} 批量插入
db.Model(&User{}).Create([]map[string]interface{}{
	{"Name": "jinzhu_1", "Age": 18},
	{"Name": "jinzhu_2", "Age": 20},
})
```

> **注意（官方原话）**：当使用 map 来创建时，**钩子方法不会执行，关联不会被保存，且不会回写主键**。

**为什么？** 因为 map 没有"字段地址"，GORM 无法把结果写回去，也无法调用 `func (u *User) BeforeCreate` 这类方法 —— 它根本不知道你要映射到哪个结构体，只知道 `Model(&User{})` 给出的表名。

**什么时候用 map？**

- 动态字段插入（字段集合在运行时才能确定）
- 需要**绕过 `default` tag** 写入零值（见下方对照）
- ⚠️ 不需要为了"写零值"而改 map —— 没有 `default` tag 的字段，struct 本来就会写零值

```go
// ── 情况一：Age 没有 default tag ── 两种写法行为完全一样，零值都会被写入
db.Create(&User{Name: "bai", Age: 0})
// INSERT INTO `users` (`name`,`age`) VALUES ('bai',0)

db.Model(&User{}).Create(map[string]interface{}{"Name": "bai", "Age": 0})
// INSERT INTO `users` (`name`,`age`) VALUES ('bai',0)

// ── 情况二：Age 带 default tag（`gorm:"default:18"`）── 这里才分出胜负
db.Create(&User{Name: "bai", Age: 0})
// INSERT INTO `users` (`name`,`age`) VALUES ('bai',18)  ← 零值被默认值顶掉，你写不进 0

db.Model(&User{}).Create(map[string]interface{}{"Name": "bai", "Age": 0})
// INSERT INTO `users` (`name`,`age`) VALUES ('bai',0)   ← map 完全绕过 default tag，0 原样写入
```

> **⚠️ 勘误（2026-09-19 实测更正）**
>
> 本节早前写成「struct 会把零值跳过，所以 INSERT 里没有 age 列」—— **这是错的**。
> 实测（`事务与会话/main.go` 场景 6 / 7 / 8）结论是：
>
> | 字段情况 | struct 传零值 | 结果 |
> |---|---|---|
> | 无 `default` tag | `Age: 0` | **零值照写**：`(name,age) VALUES ('bai',0)` |
> | 字面量 `default:18` | `Age: 0` | 换成 **18**（写不进 0） |
> | 数据库函数 `default:uuid_generate_v3()` | `Age: ""` | 该列**从 INSERT 省略**，交给数据库算 |
>
> 所以「列被跳过」只发生在**默认值由数据库函数产生**的情况下。
> struct 与 map 的真正差异也不是零值，而是：
> **map 创建不触发钩子、不保存关联、不回写主键，也不参与 `default` tag 替换**。

---

## 7. 使用 SQL 表达式 / Context Valuer 创建

有些值必须由数据库函数计算，不能在 Go 侧算好（例如 GIS 的 `ST_PointFromText`）。GORM 提供两种方法。

### 方法一：`map` + `clause.Expr`

```go
db.Model(User{}).Create(map[string]interface{}{
	"Name": "jinzhu",
	"Location": clause.Expr{SQL: "ST_PointFromText(?)", Vars: []interface{}{"POINT(100 100)"}},
})
// INSERT INTO `users` (`name`,`location`) VALUES ("jinzhu",ST_PointFromText("POINT(100 100)"));
```

### 方法二：自定义数据类型实现 `GormValue`

```go
type Location struct {
	X, Y int
}

// Scan 实现 sql.Scanner 接口（从数据库读出时用）
func (loc *Location) Scan(v interface{}) error {
	// Scan a value into struct from database driver
}

// 告诉 GORM 这个字段在数据库里的类型
func (loc Location) GormDataType() string {
	return "geometry"
}

// 告诉 GORM 写入时应该用什么 SQL 表达式
func (loc Location) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	return clause.Expr{
		SQL:  "ST_PointFromText(?)",
		Vars: []interface{}{fmt.Sprintf("POINT(%d %d)", loc.X, loc.Y)},
	}
}

type User struct {
	Name     string
	Location Location
}

db.Create(&User{
	Name:     "jinzhu",
	Location: Location{X: 100, Y: 100},
})
// INSERT INTO `users` (`name`,`location`) VALUES ("jinzhu",ST_PointFromText("POINT(100 100)"))
```

这里的 `?` 是占位符，`Vars` 里的值会以**参数化**方式传入 —— 这点很重要，是防 SQL 注入的正确姿势。

> **对比记忆：**
> - 只是想在某次 INSERT 里用 SQL 函数 → 用 `clause.Expr` / `gorm.Expr`
> - 想让某个字段**永远**以 SQL 函数方式写入 → 实现 `gorm.Valuer` 接口
>
> 日常最常用的其实是 `gorm.Expr`，例如：`db.Model(&User{}).Create(map[string]interface{}{"age": gorm.Expr("age + 1")})`。

---

## 8. 高级选项

### 8.1 关联创建（Associations）

创建关联数据时，**如果关联值非零，这些关联会被 upsert，并且它们的 Hooks 方法也会被调用**。

```go
type CreditCard struct {
	gorm.Model
	Number string
	UserID uint
}

type User struct {
	gorm.Model
	Name       string
	CreditCard CreditCard
}

db.Create(&User{
	Name:       "jinzhu",
	CreditCard: CreditCard{Number: "411111111111"},
})
// INSERT INTO `users` ...
// INSERT INTO `credit_cards` ...
```

用 `Select` / `Omit` 跳过关联更新：

```go
// 只跳过指定的关联
db.Omit("CreditCard").Create(&user)

// 跳过所有关联
db.Omit(clause.Associations).Create(&user)
```

**关键点：**

- GORM 会自动先插 `users`，拿到 `user.ID`，再把它填到 `credit_cards.user_id` 上，最后插 `credit_cards`。这就是前面回调链里 `save_before_associations` / `save_after_associations` 的作用。
- **"关联值非零才会被保存"** —— 如果你 `CreditCard: CreditCard{}`（全零），GORM 会视为"没有这个关联"，不会插入。
- 关联的插入会触发**关联类型自己的钩子**（`CreditCard.BeforeCreate` 也会被调用）。
- 用 `FullSaveAssociations` 模式可以强制 upsert 关联的所有字段：

```go
db.Session(&gorm.Session{FullSaveAssociations: true}).Create(&user)
```

**在你的 GoFilm 里的典型应用：** 一次下单同时写 `orders` + `order_items` + 锁座，就可以靠关联创建 + 事务一次搞定。

---

### 8.2 默认值（default tag）

#### 基础用法

```go
type User struct {
	ID   int64
	Name string `gorm:"default:galeone"`
	Age  int64  `gorm:"default:18"`
}
```

**`default` tag 有两重语义，作用点完全不同：**

1. **建表时** —— `AutoMigrate` 给该列生成 `DEFAULT` 约束（数据库层的兜底）。
2. **插入时** —— GORM 在 **Go 侧**就把默认值填进去了：只要字段是零值，就用 `default` 里写的字面量替换。

第 2 点最容易被误解（本节早前版本就写错了，已更正）。实测（`事务与会话/main.go` 场景 7b）：

```go
type DemoDefault struct {
	ID     uint
	Name   string `gorm:"default:anon"`
	Age    int64  `gorm:"default:18"`
	AgePtr *int64 `gorm:"default:18"`
}

d := DemoDefault{}
db.Create(&d)
// INSERT INTO `demo_defaults` (`name`,`age`,`age_ptr`) VALUES ('anon',18,18)

// ⚠️ GORM 还会把默认值「回写」到你传进去的结构体上
fmt.Printf("%q %d\n", d.Name, d.Age) // "anon" 18
```

对应源码（`callbacks/create.go`，`ConvertToCreateValues` 的 struct 分支）：

```go
for idx, column := range values.Columns {
	field := stmt.Schema.FieldsByDBName[column.Name]
	if values.Values[0][idx], isZero = field.ValueOf(stmt.Context, stmt.ReflectValue); isZero {
		if field.DefaultValueInterface != nil {
			values.Values[0][idx] = field.DefaultValueInterface                    // 零值 → 换成默认值
			stmt.AddError(field.Set(stmt.Context, stmt.ReflectValue,               // 并写回你的结构体
				field.DefaultValueInterface))
		} else if field.AutoCreateTime > 0 || field.AutoUpdateTime > 0 {
			// CreatedAt / UpdatedAt 也是这么自动填的
		}
	}
}
```

> **踩坑提醒**：`db.Create(&d)` 之后 `d` 已经**不再是你传进去的那个值**了。
> 写单测、做断言、或者 Create 后紧接着用这个对象拼响应时，都要留意。

#### ⚠️ 零值陷阱（重点）

> **注意（官方原话）**：当结构体的字段默认值是零值的时候，比如 `0`、`''`、`false`，这些字段值将不会被保存到数据库中。你可以使用**指针类型**或者 **Scanner/Valuer** 来避免这种情况。

```go
type User struct {
	gorm.Model
	Name   string
	Age    *int           `gorm:"default:18"`
	Active sql.NullBool   `gorm:"default:true"`
}
```

**这个「陷阱」有严格前提：只对带 `default` tag 的字段成立。** 实测对照（场景 7c）：

| 写法（字段带 `gorm:"default:18"`） | 传 `0` 时实际写入 | 不传 / 传 `nil` 时 | 说明 |
|---|---|---|---|
| `Age int64` | **18** | 18 | 零值被默认值顶掉，**你写不进 0** |
| `Age *int64` | **0** | 18（`nil` 是零值） | ✅ 唯一能表达"显式 0"的方案 |
| `Age sql.NullInt64` | **0**（`Valid=true`） | 18（`Valid=false`） | 标准库方案，可读性略差 |
| `db.Select("Age")` | **18** | 18 | ⚠️ **Select 救不了零值**（实测，与很多教程说法相反） |

实测 SQL（`事务与会话/main.go` 场景 7c，`Age: 0, AgePtr: &zero`）：

```sql
INSERT INTO `demo_defaults` (`name`,`age`,`age_ptr`) VALUES ('anon',18,0)
--                                                    ↑ 被顶掉  ↑ 指针保住
```

**一句话记忆（更正版）：**

| 字段是否带 `default` tag | 传零值会怎样 | 你要做什么 |
|---|---|---|
| **带** | 被默认值替换 | 想写 `0` / `''` / `false` → 上指针或 `sql.NullXxx` |
| **不带** | 零值原样写入 | **什么都不用做** |

判断依据在源码里是 `field.DefaultValueInterface != nil`。它 != nil 的条件（`schema/field.go`）是：

```go
skipParseDefaultValue := strings.Contains(field.DefaultValue, "(") && strings.Contains(field.DefaultValue, ")") ||
	strings.ToLower(field.DefaultValue) == "null" || field.DefaultValue == ""
if field.HasDefaultValue && !skipParseDefaultValue {
	field.DefaultValueInterface, err = strconv.ParseInt(field.DefaultValue, 0, 64) // 能解析成 Go 值
}
```

- `default:18`、`default:anon` → **能解析** → `DefaultValueInterface != nil` → 走「Go 侧替换」
- `default:uuid_generate_v3()`（含括号）、`default:null`、空 → **不解析** → 走「省略该列，交给数据库」

这也解释了为什么 `default` 里能写 `gen_random_uuid()` 这类数据库函数 —— 它会被识别成"数据库负责"，而不是当成字符串字面量。

#### 让数据库拥有默认值

> **注意（官方原话）**：若要让字段在数据库中拥有默认值，则必须使用 `default` Tag 来为结构体字段设置默认值。如果想要在数据库迁移的时候跳过默认值，可以使用 `default:(-)`。

```go
type User struct {
	ID        string `gorm:"default:uuid_generate_v3()"` // 数据库函数
	FirstName string
	LastName  string
	Age       uint8
	FullName  string `gorm:"->;type:GENERATED ALWAYS AS (concat(firstname,' ',lastname));default:(-);"`
}
```

这里有两个新东西：

- `default:(-)` —— 迁移时**不生成 DEFAULT 子句**（用于数据库自己算的虚拟列/生成列）。
- `->;` —— **字段权限**：`->` 表示**只读**（只允许读，不允许写），`:<-` 表示只写，`-` 表示完全忽略该字段。虚拟列必须只读，否则 INSERT 会报错。

字段级别的完整权限语法见 [Models → 字段权限](https://gorm.io/zh_CN/docs/models.html#field_permission)。

#### ⚠️ SQLite 的特殊坑

> **注意（官方原话）**：**SQLite 不支持批量插入的时候使用默认值。**

```go
type Pet struct {
	Name string `gorm:"default:cat"`
}

// 在 SQLite 中这是不允许的，所以 GORM 会构建出错误的 SQL 并返回错误：
// INSERT INTO `pets` (`name`) VALUES ("dog"),(DEFAULT) RETURNING `name`
db.Create(&[]Pet{{Name: "dog"}, {}})
```

**替代方案 —— 用钩子补默认值：**

```go
func (p *Pet) BeforeCreate(tx *gorm.DB) (err error) {
	if p.Name == "" {
		p.Name = "cat"
	}
	return
}
```

更多讨论见 [issues#6335](https://github.com/go-gorm/gorm/issues/6335)。

> 好消息：你用 MySQL，不受这个限制。但"用钩子兜默认值"这个技巧本身很值得记住 —— 它比依赖数据库默认值更可控。

---

### 8.3 Upsert 及冲突（OnConflict）

"Upsert" = Update + Insert，即"存在就更新，不存在就插入"。GORM 通过 `clause.OnConflict` 提供了跨数据库的兼容支持。

```go
import "gorm.io/gorm/clause"
```

#### ① 冲突时什么都不做

```go
db.Clauses(clause.OnConflict{DoNothing: true}).Create(&user)
```

#### ② 冲突时更新为指定值

```go
db.Clauses(clause.OnConflict{
	Columns:   []clause.Column{{Name: "id"}},
	DoUpdates: clause.Assignments(map[string]interface{}{"role": "user"}),
}).Create(&users)
// SQL Server:  MERGE INTO "users" USING *** WHEN NOT MATCHED THEN INSERT *** WHEN MATCHED THEN UPDATE SET ***
// MySQL:       INSERT INTO `users` *** ON DUPLICATE KEY UPDATE ***
```

#### ③ 冲突时用 SQL 表达式更新

```go
db.Clauses(clause.OnConflict{
	Columns:   []clause.Column{{Name: "id"}},
	DoUpdates: clause.Assignments(map[string]interface{}{
		"count": gorm.Expr("GREATEST(count, VALUES(count))"),
	}),
}).Create(&users)
// INSERT INTO `users` *** ON DUPLICATE KEY UPDATE `count`=GREATEST(count, VALUES(count));
```

#### ④ 冲突时用"新值"更新指定列

```go
db.Clauses(clause.OnConflict{
	Columns:   []clause.Column{{Name: "id"}},
	DoUpdates: clause.AssignmentColumns([]string{"name", "age"}),
}).Create(&users)
// SQL Server:  MERGE INTO "users" USING *** WHEN NOT MATCHED THEN INSERT *** WHEN MATCHED THEN UPDATE SET "name"="excluded"."name"
// PostgreSQL:  INSERT INTO "users" *** ON CONFLICT ("id") DO UPDATE SET "name"="excluded"."name","age"="excluded"."age"
// MySQL:       INSERT INTO `users` *** ON DUPLICATE KEY UPDATE `name`=VALUES(name),`age`=VALUES(age)
```

#### ⑤ 冲突时更新所有列

```go
db.Clauses(clause.OnConflict{
	UpdateAll: true,
}).Create(&users)
// INSERT INTO "users" *** ON CONFLICT ("id") DO UPDATE SET "name"="excluded"."name","age"="excluded"."age", ...;
// INSERT INTO `users` *** ON DUPLICATE KEY UPDATE `name`=VALUES(name),`age`=VALUES(age), ...;
```

> 注意 `UpdateAll: true` 会更新**除主键、以及带 SQL 函数默认值的列之外**的所有列。

#### 四种模式速查

| 配置 | 语义 | 典型场景 |
|---|---|---|
| `DoNothing: true` | 冲突就跳过 | 幂等导入、去重写入 |
| `DoUpdates: clause.Assignments(map)` | 冲突就改成固定值 / 表达式 | 计数器累加、状态重置 |
| `DoUpdates: clause.AssignmentColumns([]string{...})` | 冲突就用新记录的值覆盖指定列 | 数据同步、增量更新 |
| `UpdateAll: true` | 冲突就用新记录覆盖全部列（除主键） | 全量覆盖、快照写入 |

#### ⚠️ 两个实战坑

**坑 1：MySQL 8.0.20+ 的 `VALUES(col)` 已废弃**

官方示例里 MySQL 生成的 `ON DUPLICATE KEY UPDATE name=VALUES(name)` 在 MySQL 8.0.20+ 会产出 deprecated 警告，新语法是行别名：

```sql
INSERT INTO `users` (id, name) VALUES (1, 'a') AS new
ON DUPLICATE KEY UPDATE name = new.name;
```

用 GORM 时如果看到 warning 不必惊慌，功能仍然正常；要彻底消除可用 `clause.Assignments` + 手写 `gorm.Expr` 指定新语法，或走 Raw SQL。

**坑 2：`OnConflict` 依赖唯一索引**

`Columns: []clause.Column{{Name: "id"}}` 只有在 `id` 上有**主键或唯一索引**时才会生效。如果是业务唯一键（比如 `email`），务必先在模型上声明，否则冲突检测不会按你预期工作：

```go
type User struct {
	ID    int64
	Email string `gorm:"uniqueIndex"` // 必须有唯一索引
	Name  string
}

db.Clauses(clause.OnConflict{
	Columns:   []clause.Column{{Name: "email"}},
	DoUpdates: clause.AssignmentColumns([]string{"name"}),
}).Create(&user)
```

#### 相关：FirstOrInit / FirstOrCreate

如果你只是想"查不到就创建"，不需要真的 upsert，更合适的 API 是：

```go
// 查不到就初始化一个结构体（不落库）
db.FirstOrInit(&user, User{Name: "non_existing"})

// 查不到就创建（会落库）
db.Where(User{Name: "jinzhu"}).Attrs(User{Age: 20}).FirstOrCreate(&user)
```

详见 [高级查询](https://gorm.io/zh_CN/docs/advanced_query.html)。

---

## 9. 坑点速查表

| # | 坑 | 现象 | 解决 |
|---|---|---|---|
| 1 | 传值而非指针 | 主键不回写、数据可能与预期不符 | 始终 `db.Create(&u)` |
| 2 | 零值被跳过 | `Age: 0` 没写进数据库 | 用 `*int64` / `sql.NullInt64` / `db.Select("Age")` / map |
| 3 | map 创建 | 钩子不执行、关联不保存、主键不回写 | 需要这些行为就改用 struct |
| 4 | 批量插入过大 | MySQL `max_allowed_packet` 报错 | `CreateInBatches(users, 100)` 或 `CreateBatchSize` |
| 5 | 钩子里用外层 `db` | 脱离事务，回滚不一致 | 钩子里一律用参数 `tx` |
| 6 | 钩子返回 error | 插入中止 —— 这是预期行为 | 若想"只警告不中止"，别 return err |
| 7 | 关联为全零值 | 关联表没数据 | 关联字段填非零值，或改用 `FullSaveAssociations` |
| 8 | 插入了不想写的字段 | 比如用户能自设 `Role` | `db.Omit("Role").Create(&u)` |
| 9 | 虚拟列被写入 | `INSERT` 报错 | 加字段权限 `gorm:"->"` 或 `default:(-)` |
| 10 | `OnConflict` 不生效 | 冲突没被识别 | 确认 `Columns` 上存在主键/唯一索引 |
| 11 | 默认事务开销 | 单条插入也有 BEGIN/COMMIT | `SkipDefaultTransaction: true`（慎用） |
| 12 | 唯一键冲突报错难判断 | 拿不到结构化错误 | `gorm.Config{TranslateError: true}` + `errors.Is(err, gorm.ErrDuplicatedKey)` |
| 13 | 记错钩子顺序 | 以为后钩子是 `AfterSave → AfterCreate` | 实际是 `AfterCreate → AfterSave`，前后不对称 |
| 14 | `ctx` 传 `nil` | panic | 用 `context.Background()` 或 `context.TODO()` |
| 15 | `defer cancel()` 忘了写 | 定时器泄漏，`go vet` 报 `lostcancel` | `ctx, cancel := ...; defer cancel()` 紧挨着写 |
| 16 | 传统 API 忘了 `WithContext` | 超时/取消完全失效 | `db.WithContext(ctx)...` |
| 17 | `result.Result` 未判空 | PG/SQLite 上 panic | `if result.Result != nil { ... }` |
| 18 | 复用同一个 `WithResult()` | 并发下数据竞争 | 每次操作现场 `gorm.WithResult()` |

第 12 条的用法：

```go
db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})

err := db.Create(&u).Error
if errors.Is(err, gorm.ErrDuplicatedKey) {
	// 唯一键冲突，可以友好提示用户
}
```

---

## 10. 动手练习

基于你现有的 `GORM/CRUD/main.go`，建议按这个顺序练一遍，每一步都用 `logger.Info` 看 SQL 变化：

```go
package main

import (
	"errors"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

type User struct {
	ID     int64  `gorm:"primaryKey"`
	Name   string `gorm:"size:64;not null;uniqueIndex"`
	Age    int    `gorm:"default:18"`
	Email  string `gorm:"size:128"`
	Active *bool  `gorm:"default:true"`
}

func main() {
	dsn := "root:bjy12345@(127.0.0.1:3306)/sql_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Info),
		TranslateError: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal(err)
	}

	// ① 单条创建 + 主键回写
	u := User{Name: "bai", Age: 19}
	if err := db.Create(&u).Error; err != nil {
		log.Fatal(err)
	}
	fmt.Println("插入后 ID =", u.ID)

	// ② 零值问题：Age 传 0 会被跳过吗？换成指针试试
	zero := 0
	t := true
	db.Create(&User{Name: "zero-age", Age: zero, Active: &t})
	// 观察 SQL：Active 被写入，Age=0 是否出现？

	// ③ Select / Omit
	db.Select("Name", "Age").Create(&User{Name: "s1", Age: 20, Email: "a@b.c"})
	db.Omit("Email").Create(&User{Name: "s2", Age: 21, Email: "x@y.z"})

	// ④ 批量插入 + 分批
	users := []User{{Name: "b1"}, {Name: "b2"}, {Name: "b3"}}
	db.CreateInBatches(&users, 2) // 看日志里是不是 2 条 + 1 条

	// ⑤ Upsert
	db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoUpdates: clause.AssignmentColumns([]string{"age"}),
	}).Create(&User{Name: "bai", Age: 99})

	// ⑥ 唯一键冲突 → 结构化错误
	err = db.Create(&User{Name: "bai"}).Error
	fmt.Println("是否重复键错误:", errors.Is(err, gorm.ErrDuplicatedKey))
}
```

**练习任务清单：**

- [ ] 给 `User` 实现 `BeforeCreate` / `AfterCreate`，打印日志观察调用顺序
- [ ] 在 `BeforeCreate` 里给 `Email` 为空时补一个默认邮箱
- [ ] 试一次 `db.Session(&gorm.Session{SkipHooks: true}).Create(&u)`，确认钩子没被调用
- [ ] 用 `db.Model(&User{}).Create(map[string]any{...})` 插入 `Age: 0`，对比 struct 的行为
- [ ] 加一张 `CreditCard` 表，用关联创建一次性插入两张表的数据
- [ ] 故意让 `BeforeCreate` 返回 error，观察是否整条插入被中止

---

### 练习 B：验证钩子顺序 + 用一遍 Generics API

这段代码一次把三件事都验证了：**钩子真实顺序**、`ctx` 超时、`WithResult()`。

```go
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type User struct {
	ID       int64 `gorm:"primaryKey"`
	Name     string
	Email    string `gorm:"uniqueIndex"`
	Password string
}

// ---- 四个钩子，只打印，观察真实调用顺序 ----
func (u *User) BeforeSave(tx *gorm.DB) error   { fmt.Println("1. BeforeSave"); return nil }
func (u *User) BeforeCreate(tx *gorm.DB) error { fmt.Println("2. BeforeCreate"); return nil }
func (u *User) AfterCreate(tx *gorm.DB) error  { fmt.Println("3. AfterCreate"); return nil }
func (u *User) AfterSave(tx *gorm.DB) error    { fmt.Println("4. AfterSave"); return nil }

func main() {
	dsn := "root:bjy12345@(127.0.0.1:3306)/sql_test?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal(err)
	}

	// ============ ① 传统 API：先跑一遍，确认钩子顺序 ============
	fmt.Println("--- Traditional API ---")
	u1 := User{Name: "bai", Email: "BAI@Example.COM "}
	if err := db.Create(&u1).Error; err != nil {
		log.Fatal(err)
	}
	fmt.Println("主键回写:", u1.ID)
	// 预期打印顺序：1 BeforeSave → 2 BeforeCreate → 3 AfterCreate → 4 AfterSave
	// 如果打出来是 AfterSave 在 AfterCreate 前面，说明你看的教程是错的

	// ============ ② Generics API + ctx 超时 + WithResult ============
	fmt.Println("--- Generics API ---")

	// 给 ctx 一个 3 秒超时。生产环境应该用 r.Context() 起头
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel() // ← 千万别忘，否则 go vet 会报 lostcancel

	result := gorm.WithResult() // 每次操作都新建一个
	u2 := User{Name: "jinzhu", Email: "Jinzhu@Example.com"}
	err = gorm.G[User](db, result).Create(ctx, &u2)

	if err != nil {
		// 区分超时 / 主动取消 / 业务错误
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			fmt.Println("SQL 超时了，去查慢查询日志")
		case errors.Is(err, context.Canceled):
			fmt.Println("调用方主动取消")
		default:
			fmt.Println("其他错误:", err)
		}
		return
	}

	fmt.Println("RowsAffected =", result.RowsAffected) // 1
	fmt.Println("u2.ID        =", u2.ID)               // 主键照样回写
	if result.Result != nil {                          // MySQL 有值，PG/SQLite 可能为 nil
		lastID, _ := result.Result.LastInsertId()
		fmt.Println("LastInsertId =", lastID)          // 与 u2.ID 一致
	}

	// ============ ③ 传统 API 的等价写法（对比记忆） ============
	r := db.WithContext(ctx).Create(&User{Name: "third", Email: "third@example.com"})
	fmt.Println("传统 API:", r.Error, r.RowsAffected)

	// ============ ④ 观察 ctx 被取消后的行为 ============
	canceledCtx, cancel2 := context.WithCancel(context.Background())
	cancel2() // 立刻取消
	err = gorm.G[User](db).Create(canceledCtx, &User{Name: "never", Email: "never@example.com"})
	fmt.Println("在已取消的 ctx 上插入:", errors.Is(err, context.Canceled)) // true

	// ============ ⑤ 钩子校验失败会中止整个插入 ============
	// 在 BeforeCreate 里加上：
	//     if u.Email == "" { return errors.New("邮箱不能为空") }
	// 再跑一次 Create，观察错误被返回、且数据库里没有新记录
}
```

**练习任务清单（续）：**

- [ ] 跑一遍上面的代码，确认四个钩子的打印顺序是 1→2→3→4（`AfterCreate` 在 `AfterSave` 前）
- [ ] 把 `ctx` 的超时改成 `1 * time.Nanosecond`，观察报错是不是 `DeadlineExceeded`
- [ ] 把 `defer cancel()` 注释掉，跑 `go vet ./...` 看它报什么错
- [ ] 在 `BeforeCreate` 里加 `if u.Email == "" { return errors.New("邮箱不能为空") }`，确认插入被中止
- [ ] 把 `result := gorm.WithResult()` 挪到 `for` 循环外面复用，配合并发 `go` 关键字跑一下，用 `go run -race` 观察数据竞争
- [ ] 给 `User` 加一个 `AfterCreate` 往 `AuditLog` 表写一条记录，故意让主逻辑在第 2 条语句失败，确认审计日志也一起回滚

---

## 附：一页速记

```
Create 三板斧：
  传指针     →  才能回写主键
  看 RowsAffected / Error  →  判断结果
  开 logger.Info  →  看到生成的 SQL

钩子顺序（⚠️ 后钩子顺序是反的）：
  BeforeSave → BeforeCreate → INSERT → AfterCreate → AfterSave

零值三选一：
  指针类型  |  sql.NullXxx  |  db.Select("Field")

批量两参数：
  CreateInBatches(users, n)  |  gorm.Config{CreateBatchSize: n}

Upsert 四模式：
  DoNothing  |  Assignments(map)  |  AssignmentColumns(cols)  |  UpdateAll

Generics API 两件事：
  ctx      → 等价于传统 API 的 db.WithContext(ctx)，别传 nil、别忘 defer cancel()
  result   → gorm.WithResult() 拿行数/错误，类型未导出，只能这么构造
```

---

*原文版权归 GORM 官方文档所有，本文档为学习整理与扩展说明。*
