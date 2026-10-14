# GORM 答疑记录

> **这份文档是什么**
>
> 学习 GORM 时卡住的每个问题，都在这里留一条记录：**问题原话 → 结论 → 依据（源码 / 实测） → 牵连出的知识点**。
> 前面是展开讲解的**专题**，最后是**问答台账**（一句话索引，方便回查）。
> 你后面再问的问题，我会继续往下追加。

**运行环境**（与 `CRUD/go.mod` 一致）

| 项 | 值 |
|---|---|
| GORM | `gorm.io/gorm v1.31.2` |
| MySQL 驱动 | `gorm.io/driver/mysql v1.6.0` |
| 数据库 | MySQL 8.0（`sql_test`） |
| 源码位置 | `%USERPROFILE%\go\pkg\mod\gorm.io\gorm@v1.31.2\` |

**配套文件**

| 文件 | 内容 |
|---|---|
| [`GORM创建Create详解.md`](./GORM创建Create详解.md) | 官方 Create 页面逐节详解（Q1 就出自它的 5.6 节） |
| [`GORM学习路线与整体概览.md`](./GORM学习路线与整体概览.md) | 六阶段学习路线 + 知识地图 |
| [`../事务与会话/main.go`](../事务与会话/main.go) | **本文件所有结论的可运行验证程序** |

---

## 目录

- [Q1 · 「用 `db` 会脱离当前事务，回滚时数据不一致」是什么意思？](#q1--用-db-会脱离当前事务回滚时数据不一致是什么意思)
- [Q2 · 「会话级」是什么意思？](#q2--会话级是什么意思)
- [牵连出的知识点 K1–K7](#牵连出的知识点)
- [问答台账](#问答台账)

---

# Q1 · 「用 `db` 会脱离当前事务，回滚时数据不一致」是什么意思？

**出处**：`GORM创建Create详解.md` → 5.6 使用钩子的注意事项·第 3 条；对应 5.5 实例 3「审计日志」

**原始表述**

> 必须用参数 `tx`，不用外层 `db` —— 用 `db` 会脱离当前事务，回滚时数据不一致

这句话浓缩了太多东西，逐字拆开是三个问题：

1. **什么叫「脱离」？** —— 它跑到哪儿去了？
2. **为什么会「不一致」？** —— 不一致体现在哪两个数上？
3. **怎么证明？** —— 日志能看出来吗？

---

## Q1-1 先把那段代码放回现场

```go
// ✅ 正确：用参数 tx
func (u *User) BeforeCreate(tx *gorm.DB) error {
	return tx.Model(&Counter{}).Where("id = ?", 1).
		UpdateColumn("count", gorm.Expr("count + 1")).Error
}

// ❌ 错误：用外层全局 db
func (u *User) BadBeforeCreate(tx *gorm.DB) error {
	return db.Model(&Counter{}).Where("id = ?", 1).
		UpdateColumn("count", gorm.Expr("count + 1")).Error
}
```

两段代码**只差一个变量名**：`tx` 还是 `db`。生成的 SQL 也**一模一样**：

```sql
UPDATE `counters` SET `count`=count + 1 WHERE id = 1
```

但执行它们的**连接不是同一条**。这就是全部问题的根源。

---

## Q1-2 一句话结论

> **`*gorm.DB` 不是数据库连接，它是「会话句柄」。真正决定"走哪条连接"的，是它内部的 `Statement.ConnPool` 字段。**
>
> 事务做的事就一件：把 `ConnPool` 从「连接池」换成「一条已开启事务的连接」。
>
> 全局 `db` 的 `ConnPool` 始终是连接池 —— 所以钩子里的 `db.Xxx()` 会去池子里**另取一条连接**，
> 这条连接处于**自动提交**状态，语句一执行完就落库了。它和外层事务**毫无关系**。

---

## Q1-3 先分清三层概念（这是理解的基础）

学 GORM 事务时最容易把这三层混在一起：

| 层 | 是什么 | 生命周期 | 谁负责 |
|---|---|---|---|
| **会话（Session）** | 一次链式调用共享的配置 + 条件，载体是 `gorm.DB` 实例 | 一条链，执行完就结束 | GORM |
| **连接（Connection）** | 一条真实的数据库连接（TCP） | 从池里借出到归还 | `database/sql` 连接池 |
| **事务（Transaction）** | 一条连接上的一段 `BEGIN ... COMMIT` 区间 | `Begin` 到 `Commit`/`Rollback` | 数据库 |

**关键点**：**事务绑在「连接」上，不是绑在「会话」上。**

所以一个「会话」到底在不在事务里，唯一判据是看它的 `Statement.ConnPool` 是不是一个事务对象（`*sql.Tx`）。

```
全局 db  ──Statement.ConnPool──▶ *sql.DB    （连接池）
                                              │ 借一条
db.Begin() ──Statement.ConnPool──▶ *sql.Tx  ◀─┘
                                       ↑
                            这条连接被"锁"在事务里
                            池子里再也拿不到它
```

---

## Q1-4 源码链条：四步串起来

### ① 写操作默认会自动开事务

`callbacks/transaction.go`

```go
func BeginTransaction(db *gorm.DB) {
	if !db.Config.SkipDefaultTransaction && db.Error == nil {
		if tx := db.Begin(); tx.Error == nil {
			db.Statement.ConnPool = tx.Statement.ConnPool   // ← 换 ConnPool
			db.InstanceSet("gorm:started_transaction", true)
		}
		// ...
	}
}

func CommitOrRollbackTransaction(db *gorm.DB) {
	if !db.Config.SkipDefaultTransaction {
		if _, ok := db.InstanceGet("gorm:started_transaction"); ok {
			if db.Error != nil {
				db.Rollback()
			} else {
				db.Commit()
			}
			db.Statement.ConnPool = db.ConnPool              // ← 换回连接池
		}
	}
}
```

**「开事务」在代码层面就是给 `Statement.ConnPool` 赋了个新值。** 就这么朴素。

### ② `Commit` / `Rollback` 也是在 `ConnPool` 上做文章

`finisher_api.go`

```go
func (db *DB) Commit() *DB {
	if committer, ok := db.Statement.ConnPool.(TxCommitter); ok && committer != nil && !reflect.ValueOf(committer).IsNil() {
		db.AddError(committer.Commit())
	} else {
		db.AddError(ErrInvalidTransaction)
	}
	return db
}
```

### ③ 钩子拿到的 `tx` 是这么造出来的

`callbacks/callmethod.go` —— **整个文件就 32 行，这是钩子机制的全部**

```go
func callMethod(db *gorm.DB, fc func(value interface{}, tx *gorm.DB) bool) {
	tx := db.Session(&gorm.Session{NewDB: true})          // ← 就是它
	if called := fc(db.Statement.ReflectValue.Interface(), tx); !called {
		switch db.Statement.ReflectValue.Kind() {
		case reflect.Slice, reflect.Array:
			db.Statement.CurDestIndex = 0
			for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
				if value := reflect.Indirect(db.Statement.ReflectValue.Index(i)); value.CanAddr() {
					fc(value.Addr().Interface(), tx)      // ← 批量时，每条元素传的是同一个 tx
				}
				// ...
			}
		}
	}
}
```

### ④ 为什么这个 `tx` 天然还在同一个事务里

`gorm.go` → `Session` + `getInstance`

```go
func (db *DB) Session(config *Session) *DB {
	tx := &DB{Config: &txConfig, Statement: db.Statement, Error: db.Error, clone: 1}
	// ...
	if !config.NewDB {
		tx.clone = 2          // NewDB:true 时保持 1
	}
	return tx
}

func (db *DB) getInstance() *DB {
	if db.clone > 0 {
		tx := &DB{Config: db.Config, Error: db.Error}
		if db.clone == 1 {
			// clone with new statement —— 全新、干净、无条件的 Statement
			tx.Statement = &Statement{
				DB:        tx,
				ConnPool:  db.Statement.ConnPool,   // ★★★ 关键的一行
				Context:   db.Statement.Context,
				Clauses:   map[string]clause.Clause{},
				Vars:      make([]interface{}, 0, 8),
				SkipHooks: db.Statement.SkipHooks,
			}
		} else {
			tx.Statement = db.Statement.clone()      // 带着条件克隆
			tx.Statement.DB = tx
		}
		return tx
	}
	return db
}
```

**`ConnPool: db.Statement.ConnPool` —— 就这一行，让钩子的 `tx` 继承了事务。**

同时注意它是「全新空 Statement」：`Clauses` 是空的。这正对应官方那句话：

> `tx` 是带有 `NewDB` 选项的新会话模式，基于 `tx` 的操作会在同一个事务中，但**不会带上任何当前的条件**。

而**全局 `db` 的 `Statement.ConnPool` 是 `*sql.DB` 连接池**，永远不变。这就是「脱离」的全部机制。

---

## Q1-5 实测证据

我用 `SELECT CONNECTION_ID()` 直接看「这次操作实际走了哪条连接」。
为此在 `CRUD` 的 MySQL 上建了独立的 `demo_` 表，程序跑完自动删掉。完整代码：[`../事务与会话/main.go`](../事务与会话/main.go)

### 场景 1：钩子里用 `tx`

```
-- 事务外的连接 ID = 16
-- 事务内的连接 ID = 16
[rows:1] SELECT CONNECTION_ID()          ← 钩子里查到的也是 16
[rows:1] UPDATE `demo_counters` SET `total`=total + 1 WHERE id = 1
[rows:1] INSERT INTO `demo_users` (`name`) VALUES ('订单A')
-- 事务返回：后续步骤失败，整个事务回滚
-- 钩子里的连接 ID = 16
   ✅ 钩子与事务是同一条连接（16 == 16）→ 同一事务
-- 事务结束后：计数器 = 0，用户数 = 0
   ✅ 用户被回滚了（=0）
   ✅ 计数器被回滚了（=0）
```

### 场景 2：钩子里用外层 `db`（错误写法）

```
-- 事务外的连接 ID = 16
-- 事务内的连接 ID = 16
[rows:1] SELECT CONNECTION_ID()          ← 钩子里查到的变成了 17 ！
[rows:1] UPDATE `demo_counters` SET `total`=total + 1 WHERE id = 1
[rows:1] INSERT INTO `demo_users` (`name`) VALUES ('订单A')
-- 事务返回：后续步骤失败，整个事务回滚
-- 钩子里的连接 ID = 17
   ❌ 钩子用了另一条连接（17 != 16）→ 已经脱离事务
-- 事务结束后：计数器 = 1，用户数 = 0
   ✅ 用户被回滚了（=0）
   ❌ 计数器没回滚（= 1）—— 这就是「数据不一致」
```

### 把两个场景摆在一起看

| | 场景 1（用 `tx`） | 场景 2（用 `db`） |
|---|---|---|
| 钩子用的连接 ID | **16**（= 事务那条） | **17**（另一条） |
| 事务回滚后 · 用户数 | 0 ✅ | 0 ✅ |
| 事务回滚后 · 计数器 | **0** ✅ | **1** ❌ |
| 结果 | 两处修改同生共死 | **用户没了，计数却涨了** |

**「数据不一致」= 本该一起回滚的两条数据，只回滚了一条。**
在你的 GoFilm 里，它的后果可能是：下单失败回滚了订单，但「场次剩余座位数」已经减掉了 —— 这个座位就永远卖不出去了。

---

## Q1-6 为什么这个坑特别难发现

看场景 1 和场景 2 的 SQL 日志 —— **完全一样，字符级一样**：

```sql
UPDATE `demo_counters` SET `total`=total + 1 WHERE id = 1
INSERT INTO `demo_users` (`name`) VALUES ('订单A')
```

更糟的是 **GORM 的日志里根本看不到 `BEGIN` / `COMMIT` / `ROLLBACK`**（我特意 grep 过源码，没有任何 trace 点；只有 `SAVEPOINT` 因为走 `Exec` 才被打印）。

| 你以为的 | 实际上的 |
|---|---|
| 日志里有 `BEGIN` 就是有事务 | 日志里**永远没有** `BEGIN` |
| SQL 一样，行为就一样 | SQL 一样，**连接不一样** |

**所以这类 bug 只能靠「效果」发现**（回滚后数据对不对），或者靠连接 ID / `information_schema.innodb_trx` 这类手段。

> **可迁移的调试技巧**：怀疑事务问题时，在关键位置打一行
> `db.Session(&gorm.Session{}).Raw("SELECT CONNECTION_ID()").Scan(&id)`，
> 比较 ID 是否一致。这是判断「在不在同一个事务里」最快的方法。

---

## Q1-7 怎么修

### 规则：钩子/回调里只用参数 `tx`，绝不用外层 `db`

```go
// ✅ 正确
func (u *User) BeforeCreate(tx *gorm.DB) error {
	return tx.Model(&Counter{}).Where("id = ?", 1).
		UpdateColumn("count", gorm.Expr("count + 1")).Error
}
```

### 那「外层 db」是怎么进来的？——三种常见路径

| 路径 | 例子 | 怎么防 |
|---|---|---|
| **包级全局变量** | `var db *gorm.DB`（你的 `CRUD/main.go` 就是） | 钩子/方法里**禁止**直接引用它 |
| **结构体字段** | `type Svc struct{ db *gorm.DB }`，方法里用 `s.db` | 数据写入统一走传入的 `tx` |
| **闭包捕获** | `db.Transaction(func(tx *gorm.DB) error { ... db.Xxx() ... })` | 闭包里只用 `tx` |

**特别提醒第 3 种**：`db` 和 `tx` 同时在作用域里时，名字很像、IDE 补全也像，最容易手滑。建议：

```go
// 让"用错"变得不可能：把全局 db 藏起来
func (r *ReviewRepo) Create(tx *gorm.DB, rv *Review) error {
	return tx.Create(rv).Error     // 只接收 tx，没有 db 可用
}
```

这也是为什么建议 **`dao` 层函数统一接收 `tx *gorm.DB` 参数**（参见 `GORM学习路线与整体概览.md` 的分层建议）。

### 反过来：什么时候**必须**用全局 `db`？

| 场景 | 用哪个 | 原因 |
|---|---|---|
| 钩子 / 回调 / service 的事务内 | `tx` | 必须同事务 |
| 事务**之外**的独立操作 | `db` | 本来就该独立提交 |
| 想**故意**留痕（审计日志不可回滚） | `db` | ⚠️ 但要写清注释说明这是有意为之 |

最后一条是个真实需求：有些审计日志**就是要**独立提交，不能被业务回滚掉。这时用 `db` 是对的 —— **关键是"你清楚自己在做什么"，而不是"手滑写错了"。**

---

# Q2 · 「会话级」是什么意思？

> 原始问题：`SkipHooks` 是**会话级**的（`GORM创建Create详解.md` 5.7 节），"会话级"是什么意思？

## Q2-1 定义

> **会话（Session）= 一条链式调用上共享的一整套「临时配置 + 运行状态」。**
>
> - **载体**：`gorm.Session`（配置）+ `gorm.Statement`（状态：条件、表名、变量…）
> - **生命周期**：从链式调用派生出的那个 `*gorm.DB` 实例开始，到终结方法执行完为止
> - **「会话级」= 只对这一条链（及其派生链）生效，用完即弃，不影响别处**

**和它相对的是「全局」**：`gorm.Open` 时传的 `gorm.Config`、以及包级 `db` 变量。

## Q2-2 一句话抓住本质

> **GORM 里 `db` 不是连接，是「模板」。每次链式调用都是从这个模板「派生」出一个新会话。**
>
> 「会话级」的意思就是：**这次的改动只写在派生出来的副本上，模板本身不动。**

这也是为什么下面这段代码是安全的：

```go
db.Debug().Create(&u1)   // 只是这一条链开 debug
db.Create(&u2)           // 这条链是普通日志，不受影响
```

## Q2-3 方法分三类（官方分法）

官方 [链式方法](https://gorm.io/zh_CN/docs/method_chaining.html) 页面把方法分成三类，`*gorm.DB` 能不能复用就取决于这个分类：

| 类别 | 作用 | 例子 | 返回值可复用吗 |
|---|---|---|---|
| **Chain Method**（链式方法） | 往当前 `Statement` 追加条件 | `Where` `Select` `Omit` `Joins` `Preload` `Scopes` | ❌ 不可复用（会累积条件） |
| **Finisher Method**（终结方法） | 立即执行 SQL | `Create` `First` `Find` `Save` `Update` `Delete` `Count` | ❌ 不可复用 |
| **New Session Method**（新会话方法） | **造一个可安全复用的新会话** | `Session` `WithContext` `Debug` | ✅ **可复用** |

**记住最后一行 —— 这就是"会话级"的操作入口。**

## Q2-4 一个会话能控制什么

`gorm.go` 里的 `Session` 结构体，**字段清单就是"会话级选项"的全集**：

```go
type Session struct {
	DryRun                   bool            // 只生成 SQL 不执行
	PrepareStmt              bool            // 预编译语句
	NewDB                    bool            // 丢弃之前累积的条件，从干净状态开始
	Initialized              bool            // 执行初始化回调
	SkipHooks                bool            // 跳过钩子
	SkipDefaultTransaction   bool            // 本条链不自动开事务
	DisableNestedTransaction bool            // 禁用嵌套事务
	AllowGlobalUpdate        bool            // 允许无 WHERE 的全表更新/删除
	FullSaveAssociations     bool            // 关联全量保存
	PropagateUnscoped        bool            // 传播 Unscoped
	QueryFields              bool            // 查询时列出字段名
	Context                  context.Context // 超时/取消/链路透传
	Logger                   logger.Interface// 本条链的日志级别
	NowFunc                  func() time.Time// 本条链的时间函数
	CreateBatchSize          int             // 本条链的批量插入大小
}
```

**每一个都是「会话级」**：只影响 `Session(...)` 这一次链式调用，`db` 本身不变。

## Q2-5 三个最常用的会话语义（都实测过）

### ① `SkipHooks`：临时跳过钩子 —— 用完即失效

```go
hookCalls = 0
db.Session(&gorm.Session{SkipHooks: true}).Create(&DemoUser{Name: "跳过钩子"})
// → 钩子调用次数 = 0     ✅ 跳过了

db.Create(&DemoUser{Name: "正常钩子"})
// → 钩子调用次数 = 1     ✅ 自动恢复，db 本身没被改
```

**这就是"会话级"最典型的样子：影响这一次，不影响下一次。**

### ② 链式对象复用 → 条件累积（官方 Bad case）

```go
stmt := db.Where("name = ?", "会话A")     // 注意：没有立刻终结

stmt.Model(&DemoUser{}).Where("name = ?", "会话B").Count(&n1)
// SQL: SELECT count(*) FROM `demo_users` WHERE name = '会话A' AND name = '会话B'
// n1 = 0

stmt.Model(&DemoUser{}).Count(&n2)        // 本以为只剩"会话A"
// SQL: SELECT count(*) FROM `demo_users` WHERE name = '会话A' AND name = '会话B'   ← 污染了！
// n2 = 0  ❌  应该是 1
```

**原因**：`stmt` 是从全局 `db` 派生出来的 `clone=0` 实例，对它调链式方法会**在它自己身上就地修改**，条件就留下了。

### ③ `Session(&gorm.Session{})` → 安全复用（官方 Good case）

```go
base := db.Where("name = ?", "会话A").Session(&gorm.Session{})

base.Model(&DemoUser{}).Where("name = ?", "会话B").Count(&f1)
// WHERE name = '会话A' AND name = '会话B'   → 0（'会话A' 是 base 的固有条件，合理）

base.Model(&DemoUser{}).Count(&f2)   // WHERE name = '会话A'  → 1  ✅
base.Model(&DemoUser{}).Count(&f3)   // WHERE name = '会话A'  → 1  ✅ 反复用都不脏
```

**注意一个易错点**：`Session(&gorm.Session{})` 保证的是「**`base` 不会被后续调用改坏**」，
它**不会**丢掉 `base` 上已经有的条件。想丢条件要用 `NewDB`：

```go
dirty := db.Where("name = ?", "会话A")
dirty.Model(&DemoUser{}).Count(&g1)                                   // WHERE name='会话A' → 1
dirty.Session(&gorm.Session{NewDB: true}).Model(&DemoUser{}).Count(&g2) // 条件清空 → 2
```

**三兄弟对照表：**

| 写法 | 派生的新链 | 原对象 | 已有条件 | 用途 |
|---|---|---|---|---|
| `stmt` 直接复用 | 就是原对象本身 | **被改坏** | 累积下去 ❌ | 别这么干 |
| `.Session(&gorm.Session{})` | 每次都克隆 | 安全 ✅ | **保留** | 做「可复用查询模板」 |
| `.Session(&gorm.Session{NewDB: true})` | 每次全新 | 安全 ✅ | **清空** | 做「干净的新开始」（钩子的 `tx` 就是这个） |

## Q2-6 什么时候该关心"会话级"

| 你想做的事 | 用会话级 |
|---|---|
| 这条链开 debug 日志 | `db.Debug()` 或 `Session{Logger: ...}` |
| 这条链跳过钩子（迁移脚本） | `Session{SkipHooks: true}` |
| 这条链不自动开事务（只读、且确定无并发风险） | `Session{SkipDefaultTransaction: true}` |
| 这个请求带超时/取消 | `WithContext(ctx)` |
| 临时调整批量大小 | `Session{CreateBatchSize: n}` |
| 允许全表更新（`AllowGlobalUpdate`） | `Session{AllowGlobalUpdate: true}` |

**判断口诀：**
> **临时的、这一条链的 → 会话级；长期的、整个程序的 → 全局 `gorm.Config`。**
> 连接串、连接池参数、命名策略、注册的插件 → 全局。
> 开关、条件、日志级别、超时、批次大小 → 会话级。

---

# 牵连出的知识点

> Q1 和 Q2 一路牵连出来的东西。每一条都标了「怎么验证」。

## K1 · GORM 的写操作默认自带事务

**结论**：`Create` / `Update` / `Delete` 默认都会包一层事务，**哪怕只写一行**。

依据：`callbacks/transaction.go` 的 `BeginTransaction` / `CommitOrRollbackTransaction` 挂在默认回调链上。
`gorm.Open` 时传 `SkipDefaultTransaction: true` 可关掉。

**为什么重要**：这是"钩子返回 error 能回滚"的前提。没有这层事务，钩子抛错时前面已写的 INSERT 早就提交了。

⚠️ 注意这是**默认链**的行为，`Session{SkipDefaultTransaction: true}` 只影响你显式指定的那条链。

## K2 · `clone` 机制：`*gorm.DB` 到底能不能复用

**结论**：`DB.clone` 决定 `getInstance()` 的行为，三种取值：

| `clone` | 含义 | 派生出的实例 |
|---|---|---|
| `1` | 全新会话 | **新建空 Statement**（条件清空），继承 `ConnPool` |
| `2` | 带条件克隆 | `Statement.clone()`（条件复制一份，原对象安全） |
| `0` | 普通实例 | `getInstance()` 返回**它自己**，链式方法就地修改 ⚠️ |

来源：`gorm.Open` 返回的 `db` 是 `clone=1`；`db.Where(...)` 之类派生出的实例是 `clone=0`；
`Session(&gorm.Session{})` 产出 `clone=2`；`Session(&gorm.Session{NewDB: true})` 产出 `clone=1`。

**怎么验证**：写 `stmt := db.Where(...)` 然后复用两次，看第二次的 SQL 是不是多了条件（就是 Q2-5 的场景 ②）。

## K3 · 钩子的 `tx` 是怎么造出来的

```go
// callbacks/callmethod.go
tx := db.Session(&gorm.Session{NewDB: true})
```

所以钩子的 `tx` 天然是「**干净（无条件）+ 同事务（继承 ConnPool）**」。官方那句
"`tx` 是带有 `NewDB` 选项的新会话模式，基于 `tx` 的操作会在同一个事务中，但不会带上任何当前的条件"
说的就是这两件事。

**顺带一个反直觉的好消息**：因为 `NewDB: true` 让 `clone=1`，
**在批量插入时，钩子里的 `tx.Model(...).Where(...)` 不会跨元素累积条件**。

实测（场景 5）：批量插入 3 条，每条在 `AfterCreate` 里 `tx.Model(&Counter{}).Where("id = ?", it.ID)`：

```sql
[rows:1] UPDATE `demo_counters` SET `total`=10 WHERE id = 1
[rows:1] UPDATE `demo_counters` SET `total`=20 WHERE id = 2
[rows:1] UPDATE `demo_counters` SET `total`=30 WHERE id = 3
[rows:3] INSERT INTO `demo_items` (`name`,`amount`) VALUES ('a',10),('b',20),('c',30)
```

三条各更新自己的计数器，**没有出现 `WHERE id = 1 AND id = 2` 这种累积**。
如果钩子拿到的是 `clone=0` 的实例，第 2 条就会更新 0 行、静默丢数据 —— GORM 用 `NewDB` 帮你避开了这个坑。

> 但**你自己写的**链式对象没有这层保护，所以 Q2-5 场景 ② 的坑仍然存在。区别在于：
> **框架给你的 `tx` 是安全的，你自己攒的 `stmt` 不是。**

## K4 · 想确认"在不在同一个事务里"，用连接 ID

GORM 日志**不打印** `BEGIN` / `COMMIT` / `ROLLBACK`（源码里没有对应的 trace 点），所以不能靠日志判断。

**三个可靠办法**：

```go
// ① 连接 ID（最快）
func connID(d *gorm.DB) int64 {
	var id int64
	d.Session(&gorm.Session{}).Raw("SELECT CONNECTION_ID()").Scan(&id)
	return id
}
// 同一个事务里，connID 必然相同

// ② 看当前活跃事务（MySQL）
// SELECT trx_id, trx_state, trx_started, trx_mysql_thread_id FROM information_schema.innodb_trx;

// ③ 看连接占用（MySQL）
// SHOW PROCESSLIST;   -- 事务会一直占着一条连接
```

> 注意 ① 里的 `Session(&gorm.Session{})` —— 加它是为了**不让这次查询污染 `d`**
> （`Raw` 会把 SQL 写进 `Statement`）。这本身就是"会话级"的一个实用例子。

## K5 · 嵌套 `Transaction` = `SAVEPOINT`，不是新事务

```go
db.Transaction(func(tx *gorm.DB) error {
	tx.Create(&DemoUser{Name: "外层"})
	inner := tx.Transaction(func(tx2 *gorm.DB) error {   // ← 起点是 tx，不是 db
		return tx2.Create(&DemoUser{Name: "内层"}).Error
	})
	return errors.New("外层失败")                        // 外层失败
})
```

**实测日志**（注意只有 `SAVEPOINT`，没有第二个 `BEGIN`）：

```sql
[rows:1] INSERT INTO `demo_users` (`name`) VALUES ('外层')
[rows:0] SAVEPOINT sp581560879664073303
[rows:1] INSERT INTO `demo_users` (`name`) VALUES ('内层')
-- 内层事务返回：<nil>       ← 看起来"提交成功"了
-- 外层事务返回 error，触发整体回滚
-- 最终用户数 = 0            ← 内层也一起回滚了
```

**结论**：内层返回 `nil` 不代表数据落库 —— 它只是释放了一个 `SAVEPOINT`，
**外层回滚时内层照样没了**。别被 `<nil>` 骗了。

（源码：`finisher_api.go` 的 `Transaction`，检测到 `ConnPool` 已经是 `TxCommitter` 时走 `SavePoint` 分支。）

## K6 · 零值到底什么时候被"跳过"（勘误）

> 这条是**顺手查出来的**，顺带更正了 `GORM创建Create详解.md` 里的三处错误说法。

**准确规则**（全部实测过）：

| 字段情况 | `Create` 传零值的结果 | 实测 SQL |
|---|---|---|
| **无 `default` tag** | **零值照写** | `INSERT INTO demo_zeros (name,count) VALUES ('a',0)` |
| 字面量 `default:18` | 零值**被替换成 18** | `INSERT INTO demo_defaults (name,age,age_ptr) VALUES ('anon',18,18)` |
| `default:uuid_generate_v3()` | 该列**从 INSERT 省略**，交数据库算 | — |
| map 方式创建 | **原样写入，`default` tag 完全不生效** | `INSERT INTO demo_defaults (age,name) VALUES (0,'')` |

**源码判据**（`schema/field.go`）：

```go
skipParseDefaultValue := strings.Contains(field.DefaultValue, "(") && strings.Contains(field.DefaultValue, ")") ||
	strings.ToLower(field.DefaultValue) == "null" || field.DefaultValue == ""
if field.HasDefaultValue && !skipParseDefaultValue {
	field.DefaultValueInterface, err = strconv.ParseInt(field.DefaultValue, 0, 64)
}
```

- 能解析成 Go 字面量 → `DefaultValueInterface != nil` → 走「Go 侧替换零值」
- 含括号（数据库函数）/ `null` / 空 → `nil` → 走「省略该列，交给数据库」

**还有个副作用**（`callbacks/create.go`）：

```go
if values.Values[0][idx], isZero = field.ValueOf(...); isZero {
	if field.DefaultValueInterface != nil {
		values.Values[0][idx] = field.DefaultValueInterface                 // 换值
		stmt.AddError(field.Set(stmt.Context, stmt.ReflectValue,            // 并回写结构体！
			field.DefaultValueInterface))
	}
}
```

实测：

```go
d := DemoDefault{}
db.Create(&d)
fmt.Printf("name=%q age=%d\n", d.Name, d.Age)   // name="anon" age=18
```

**你传进去的结构体被改了。** Create 后立刻用这个对象拼响应或写断言时，要留意。

**一句话记忆（更正版）**：

- **带 `default` tag**：零值会被默认值顶掉 → 想写 `0`/`''`/`false` 必须上指针或 `sql.NullXxx`
- **不带 `default` tag**：零值照写，什么都不用做
- `db.Select("Age")` **救不了**这个场景（实测 Select 强制包含列后，零值**依然**被替换成默认值）

## K7 · 为什么"开日志"看不到事务

GORM 的 logger 只在**执行 SQL 语句**时打点：`callbacks/transaction.go` 里的
`Begin` / `Commit` / `Rollback` 都直接调 `database/sql`，没有 `Trace` 调用。

| 语句 | 日志里有吗 | 原因 |
|---|---|---|
| `BEGIN` / `COMMIT` / `ROLLBACK` | ❌ 没有 | 走 `sql.DB.BeginTx` / `Tx.Commit`，无 trace 点 |
| `SAVEPOINT` / `ROLLBACK TO` | ✅ 有 | 走 `db.Exec(...)`，是正常 SQL 执行路径 |
| `INSERT` / `UPDATE` / `SELECT` | ✅ 有 | 当然有 |

**实际影响**：不能靠日志判断"这条 SQL 到底在不在事务里"。要用 K4 的方法。

---

# 问答台账

> 后续问题都追加在这里，一行一条，方便回查。

| # | 日期 | 问题 | 结论一句话 | 详见 |
|---|---|---|---|---|
| Q1 | 2026-09-19 | 「用 `db` 会脱离当前事务，回滚时数据不一致」是什么意思 | `*gorm.DB` 是会话句柄，事务绑在 `Statement.ConnPool` 上；全局 `db` 用连接池→另取一条连接→自动提交→不跟着回滚 | [Q1](#q1--用-db-会脱离当前事务回滚时数据不一致是什么意思) |
| Q1a | 2026-09-19 | 相关的知识点有哪些 | 见 K1–K7：默认事务 / clone 机制 / 钩子 tx 来源 / 连接 ID 调试 / SAVEPOINT / 零值边界 / 日志无事务 | [K1–K7](#牵连出的知识点) |
| Q2 | 2026-09-19 | 「会话级」是什么意思 | 只对当前这条链式调用生效、用完即弃的配置与状态；载体是 `gorm.Session` + `Statement`；与「全局 `gorm.Config`」相对 | [Q2](#q2--会话级是什么意思) |
| Q3 | 2026-09-19 | `SkipHooks` 为什么是会话级 | 它写在 `Session` 结构体里，只改那一次派生的实例，`db` 本身不变（实测：用完自动恢复） | [Q2-5](#q2-5-三个最常用的会话语义都实测过) |

## 待解决 / 下次想弄明白的

- [ ] `PrepareStmt`（预编译）在事务里为什么会报错？什么场景该开
- [ ] `AllowGlobalUpdate` 之外，还有哪些"防呆"开关（防全表更新、防误删）
- [ ] `Context` 超时在事务中的表现：`DefaultTransactionTimeout` 是怎么用的
- [ ] 读写分离（`dbresolver`）下，事务会不会走从库
- [ ] 连接池参数（`SetMaxOpenConns` / `SetConnMaxLifetime`）怎么和事务、长事务配合

---

## 附：怎么自己重跑这些结论

```bash
cd D:/code/go/repository/learn-Go/GORM/事务与会话
go run .
```

程序会自动建 `demo_*` 表、跑完 8 个场景、然后删表，不会碰你的业务表。
输出里每个场景都有 `✅` / `❌` 标记和真实 SQL 日志。

想进数据库亲自看看，把 `main.go` 顶部的 `keepTables` 改成 `true`。
