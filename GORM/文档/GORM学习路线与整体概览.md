# GORM 学习路线图与整体概览

> 官方文档入口：<https://gorm.io/zh_CN/docs/index.html>
> 本地环境：`gorm.io/gorm v1.31.2` + `gorm.io/driver/mysql v1.6.0`（Go 1.26）
> 文档章节清单核对自官方文档仓库 `go-gorm/gorm.io` 的 `pages/docs` 目录，共 45 个页面。
> 配套阅读：《GORM 创建 Create 详解》（同目录）

---

## 目录

- [第一部分：GORM 整体概览](#第一部分gorm-整体概览)
  - [1.1 一句话定位](#11-一句话定位)
  - [1.2 官方特性清单的六类归纳](#12-官方特性清单的六类归纳)
  - [1.3 GORM 与同类方案对比](#13-gorm-与同类方案对比)
  - [1.4 四个设计取向](#14-四个设计取向)
  - [1.5 版本与生态](#15-版本与生态)
- [第二部分：GORM 的心智模型](#第二部分gorm-的心智模型)
  - [2.1 四层结构](#21-四层结构)
  - [2.2 为什么这个模型能解释 80% 的"奇怪行为"](#22-为什么这个模型能解释-80-的奇怪行为)
- [第三部分：官方文档知识地图](#第三部分官方文档知识地图)
- [第四部分：学习思路（五条原则）](#第四部分学习思路五条原则)
- [第五部分：学习步骤（六个阶段）](#第五部分学习步骤六个阶段)
- [第六部分：常见学习误区](#第六部分常见学习误区)
- [第七部分：学习路线映射到 GoFilm](#第七部分学习路线映射到-gofilm)
- [附录：核心 API 速查地图](#附录核心-api-速查地图)

---

## 第一部分：GORM 整体概览

### 1.1 一句话定位

**GORM 是 Go 生态里最流行的全功能 ORM（Object Relational Mapping，对象关系映射）库**，作者 Jinzhu，官网 <https://gorm.io>。

它要做的事只有一件：**让你用 Go 结构体和链式方法操作数据库，而不是手写 SQL 和 `rows.Scan()`。**

```go
// 原生 database/sql：手写 SQL + 手动映射
rows, _ := db.Query("SELECT id, name, age FROM users WHERE age > ?", 18)
for rows.Next() {
    var u User
    rows.Scan(&u.ID, &u.Name, &u.Age) // 字段一多就全是重复劳动
}

// GORM：声明结构体，剩下的交给它
var users []User
db.Where("age > ?", 18).Find(&users)
```

注意它**没有取代 SQL，而是取代了重复的 SQL 拼装和结果映射**。底层依然是 `database/sql` + 官方驱动。

### 1.2 官方特性清单的六类归纳

官方首页列了 18 条特性。直接背没意义，按能力归成六类更好记：

| 类别 | 官方特性（对应） | 一句话说明 |
|---|---|---|
| **① 模型与映射** | 全功能 ORM、复合主键、索引、约束、Data Types | 结构体 ↔ 表的双向映射，含自定义类型 |
| **② CRUD** | Create/Save/Update/Delete/Find、批量插入、FindInBatches、Map 创建、SQL 表达式 | 日常 90% 的代码都在这 |
| **③ 关联** | Has One / Has Many / Belongs To / Many To Many / 多态 / 单表继承、Preload、Joins | GORM 最出彩也最容易踩坑的部分 |
| **④ 事务与并发** | 事务、嵌套事务、SavePoint、Rollback To、数据库锁、Optimizer/Index/Comment Hint | 保证数据一致性 |
| **⑤ 可扩展性** | 钩子（Hook）、Scopes、插件 API、数据库解析器（读写分离）、Prometheus、Sharding、Serializer | 把 GORM 改造成你需要的样子 |
| **⑥ 工程化** | Auto Migration、自定义 Logger、Context、预编译模式、DryRun、泛型 API | 上线必须关心的部分 |

### 1.3 GORM 与同类方案对比

| 方案 | 抽象层级 | 优点 | 代价 | 适用场景 |
|---|---|---|---|---|
| **`database/sql`** | 无 | 完全可控、零额外依赖 | 手写 SQL + 手动 Scan，重复劳动多 | 极简项目、极致性能 |
| **`sqlx`** | 轻 | 保留 SQL、自动 Scan 到结构体 | 不生成 SQL、无关联、无迁移 | 喜欢写 SQL 的团队 |
| **`GORM`** | 重 | 自动 SQL、关联、钩子、迁移、插件 | 有"魔法"、性能有开销、出错难排查 | 业务系统、快速开发 |
| **`ent`** | 重 | Schema as Code、类型安全、图查询强 | 学习曲线陡、代码生成流程重 | 复杂图关系、Facebook 系 |
| **`sqlc`** | 中 | 从 SQL 生成类型安全代码、零运行时反射 | 需要写 SQL、动态查询弱 | 强类型爱好者 |

**结论：** 如果你要**快速搭建一个业务系统**（比如 GoFilm），GORM 是性价比最高的选择。它的"魔法"来自反射和回调链 —— 这正是第二部分要讲清楚的。

### 1.4 四个设计取向

理解这四点，你就理解了 GORM 的"性格"：

1. **约定优于配置（Conventions over Configuration）**
   不加任何 tag，GORM 就能工作：`User{}` → 表 `users`、字段 `Name` → 列 `name`。想改再用 tag 覆盖。详见 [约定](https://gorm.io/zh_CN/docs/conventions.html)。

2. **链式调用 + 会话克隆（Method Chaining + Session）**
   `db.Where(...).Order(...).Limit(10).Find(&users)` —— 注意**每个方法都返回新的 `*gorm.DB`**，原 `db` 不被污染。这个"克隆"机制是 GORM 最核心的设计之一，也是"为什么我的条件凭空出现了"这类 bug 的根源。

3. **回调驱动（Callback-driven）**
   CRUD 不是写死的流程，而是一条**可插拔的回调链**。所以钩子（Hook）、Scopes、插件才得以存在。

4. **默认安全、可关可换**
   写操作默认开事务、参数一律预编译占位符、软删除默认开启。性能敏感时可以逐一关掉（`SkipDefaultTransaction`、`Unscoped` 等）。**先知道默认值是什么，再决定要不要改。**

### 1.5 版本与生态

| 项目 | 说明 |
|---|---|
| 主线 | `gorm.io/gorm` —— **v2 之后的唯一主线**，全功能支持 |
| 已废弃 | `github.com/jinzhu/gorm` —— 旧的 v1，**不要再用**，网上大量老教程仍写这个 import |
| 驱动 | `gorm.io/driver/mysql` / `sqlite` / `postgres` / `sqlserver` / `clickhouse` / `tidb` 等 |
| 官方插件 | `gorm.io/plugin/dbresolver`（读写分离）、`prometheus`、`sharding`（分表）、`serializer` |
| 辅助工具 | `gorm.io/gen`（类型安全代码生成）、`gorm.io/hints`（索引提示） |

> ⚠️ **第一个坑**：搜到的中文教程有相当比例是 v1 时代的写法（`db.Table("users").Where(...)` 里 `Where` 参数风格、`db.LogMode(true)` 等）。**v2 的 `LogMode` 已改为 `Logger: logger.Default.LogMode(logger.Info)`**。判断方法：看 import 路径里有没有 `.io`。

---

## 第二部分：GORM 的心智模型

### 2.1 四层结构

不要一开始就想"GORM 有哪些 API"（有几百个），而要先建立这个内部结构：

```
┌──────────────────────────────────────────────────────────────┐
│  调用层    db.Create(&u) · db.Where(...).Find(&users)         │  ← 你写的代码
└───────────────────────────────┬──────────────────────────────┘
                                ▼
┌──────────────────────────────────────────────────────────────┐
│  会话层    *gorm.DB 克隆体                                    │  ← 携带 Config / 事务
│            Config · Statement · Context · 事务句柄            │     / Context
└───────────────────────────────┬──────────────────────────────┘
                                ▼
┌──────────────────────────────────────────────────────────────┐
│  语句层    Statement + Clause                                 │  ← 把"要做什么"
│            积木式拼出 SQL 片段，再由 Dialector 方言化           │     翻译成 SQL
└───────────────────────────────┬──────────────────────────────┘
                                ▼
┌──────────────────────────────────────────────────────────────┐
│  回调层    四条可插拔回调链                                     │  ← 钩子/Scopes/
│            create / query / update / delete                   │     插件挂载点
└───────────────────────────────┬──────────────────────────────┘
                                ▼
┌──────────────────────────────────────────────────────────────┐
│  驱动层    database/sql + Dialector                           │  ← 真正的连接与网络
│            MySQL / PostgreSQL / SQLite / SQL Server           │
└──────────────────────────────────────────────────────────────┘
```

| 层 | 归谁管 | 你要知道的 |
|---|---|---|
| **调用层** | 你的代码 | `db.Create(&u)` 只是入口，真正的活在下面三层 |
| **会话层** | `*gorm.DB` | 它**不是连接**，而是"一条会话的配置 + 上下文 + 事务句柄"。`db.Where(...)` 返回的是克隆体 |
| **语句层** | `Statement` + `Clause` | 把"你要做什么"翻译成 SQL 片段的积木盒。`db.Statement.Raw`、`db.Statement.SQL` 可以打印出来调试 |
| **回调层** | `callbacks/*.go` | 四条链：`create` / `query` / `update` / `delete`，每条链由若干命名回调组成，可替换、可插入 |
| **驱动层** | `Dialector` + `database/sql` | 负责占位符差异（`?` vs `$1`）、`RETURNING` 支持、`LastInsertId` 语义等 |

**关键区分：`*gorm.DB` 不是数据库连接。** 真正的连接池藏在 `sql.DB` 里：

```go
sqlDB, _ := db.DB()          // 从 gorm.DB 取出底层 *sql.DB
sqlDB.SetMaxOpenConns(100)   // 连接池配置在这里
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(time.Hour)
```

这一点不搞清，后面遇到"并发卡死""连接耗尽"会完全摸不着头脑。

### 2.2 为什么这个模型能解释 80% 的"奇怪行为"

举几个典型的"GORM 玄学"，用上面的模型全部能解释：

| 现象 | 根因（在哪一层） |
|---|---|
| `db.Where("a = ?", 1)` 之后 `db.Find()` 居然带了条件 | 会话层 —— 变量复用；`Where` 返回克隆体，但如果你写 `db = db.Where(...)` 就污染了 `db` |
| 钩子里查不到刚插入的数据 | 回调层 + 事务层 —— 钩子跑在事务内，且 `BeforeCreate` 时 INSERT 还没执行 |
| 批量插入拿不到全部主键 | 驱动层 —— MySQL 没有 `RETURNING`，只能用 `LAST_INSERT_ID()` 推导 |
| `Age: 0` 没被写进库 | 语句层 —— 反射判定为零值 → 该字段被从 SQL 里剔除 |
| `Preload` 和 `Joins` 结果不一样 | 关联层 —— `Preload` 发第二条查询，`Joins` 拼成一条 JOIN |
| 加了 `db.Debug()` 但 SQL 只在控制台闪一眼 | 会话层 —— 应改在 `gorm.Open` 时配置 `Logger` |

> **学习方法论**：遇到任何 GORM 的迷惑行为，先问三个问题：
> ① 现在在哪个回调阶段？② 生成的实际 SQL 长什么样？③ 当前会话的配置是什么？
> 把 `Logger` 开到 `Info`，90% 的问题会自己现形。

---

## 第三部分：官方文档知识地图

官方侧边栏的顺序**基本就是推荐阅读顺序**。下表的"重要度"是我的主观标注，帮你决定哪些要精读、哪些用到再查。

### 模块 A：开始（打地基，必读）

| 页面 | 解决什么问题 | 重要度 |
|---|---|---|
| [概览](https://gorm.io/zh_CN/docs/index.html) | 特性、安装、快速入门 | ★★★ |
| [连接数据库](https://gorm.io/zh_CN/docs/connecting_to_the_database.html) | DSN、四种驱动、连接池 | ★★★ |
| [约定](https://gorm.io/zh_CN/docs/conventions.html) | 表名/列名/时间戳/软删除的默认规则 | ★★★ |
| [GORM 配置](https://gorm.io/zh_CN/docs/gorm_config.html) | `gorm.Config{}` 里每个开关的含义 | ★★★ |
| [方法链](https://gorm.io/zh_CN/docs/method_chaining.html) | 链式调用、`Session`、`WithContext` | ★★★ |
| [会话](https://gorm.io/zh_CN/docs/session.html) | `Session{}` 的完整配置项 | ★★☆ |
| [上下文](https://gorm.io/zh_CN/docs/context.html) | `WithContext`、超时/取消 | ★★☆ |
| [事务](https://gorm.io/zh_CN/docs/transactions.html) | 手动/嵌套/保存点 | ★★★ |
| [错误处理](https://gorm.io/zh_CN/docs/error_handling.html) | `ErrRecordNotFound`、`TranslateError` | ★★★ |
| [泛型接口](https://gorm.io/zh_CN/docs/generic_interface.html) | `gorm.G[T]` 的使用 | ★★☆ |

### 模块 B：CRUD 教程（核心，必须逐个动手）

| 页面 | 解决什么问题 | 重要度 |
|---|---|---|
| [模型定义](https://gorm.io/zh_CN/docs/models.html) | tag 全表、`gorm.Model`、字段权限、命名策略 | ★★★ |
| [创建](https://gorm.io/zh_CN/docs/create.html) | 见配套文档 | ★★★ |
| [查询](https://gorm.io/zh_CN/docs/query.html) | `Where`/`First`/`Find`/`Order`/`Limit` | ★★★ |
| [高级查询](https://gorm.io/zh_CN/docs/advanced_query.html) | 智能选择字段、锁、子查询、`FirstOrInit` | ★★☆ |
| [更新](https://gorm.io/zh_CN/docs/update.html) | `Save`/`Update`/`Updates` 与零值陷阱 | ★★★ |
| [删除](https://gorm.io/zh_CN/docs/delete.html) | 软删除、`Unscoped`、批量删除 | ★★★ |
| [SQL 构建器](https://gorm.io/zh_CN/docs/sql_builder.html) | `Raw`/`Exec`/`ToSQL`/`Row` | ★★★ |
| [索引](https://gorm.io/zh_CN/docs/indexes.html) | `index`/`uniqueIndex`/复合索引 | ★★☆ |
| [约束](https://gorm.io/zh_CN/docs/constraints.html) | 外键、`CHECK`、`OnUpdate`/`OnDelete` | ★★☆ |
| [复合主键](https://gorm.io/zh_CN/docs/composite_primary_key.html) | 联合主键的写法 | ★☆☆ |
| [Data Types](https://gorm.io/zh_CN/docs/data_types.html) | 自定义类型、`Scanner`/`Valuer` | ★★☆ |
| [安全](https://gorm.io/zh_CN/docs/security.html) | SQL 注入防护、`?` 占位符 | ★★☆ |
| [钩子](https://gorm.io/zh_CN/docs/hooks.html) | 七个钩子接口与完整生命周期 | ★★★ |
| [Scopes](https://gorm.io/zh_CN/docs/scopes.html) | 把复用条件封装成函数 | ★★☆ |
| [迁移](https://gorm.io/zh_CN/docs/migration.html) | `AutoMigrate`、`Migrator`、生产慎用 | ★★★ |
| [Hints](https://gorm.io/zh_CN/docs/hints.html) | 索引提示、优化器提示 | ★☆☆ |
| [写驱动](https://gorm.io/zh_CN/docs/write_driver.html) / [写插件](https://gorm.io/zh_CN/docs/write_plugins.html) | 扩展 GORM 本身 | ★☆☆ |

### 模块 C：关联（GORM 的分水岭）

| 页面 | 解决什么问题 | 重要度 |
|---|---|---|
| [Belongs To](https://gorm.io/zh_CN/docs/belongs_to.html) | 属于（多对一） | ★★★ |
| [Has One](https://gorm.io/zh_CN/docs/has_one.html) | 一对一 | ★★★ |
| [Has Many](https://gorm.io/zh_CN/docs/has_many.html) | 一对多 | ★★★ |
| [Many To Many](https://gorm.io/zh_CN/docs/many_to_many.html) | 多对多 + 中间表 | ★★★ |
| [多态关联](https://gorm.io/zh_CN/docs/polymorphism.html) | 一个表关联多种父表 | ★★☆ |
| [关联模式](https://gorm.io/zh_CN/docs/associations.html) | `Association()` 增删改查关联、`Select`/`Omit` | ★★★ |
| [预加载](https://gorm.io/zh_CN/docs/preload.html) | `Preload` vs `Joins`，避免 N+1 | ★★★ |

> **关于 Joins**：官方把 `Joins` 预加载的内容并入了[预加载](https://gorm.io/zh_CN/docs/preload.html) 页面，没有独立的 Joins 页。

### 模块 D：高级主题与工程化

| 页面 | 解决什么问题 | 重要度 |
|---|---|---|
| [Logger](https://gorm.io/zh_CN/docs/logger.html) | 日志级别、慢 SQL 阈值、自定义日志 | ★★★ |
| [性能](https://gorm.io/zh_CN/docs/performance.html) | 预编译、连接池、关闭默认事务等 | ★★☆ |
| [数据库解析器](https://gorm.io/zh_CN/docs/dbresolver.html) | 读写分离、多库 | ★★☆ |
| [Sharding](https://gorm.io/zh_CN/docs/sharding.html) | 分表 | ★☆☆ |
| [Serializer](https://gorm.io/zh_CN/docs/serializer.html) | 字段级序列化（存 JSON 等） | ★★☆ |
| [Prometheus](https://gorm.io/zh_CN/docs/prometheus.html) | 指标采集 | ★☆☆ |
| [Changelog](https://gorm.io/zh_CN/docs/changelog.html) / [v2 版本说明](https://gorm.io/zh_CN/docs/v2_release_note.html) | v1→v2 迁移差异 | ★☆☆ |

---

## 第四部分：学习思路（五条原则）

### 原则一：把 Logger 打开，用 SQL 学习 GORM

这是**最高优先级的一条**。学习阶段不要"猜 GORM 会生成什么 SQL"，直接看：

```go
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})
```

从此每一条操作都会打印真实的 SQL 与耗时。**你对"GORM 在做什么"的直觉，全靠这一步建立。**

### 原则二：先传统 API，后泛型 API

v1.30 引入了泛型 API（`gorm.G[T]`），你的 v1.31.2 两种都支持。但：

- **传统 API**：`db.Create(&u)`、`db.Where(...).Find(&users)` —— 社区资料最多、StackOverflow 答案最多、错误信息最完整。
- **泛型 API**：类型更安全、不用"占位结构体"，但生态和示例还在补齐中。

**建议：前 5 个阶段一律用传统 API。** 等你能看懂 GORM 生成的 SQL、能自己排查问题时，再回头评估泛型 API。

### 原则三：不要背 API，要理解"默认值 vs 显式覆盖"

GORM 的 API 有几百个，背不完。但你可以记住**默认行为**，然后按需覆盖：

| 默认行为 | 想改怎么办 |
|---|---|
| 表名 = 结构体名复数蛇形 | `TableName()` 或 `NamingStrategy` |
| 主键 = `ID` | `gorm:"primaryKey"` |
| 时间戳 = `CreatedAt`/`UpdatedAt` | tag 或 `autoCreateTime:false` |
| 删除 = 软删除（有 `DeletedAt` 时） | `Unscoped()` |
| 写操作 = 开事务 | `SkipDefaultTransaction: true` |
| 零值字段 = 不写入 | 指针 / `Select` / map |
| 查询 = 不加条件全表 | 用 `Where` |

**这张表就是 GORM 的"性格说明书"。**

### 原则四：围绕自己的项目学，不要做玩具 demo

`SELECT * FROM users` 这种例子学不到 GORM 的精髓。真正把你卡住的永远是：

- 复合条件分页查询
- 一对多关联的预加载 + 过滤
- 并发下的乐观锁 / 事务隔离
- 软删除与唯一索引的冲突

**所以：直接拿 GoFilm 当练习场。** 第七部分给了映射表。

### 原则五：卡住时，从"回调链"和"生成的 SQL"两头找答案

- 行为不符合预期 → 先看 SQL（原则一）
- SQL 也不符合预期 → 想它在哪条回调链的第几步
- 还想不通 → 读源码 `callbacks/create.go` / `query.go`，只有几百行，比想象中好读

---

## 第五部分：学习步骤（六个阶段）

### 阶段 1 · 打通 CRUD 最小闭环

**目标：** 让一条数据从 Go 结构体进到 MySQL，再从 MySQL 变成结构体读回来。

**范围：** 概览、连接数据库、约定、模型定义（基础）、创建、查询（基础）、更新（基础）、删除（基础）

**关键 API：**

```go
gorm.Open(mysql.Open(dsn), &gorm.Config{})
db.AutoMigrate(&User{})
db.Create(&u)
db.First(&u, 1)
db.Where("age > ?", 18).Find(&users)
db.Model(&u).Update("Age", 20)
db.Delete(&u, 1)
```

**验收标准：**

- [ ] 能从零建 `go.mod` + 连上 MySQL 并跑通增删改查
- [ ] 打开 Logger 后，能看懂每一条打印出来的 SQL
- [ ] 知道 `db.Create(&u)` 里 `u.ID` 为什么会被填上
- [ ] 知道 `gorm.Model` 里有什么、为什么表会多出 4 个字段
- [ ] 能说清"约定"默认生成了什么表名和列名

> 你现在就停在这里（`CRUD/main.go` 已跑通 Create）。**下一步就该补全 Query/Update/Delete。**

---

### 阶段 2 · 模型与迁移（把表设计对）

**目标：** 用结构体精确表达数据库结构。

**范围：** 模型定义（完整）、Data Types、索引、约束、复合主键、迁移、AutoMigrate 与 Migrator

**关键 API：**

```go
type Movie struct {
    ID        uint      `gorm:"primaryKey"`
    Title     string    `gorm:"size:128;not null;index:idx_title"`
    UUID      string    `gorm:"type:char(36);uniqueIndex"`
    Status    int8      `gorm:"default:1"`
    CreatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

db.AutoMigrate(&Movie{}, &Cinema{})
db.Migrator().HasTable(&Movie{})
db.Migrator().DropColumn(&Movie{}, "old_field")
```

**验收标准：**

- [ ] 能独立写出带索引、唯一约束、默认值、外键的模型
- [ ] 理解 `default` tag 是建表 DDL，不是 Go 侧赋值
- [ ] 理解 `autoCreateTime` / `autoUpdateTime` / `serializer` 三个 tag
- [ ] 知道 `AutoMigrate` 在**生产环境要慎用**（只加不减、不改类型、有丢数据风险）

> 💡 这一阶段对你尤其关键 —— GoFilm 的 `halls`/`seats` 有个 `Version` 字段做乐观锁，正好在这里把 tag 写对。

---

### 阶段 3 · 查询进阶（真正写业务查询）

**目标：** 写出生产级的复杂查询。

**范围：** 查询（完整）、高级查询、SQL 构建器、Scopes

**关键 API：**

```go
// 条件组合
db.Where("status = ?", 1).Or("level > ?", 3).Not("name = ?", "x")

// 分组 + 聚合
db.Model(&Order{}).Select("cinema_id, count(*) as cnt, sum(price) as total").
    Group("cinema_id").Having("cnt > ?", 10).Scan(&results)

// 分页
db.Order("created_at desc").Limit(20).Offset(40).Find(&orders)

// 子查询 / 原生 SQL
db.Where("id IN (?)", db.Table("blacklist").Select("user_id")).Find(&users)
db.Raw("SELECT ... WHERE id = ?", id).Scan(&dest)

// Scopes 复用条件
func Paginate(page, size int) func(*gorm.DB) *gorm.DB {
    return func(d *gorm.DB) *gorm.DB {
        return d.Limit(size).Offset((page - 1) * size)
    }
}
db.Scopes(Paginate(1, 20)).Find(&movies)

// 语句不执行，先看 SQL
stmt := db.Session(&gorm.Session{DryRun: true}).Where(...).Find(&u).Statement
fmt.Println(stmt.SQL.String(), stmt.Vars)
```

**验收标准：**

- [ ] 能写出带条件、排序、分页、聚合的复杂查询
- [ ] 会用 `DryRun` 在真正执行前检查 SQL
- [ ] 会用 `Scopes` 把分页/软删除/权限过滤封装复用
- [ ] 能说清 `First` / `Take` / `Last` / `Find` 的区别（尤其**什么情况下会返回 `ErrRecordNotFound`**）
- [ ] 能判断一段查询是否存在 **N+1** 问题

---

### 阶段 4 · 关联（GORM 的分水岭）

**目标：** 掌握四类关联的读写，理解 `Preload` 与 `Joins` 的取舍。

**范围：** Belongs To、Has One、Has Many、Many To Many、多态关联、关联模式、预加载

**关键 API：**

```go
type Cinema struct {
    ID    string `gorm:"type:char(36);primaryKey"`
    Halls []Hall `gorm:"foreignKey:CinemaID"`   // Has Many
}
type Hall struct {
    ID       string `gorm:"type:char(36);primaryKey"`
    CinemaID string
    Seats    []Seat                             // Has Many
}

// 查询时预加载
db.Preload("Halls.Seats").Find(&cinemas)

// 带条件的预加载
db.Preload("Halls", "status = ?", 1).Find(&cinemas)

// Joins 预加载（一条 SQL，但要注意结果去重）
db.Joins("Halls").Find(&cinemas)

// 关联模式：不用查出来就能操作关联
db.Model(&cinema).Association("Halls").Append(&hall)
db.Model(&cinema).Association("Halls").Replace(&h1, &h2)
db.Model(&cinema).Association("Halls").Count()
```

**验收标准：**

- [ ] 能独立设计并写出四类关联的模型与外键 tag
- [ ] 能说清 `Preload`（两条 SQL、能去重）与 `Joins`（一条 SQL、需手动处理重复）的差异
- [ ] 会用 `Association()` 做关联的增删改查
- [ ] 理解"关联值非零才保存"这条规则（关联创建的前提）
- [ ] 知道 `Omit(clause.Associations)` 怎么一键跳过全部关联

---

### 阶段 5 · 事务、并发与钩子（业务正确性）

**目标：** 让数据在并发和异常下依然正确。

**范围：** 事务、钩子、错误处理、上下文、安全、复合主键

**关键 API：**

```go
// 自动事务：回调返回 error 自动回滚
db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(&order).Error; err != nil {
        return err
    }
    return tx.Model(&Seat{}).Where("id = ?", seatID).Update("status", 1).Error
})

// 手动事务 + 保存点
tx := db.Begin()
tx.SavePoint("sp1")
tx.RollbackTo("sp1")
tx.Commit()

// 悲观锁 / 乐观锁
db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&seat, id)  // FOR UPDATE

// 幂等 + 结构化错误
db.Clauses(clause.OnConflict{DoNothing: true}).Create(&m)
errors.Is(err, gorm.ErrDuplicatedKey)  // 需 TranslateError: true

// 超时控制
ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
defer cancel()
db.WithContext(ctx).Find(&movies)
```

**验收标准：**

- [ ] 能用 `db.Transaction` 把"下单 + 锁座 + 扣库存"写成一个原子操作
- [ ] 理解嵌套事务底层是 **SavePoint**，不是真嵌套
- [ ] 能实现**悲观锁**（`FOR UPDATE`）和**乐观锁**（`Version` 字段 + `WHERE version = ?`）两套方案
- [ ] 能写出 `BeforeCreate` / `AfterCreate` 等钩子，并记住真实顺序：`BeforeSave → BeforeCreate → INSERT → AfterCreate → AfterSave`（⚠️ 后钩子顺序反直觉，很多人记错）
- [ ] 知道钩子里必须用参数 `tx`，不能用外层 `db`
- [ ] 会用 `SkipHooks` / `SkipDefaultTransaction` 会话选项

> 💡 这是 GoFilm 座位并发控制的核心阶段 —— 你 `halls`/`seats` 的 `Version` 乐观锁就在这里落地。

---

### 阶段 6 · 工程化与生产化

**目标：** 从"能跑"到"能在生产跑"。

**范围：** Logger、性能、连接池、dbresolver（读写分离）、Serializer、迁移策略、Sharding

**关键 API：**

```go
// 1. 连接池
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(100)
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(time.Hour)

// 2. 慢 SQL 阈值 + 自定义日志
gorm.Config{
    Logger: logger.New(
        log.New(os.Stdout, "\r\n", log.LstdFlags),
        logger.Config{
            SlowThreshold:             200 * time.Millisecond,
            LogLevel:                  logger.Warn,
            IgnoreRecordNotFoundError: true,
            Colorful:                  true,
        },
    ),
}

// 3. 预编译语句缓存（默认开启，高并发下注意）
PreparedStmt: true

// 4. 读写分离
db.Use(dbresolver.Register(dbresolver.Config{
    Sources:  []gorm.Dialector{mysql.Open(masterDSN)},
    Replicas: []gorm.Dialector{mysql.Open(replicaDSN)},
    Policy:   dbresolver.RandomPolicy{},
}))
```

**验收标准：**

- [ ] 能配置连接池并解释每个参数的作用
- [ ] 能自定义 Logger（慢 SQL 报警、脱敏、IgnoreRecordNotFoundError）
- [ ] 知道生产环境迁移应该用版本化 SQL 脚本，而不是 `AutoMigrate`
- [ ] 知道 `PreparedStmt` 开启后与**事务/PgBouncer 的冲突**风险
- [ ] 会读 `callbacks/*.go` 源码，能自己插入一个自定义回调

---

### 阶段小结：六阶段一览

```
 基础打底 ────────────────────────────────────────────────┐
                                                          │
   [ 阶段 1 ]          [ 阶段 2 ]          [ 阶段 3 ]      │
   打通 CRUD 闭环  →   模型与迁移      →   查询进阶        │
   增删改查 + 看 SQL    tag、约定、        条件、分页、      │
                        AutoMigrate       聚合、Scopes    │
                                                          │
                                                          ▼
 业务进阶 ────────────────────────────────────────────────┐
                                                          │
   [ 阶段 4 ]          [ 阶段 5 ]          [ 阶段 6 ]      │
   关联            →   事务与钩子      →   工程化          │
   四种关系 + 多态      事务、锁、          插件、性能、     │
                        Hook、错误处理      日志、源码      │
                                                          │
                                                          ▼
                                                    可用于生产
```

| 阶段 | 你会获得的能力 | 对应 GoFilm 进度 |
|---|---|---|
| 1 | 能跑通增删改查、看懂 SQL | 已完成（`CRUD/main.go`） |
| 2 | 能把表结构用结构体表达准确 | 待做：`halls`/`seats` 的 tag |
| 3 | 能写生产级查询 | 待做：场次筛选与分页 |
| 4 | 能处理多表关系 | 待做：影院→影厅→座位、影片↔标签 |
| 5 | 能保证并发与异常下的正确性 | 待做：下单锁座 |
| 6 | 能上线 | 待做：连接池、日志、迁移策略 |

---

## 第六部分：常见学习误区

| # | 误区 | 纠正 |
|---|---|---|
| 1 | 「GORM 会自动帮我写 SQL，不用懂 SQL」 | 恰恰相反 —— **GORM 把 SQL 藏起来了，所以你必须更主动地把它挖出来看**。不懂 SQL 用 GORM，等于闭眼开车 |
| 2 | 「先把文档 45 页读完再动手」 | 边读边写。**阶段 1 的四页够你开始**，其余用时再查 |
| 3 | 「`Preload` 和 `Joins` 差不多，随便用」 | `Preload` 发两条 SQL、自动去重，适合列表页；`Joins` 一条 SQL、可加条件、但主表数据会重复，适合带过滤条件的场景 |
| 4 | 「钩子里可以用全局 db 接着操作」 | 必须用 `tx`，否则脱离当前事务，回滚时数据不一致 |
| 5 | 「`AutoMigrate` 万能，直接上生产」 | 它只加不减、不改类型、不删列，还会静默改索引。**生产用版本化迁移脚本** |
| 6 | 「`db` 可以随便复用」 | `db.Where(...)` 返回克隆体，但 `db = db.Where(...)` 会污染变量。**别把 db 当可变全局变量用** |
| 7 | 「指针字段太啰嗦，全用值类型」 | 涉及 `0` / `''` / `false` 有意义的字段必须用指针，否则值会被 GORM 判为零值丢弃 |
| 8 | 「软删除就是删除」 | 有 `DeletedAt` 就默认软删除，**唯一索引会和软删除打架**（重复数据会与已软删记录冲突） |
| 9 | 「v1 教程也能用」 | import 路径带不带 `.io` 是分水岭。`LogMode`、`Table` 参数风格都变了 |
| 10 | 「ORM 慢，所以项目不用」 | 慢的是"用错"。`Preload` 防 N+1、`Select` 只取需要的列、必要时 `Raw` 兜底，GORM 完全能撑住生产 |

---

## 第七部分：学习路线映射到 GoFilm

按阶段推进时，直接用你的项目当练习场。这样学完即用，不会"学完就忘"。

| 阶段 | 在 GoFilm 里练什么 |
|---|---|
| 阶段 1 | 用 `movies` 表跑通完整 CRUD；用 `logger.Info` 看懂 `Create`/`Find` 生成的 SQL |
| 阶段 2 | 把 `halls`（UUID + `Version` 乐观锁）、`seats`（`row`/`col`/`seattype`/`status`）的 tag 写全，加 `uniqueIndex(row,col,hall_id)` |
| 阶段 3 | 实现"按影院 + 日期 + 分类/标签筛选场次"的分页查询；用 `Scopes` 封装分页与软删除过滤 |
| 阶段 4 | `cinemas → halls → seats` 三级 Has Many 预加载；`movies ↔ tags` 用 Many To Many；`reviews` 用 Belongs To |
| 阶段 5 | 用 `db.Transaction` 写"下单 + 锁座 + 生成票务"；用 `Version` 实现乐观锁，或 `FOR UPDATE` 实现悲观锁，对比二者吞吐 |
| 阶段 6 | 配置连接池与慢 SQL 日志；用 `Serializer` 存影片的多图 JSON 字段；把 `dbresolver` 接上 |

**建议的项目目录分层**（避免把 GORM 代码写得到处都是）：

```
GoFilm/
├── dao/          // 只做数据库操作，进出都是 model 或 map
│   ├── movie.go
│   └── hall.go
├── model/        // 纯结构体 + tag，不含业务逻辑
│   ├── movie.go
│   └── hall.go
├── service/      // 业务编排，负责事务边界
├── handler/      // 只处理 HTTP 请求/响应
└── pkg/db/       // gorm.Open、配置、连接池
```

**关键原则：事务边界放在 `service` 层，`dao` 层的函数接收 `tx *gorm.DB` 作为参数。**

```go
// dao 层：不关心事务从哪来
func CreateOrder(tx *gorm.DB, o *model.Order) error {
    return tx.Create(o).Error
}

// service 层：决定事务边界
func (s *OrderService) Book(req BookReq) error {
    return s.db.Transaction(func(tx *gorm.DB) error {
        if err := dao.LockSeat(tx, req.SeatID, req.Version); err != nil {
            return err
        }
        return dao.CreateOrder(tx, buildOrder(req))
    })
}
```

---

## 附录：核心 API 速查地图

### 链式方法（返回 `*gorm.DB`，可继续链）

| 类别 | 方法 |
|---|---|
| 条件 | `Where` `Or` `Not` `Having` |
| 范围 | `Limit` `Offset` `Order` `Group` `Distinct` |
| 关联 | `Preload` `Joins` `Association` |
| 字段 | `Select` `Omit` |
| 锁 | `Clauses(clause.Locking{...})` |
| 会话 | `Session` `WithContext` `Debug` `Scopes` |
| 事务 | `Transaction` `Begin` `Commit` `Rollback` `SavePoint` |

### 终结方法（真正执行 SQL）

| 方法 | 说明 |
|---|---|
| `Create` / `CreateInBatches` | 插入 |
| `Save` | 有主键则更新，否则插入 |
| `First` / `Take` / `Last` / `Find` / `FindInBatches` | 查询 |
| `FirstOrInit` / `FirstOrCreate` | 查找或初始化/创建 |
| `Update` / `Updates` / `UpdateColumn` / `UpdateColumns` | 更新 |
| `Delete` | 删除（默认软删除） |
| `Count` / `Scan` / `Pluck` | 聚合与投影 |
| `Raw` / `Exec` / `Row` / `Rows` | 原生 SQL |
| `Association(...)` | 关联操作 |
| `Migrator()` / `AutoMigrate` | 迁移 |

### 返回值速查

```go
result := db.Create(&u)

result.Error         // error，nil 表示成功
result.RowsAffected  // int64，影响行数
result.Statement     // 完整语句信息（含 SQL 与参数）

// 单独的 RowsAffected 变量
db.Model(&User{}).Where("age > ?", 18).Update("status", 1).RowsAffected

// DB() 拿到底层 *sql.DB，用于连接池
sqlDB, err := db.DB()
```

---

## 一句话总结

> **GORM 的学习不是"记住几百个 API"，而是三件事：**
> **① 建立"反射 + 回调链 + Session 克隆"的心智模型；**
> **② 永远开着 SQL 日志，用真实 SQL 验证你的理解；**
> **③ 按"CRUD → 模型 → 查询 → 关联 → 事务 → 工程化"六阶段推进，每一步都在自己的项目里落地。**

---

*本文档基于 GORM 官方中文文档整理，原文版权归 GORM 项目所有。*
