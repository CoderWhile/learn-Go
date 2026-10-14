# GORM 查询 Query 详解

> **官方原文**：https://gorm.io/zh_CN/docs/query.html
> **适配版本**：`gorm.io/gorm v1.31.2` + `gorm.io/driver/mysql v1.6.0`（你本地 `CRUD/go.mod` 的真实版本）
> **本文档做了什么**：在官方原文基础上，补了 ① 源码级的内部机制（`ErrRecordNotFound` 怎么产生、主键条件怎么构建）；② 16 个可运行场景的真实 SQL 实测（源码在 `GORM/查询Query/main.go`，输出在 `运行输出.txt`）；③ 官方没写的坑点。

---

## 目录

- [0. 先建立心智模型：查询和「写」的本质区别](#0-先建立心智模型查询和写的本质区别)
- [1. 检索单个对象：First / Take / Last](#1-检索单个对象first--take--last)
- [2. 根据主键检索](#2-根据主键检索)
- [3. 检索全部对象：Find](#3-检索全部对象find)
- [4. 条件：Where / Not / Or](#4-条件where--not--or)
- [5. 选择特定字段：Select](#5-选择特定字段select)
- [6. 排序 / 分页：Order / Limit / Offset](#6-排序--分页order--limit--offset)
- [7. 分组聚合：Group By / Having / Distinct](#7-分组聚合group-by--having--distinct)
- [8. Joins](#8-joins)
- [9. Scan](#9-scan)
- [10. 坑点速查表](#10-坑点速查表)
- [11. 动手练习](#11-动手练习)

---

## 0. 先建立心智模型：查询和「写」的本质区别

在上一份 Create 文档里你知道了：**写操作（Create/Update/Delete）默认会包一层事务**（回调链里有 `begin_transaction` → `commit_or_rollback_transaction`）。

**查询完全没有这一层。** 看源码 `callbacks/callbacks.go` 第 51–54 行，查询回调链只有三步：

```
gorm:query        → Query        （BuildQuerySQL → 执行 SQL → Scan 扫描）
gorm:preload      → Preload      （Preload 预加载）
gorm:after_query  → AfterQuery   （触发 AfterFind 钩子）
```

**没有 `begin_transaction`，也没有 `commit_or_rollback_transaction`。** 读操作是"一锤子买卖"，一条 SQL 发出去、结果扫回来就结束了。

这个区别带来两个直接推论：

1. **查询是并发安全的**——多个 goroutine 共享同一个 `*gorm.DB` 并发 `Find` 完全没问题（但同一个"链式对象"不能复用，见第 10 节坑 6）。
2. **查询钩子只有 `AfterFind` 一个**——它没有 `BeforeQuery`，而且 `AfterFind` 是在**每条记录扫描完之后**触发的（源码 `callbacks/query.go` 的 `AfterQuery` 函数），不是查询前。

```
写操作回调链：  BEGIN → 前钩子 → INSERT/UPDATE/DELETE → 后钩子 → COMMIT/ROLLBACK
查询回调链：    构建 SQL → 执行 → Scan → AfterFind（每条）
```

---

## 1. 检索单个对象：First / Take / Last

### 1.1 官方原文

```go
// 获取第一条记录（主键升序）
db.First(&user)
// SELECT * FROM users ORDER BY id LIMIT 1;

// 获取一条记录，没有指定排序字段
db.Take(&user)
// SELECT * FROM users LIMIT 1;

// 获取最后一条记录（主键降序）
db.Last(&user)
// SELECT * FROM users ORDER BY id DESC LIMIT 1;

result := db.First(&user)
result.RowsAffected // 返回找到的记录数
result.Error        // returns error or nil

// 检查 ErrRecordNotFound 错误
errors.Is(result.Error, gorm.ErrRecordNotFound)
```

### 1.2 源码真相：三者都是 Find 的"带限制变体"

看 `finisher_api.go` 第 130–171 行，`First` 的完整定义：

```go
func (db *DB) First(dest interface{}, conds ...interface{}) (tx *DB) {
	tx = db.Limit(1).Order(clause.OrderByColumn{
		Column: clause.Column{Table: clause.CurrentTable, Name: clause.PrimaryKey},
	})
	if len(conds) > 0 { /* 追加 WHERE */ }
	tx.Statement.RaiseErrorOnNotFound = true   // ← 关键标记
	tx.Statement.Dest = dest
	return tx.callbacks.Query().Execute(tx)
}
```

三个方法的**唯一区别**就是开头那一行：

| 方法 | 本质 | 生成的 SQL |
|---|---|---|
| `First` | `Limit(1)` + `Order(主键 ASC)` | `ORDER BY id LIMIT 1` |
| `Take` | `Limit(1)`（无排序） | `LIMIT 1` |
| `Last` | `Limit(1)` + `Order(主键 DESC)` | `ORDER BY id DESC LIMIT 1` |

实测（`运行输出.txt` 场景 1）三条 SQL 一字不差地印证了这一点：

```
SELECT * FROM `demo_q_users` ORDER BY `demo_q_users`.`id` LIMIT 1       -- First
SELECT * FROM `demo_q_users` LIMIT 1                                    -- Take
SELECT * FROM `demo_q_users` ORDER BY `demo_q_users`.`id` DESC LIMIT 1  -- Last
```

### 1.3 三个方法共同点：都带 `RaiseErrorOnNotFound = true`

这就是它们和 `Find` 的**本质区别**。`Find` 不设置这个标记，所以查不到也不报错；`First/Take/Last` 设置了，查不到会返回 `ErrRecordNotFound`。

精确的产生位置在 `scan.go` 第 366–368 行（扫描全部完成后）：

```go
if db.RowsAffected == 0 && db.Statement.RaiseErrorOnNotFound && db.Error == nil {
	db.AddError(ErrRecordNotFound)
}
```

三个条件**同时**满足才报错：

1. `RowsAffected == 0` —— 一条都没查到
2. `RaiseErrorOnNotFound` —— 用的是 First/Take/Last（Find 不满足）
3. `db.Error == nil` —— 没有其他更优先的错误

而 `ErrRecordNotFound` 的定义只是 `errors.New("record not found")`（`logger/logger.go` 第 17 行），所以官方文档里用 `errors.Is(err, gorm.ErrRecordNotFound)` 来判断。

实测（场景 2）：

```
First 空表 → errors.Is(err, ErrRecordNotFound) = true | err = record not found
Find  空表 → err = <nil> | len = 0（不报错，返回空切片）
```

> **一句话**：想"找不到就报错"用 `First`；想"找不到就安静地返回空"用 `Find`。这是两种语义，不是同一个东西的两种写法。

### 1.4 想避免 ErrRecordNotFound 怎么办

官方给了两个思路：

1. 用 `Find` 代替：`db.Limit(1).Find(&user)` —— 查不到返回空，不报错。
2. 直接忽略这个错误（业务上"没有就是没有"）。

```go
err := db.First(&user).Error
if errors.Is(err, gorm.ErrRecordNotFound) {
	// 没找到，走"创建"或"返回空"的逻辑
}
```

### 1.5 主键/模型的判定规则（官方示例的完整版）

`First`/`Last` 按主键排序，**只有目标是指针或通过 `db.Model()` 指定模型时才有效**；如果模型没有主键，就按第一个字段排序。官方给了一个很实用的对照：

```go
var user User
var users []User

// ✅ 目标是指针，能推断模型
db.First(&user)
// SELECT * FROM `users` ORDER BY `users`.`id` LIMIT 1

// ✅ 用 db.Model() 显式指定模型
result := map[string]interface{}{}
db.Model(&User{}).First(&result)
// SELECT * FROM `users` ORDER BY `users`.`id` LIMIT 1

// ❌ 只有 Table 没有 Model，First 无法推断主键 → 报错
result := map[string]interface{}{}
db.Table("users").First(&result)

// ✅ 换成 Take（不需要主键）就没问题
result := map[string]interface{}{}
db.Table("users").Take(&result)

// 没有主键的模型，按第一个字段（Code）排序
type Language struct {
	Code string
	Name string
}
db.First(&Language{})
// SELECT * FROM `languages` ORDER BY `languages`.`code` LIMIT 1
```

> **为什么 `Table("users").First(&result)` 报错？** 因为 `First` 内部要 `Order(主键)`，而 `db.Table()` 没有绑定任何模型，GORM 不知道主键是什么，于是 `clause.PrimaryKey` 拿不到列名。`Take` 不需要主键，所以能用。

---

## 2. 根据主键检索

### 2.1 官方原文

```go
// 数字主键
db.First(&user, 10)
// SELECT * FROM users WHERE id = 10;

db.First(&user, "10")
// SELECT * FROM users WHERE id = 10;

// 多个主键
db.Find(&users, []int{1,2,3})
// SELECT * FROM users WHERE id IN (1,2,3);

// 字符串主键（比如 uuid）
db.First(&user, "id = ?", "1b74413f-f3b8-409f-ac47-e8c062e3472a")
// SELECT * FROM users WHERE id = "1b74413f-f3b8-409f-ac47-e8c062e3472a";
```

### 2.2 关键机制：目标结构体带非零主键 → 自动加 WHERE

这是查询里最"魔法"、也最容易被忽略的一点。源码在 `callbacks/query.go` 的 `BuildQuerySQL` 第 47–58 行：

```go
if db.Statement.ReflectValue.Kind() == reflect.Struct && db.Statement.ReflectValue.Type() == db.Statement.Schema.ModelType {
	for _, primaryField := range db.Statement.Schema.PrimaryFields {
		if v, isZero := primaryField.ValueOf(ctx, db.Statement.ReflectValue); !isZero {
			conds = append(conds, clause.Eq{Column: ..., Value: v})
		}
	}
	if len(conds) > 0 {
		db.Statement.AddClause(clause.Where{Exprs: conds})
	}
}
```

翻译成人话：**如果你传入的目标结构体，主键字段的值不是零值，GORM 就自动把它变成 `WHERE 主键 = 这个值`**。

```go
var user = User{ID: 10}
db.First(&user)
// SELECT * FROM users WHERE id = 10;   ← 因为 user.ID 已经是 10

var result User
db.Model(User{ID: 10}).First(&result)
// SELECT * FROM users WHERE id = 10;   ← 模型里带了 ID=10 也一样
```

实测（场景 3）：

```
var u3 User; u3.ID = 3; db.First(&u3)
→ SELECT * FROM `demo_q_users` WHERE `demo_q_users`.`id` = 3 ORDER BY `demo_q_users`.`id` LIMIT 1
→ 取回 id=3 name=charlie
```

### 2.3 官方明确提醒的坑：主键条件不会覆盖，而是 AND 叠加

```go
var user = User{ID: 10}
db.Where("id = ?", 20).First(&user)
// SELECT * FROM users WHERE id = 10 AND id = 20 ORDER BY id ASC LIMIT 1
// ↑ 会报 record not found
```

实测（场景 4）完全一致，而且我特意让 `id=3` 这条记录真实存在（charlie）：

```
var u4 User; u4.ID = 3; db.Where("id = ?", 20).First(&u4)
→ SELECT * FROM `demo_q_users` WHERE id = 20 AND `demo_q_users`.`id` = 3 ORDER BY ... LIMIT 1
→ err = record not found（id=3 存在却查不到，因为叠加了 id=20）
```

**根因**：`First` 内部 `BuildCondition("id = ?", 20)` 加了一个 WHERE 子句，`BuildQuerySQL` 又因为结构体主键非零加了第二个 WHERE 子句，两个 WHERE 子句合并成 `AND`。

**解法**：想用同一个变量重新查，先把主键清成零值，或者干脆用新变量。

### 2.4 延伸坑：复用同一个目标变量会带主键残留

这比官方提醒的还隐蔽——**你甚至没写 `Where`，只是复用了上个查询的变量**，主键残留就自己来了。实测（场景 1b）：

```
var u1b User
db.First(&u1b)  → SELECT * FROM `demo_q_users` ORDER BY id LIMIT 1    （u1b.ID 被回写成 1）
db.Take(&u1b)   → SELECT * FROM `demo_q_users` WHERE id = 1 LIMIT 1     （多了 WHERE id=1！）
```

第二条 `Take` 凭空多出 `WHERE id = 1`，就是第一条 `First` 把主键回写进 `u1b` 后，第二条查询又触发了 2.2 节的"自动加 WHERE"机制。

**规则**：**查询操作不要复用目标变量**。要么每次 `var u User` 新变量，要么用完把它主键清零（`u.ID = 0`）。

### 2.5 软删除模型的主键检索

如果模型带 `gorm.DeletedAt`，查询会自动追加 `deleted_at IS NULL`：

```go
type User struct {
	ID        string `gorm:"primarykey;size:16"`
	Name      string `gorm:"size:24"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

var user = User{ID: 15}
db.First(&user)
// SELECT * FROM `users` WHERE `users`.`id` = '15' AND `users`.`deleted_at` IS NULL ORDER BY `users`.`id` LIMIT 1
```

这个"自动过滤软删除"在 **所有查询** 里都生效，不只是主键检索。实测（场景 14）：

```
db.Find(&softs)          → SELECT * FROM `demo_q_soft` WHERE `demo_q_soft`.`deleted_at` IS NULL   （1 条）
db.Unscoped().Find(&softs) → SELECT * FROM `demo_q_soft`                                            （2 条，含已删）
```

> 想连已软删除的一起查，就 `.Unscoped()`。软删除的完整机制在 Update/Delete 文档里展开，这里你只需要记住：**带 `gorm.DeletedAt` 的模型，查询默认只看得见没删的。**

---

## 3. 检索全部对象：Find

```go
result := db.Find(&users)
// SELECT * FROM users;

result.RowsAffected // 找到的记录数，等于 len(users)
result.Error        // 错误
```

几个官方没明说但重要的点：

1. **`Find` 不设 `RaiseErrorOnNotFound`**，所以查不到返回空切片 + `nil` error（见 1.3 节）。
2. **`db.Find(&user)` 传单个结构体指针时**，会全表扫描但只取第一条，**结果不确定且低效**。官方原文特别警告了这一点：

> Using `Find` without a limit for single object `db.Find(&user)` will query the full table and return only the first object which is non-deterministic and not performant

   要单条就用 `First`/`Take`/`Last`，或者 `db.Limit(1).Find(&user)`。

3. `RowsAffected` 对 `Find` 来说就是 `len(users)`。

---

## 4. 条件：Where / Not / Or

### 4.1 字符串条件

```go
db.Where("name = ?", "jinzhu").First(&user)
// SELECT * FROM users WHERE name = 'jinzhu' ORDER BY id LIMIT 1;

db.Where("name <> ?", "jinzhu").Find(&users)
db.Where("name IN ?", []string{"jinzhu", "jinzhu 2"}).Find(&users)  // 切片自动展开成 IN
db.Where("name LIKE ?", "%jin%").Find(&users)
db.Where("name = ? AND age >= ?", "jinzhu", "22").Find(&users)
db.Where("updated_at > ?", lastWeek).Find(&users)
db.Where("created_at BETWEEN ? AND ?", lastWeek, today).Find(&users)
```

**要点：一律用 `?` 占位符 + 参数，防 SQL 注入。** 这是 GORM 最高频的安全红线，`?` 的值会被参数化绑定，不要手拼字符串。

### 4.2 Struct 与 Map 条件

```go
// Struct：只查非零字段
db.Where(&User{Name: "jinzhu", Age: 20}).First(&user)
// SELECT * FROM users WHERE name = "jinzhu" AND age = 20 ORDER BY id LIMIT 1;

// Map：所有 key 都参与
db.Where(map[string]interface{}{"name": "jinzhu", "age": 20}).Find(&users)
// SELECT * FROM users WHERE name = "jinzhu" AND age = 20;

// 主键切片
db.Where([]int64{20, 21, 22}).Find(&users)
// SELECT * FROM users WHERE id IN (20, 21, 22);
```

### 4.3 ★ 查询的零值陷阱（和 Create 不一样！）

官方原文的重点，也是面试高频题：

> **用 struct 查询时，GORM 只用非零字段构建条件**——字段值是 `0`、`""`、`false` 等零值时，不会被用来构建查询条件。

```go
db.Where(&User{Name: "jinzhu", Age: 0}).Find(&users)
// SELECT * FROM users WHERE name = "jinzhu";   ← Age 的 0 被跳过了
```

想查 `age = 0`，**用 map**（map 会把所有 key-value 都当条件）：

```go
db.Where(map[string]interface{}{"Name": "jinzhu", "Age": 0}).Find(&users)
// SELECT * FROM users WHERE name = "jinzhu" AND age = 0;
```

实测（场景 5）：

```
Where(&User{Age:0})   → SELECT * FROM `demo_q_users`            （零值被跳过，等价无条件，5 条）
Where(map{"age":0})   → SELECT * FROM `demo_q_users` WHERE `demo_q_users`.`age` = 0  （1 条）
```

> ⚠️ **注意一个隐蔽的坑：map 的 key 要和字段的 db 名/字段名一致。** 如果写成 `map[string]any{"Age": 0}`（大写），GORM 在 schema 里找不到字段 `Age`，会把它当成**原始列名**直接拼进 SQL（`WHERE Age = 0`）。MySQL 列名不区分大小写所以能查到，但换到 PostgreSQL 这种区分大小写的库就会报"列不存在"。**map 的 key 用小写字段名（`age`），别用 Go 结构体字段名（`Age`）。**

> **和 Create 的对照**：Create 里"零值会不会被写进 INSERT"的规则是——**只有带 `default` tag 的字段零值才会被替换**；而 Query 里"零值会不会被写进 WHERE"的规则是——**struct 条件的零值一律被跳过**。两套规则别混。

### 4.4 指定结构体查询字段

用 struct 查询时，可以显式指定"用 struct 里的哪些字段"来构建条件：

```go
db.Where(&User{Name: "jinzhu"}, "name", "Age").Find(&users)
// SELECT * FROM users WHERE name = "jinzhu" AND age = 0;

db.Where(&User{Name: "jinzhu"}, "Age").Find(&users)
// SELECT * FROM users WHERE age = 0;
```

**语义**：第二个参数往后是**白名单**——只把列出的字段拿来当条件，没列出的字段（哪怕非零）也会被忽略。

实测（场景 6）：

```
Where(&User{Name:alice}, "Age")          → WHERE `age` = 0                （Name 被忽略，只查 Age）
Where(&User{Name:alice}, "name", "Age")  → WHERE `name`='alice' AND `age`=0  （两个字段都查）
```

> 这个特性很适合：结构体里有非零的 `CreatedAt`、`UpdatedAt` 之类的自动字段，你不想让它们进查询条件时，用白名单只挑你关心的字段。

### 4.5 内联条件

条件可以直接塞进 `First`/`Find` 等方法里，等价于先 `Where`：

```go
db.First(&user, "id = ?", "string_primary_key")
db.Find(&user, "name = ?", "jinzhu")
db.Find(&users, "name <> ? AND age > ?", "jinzhu", 20)
db.Find(&users, User{Age: 20})                    // 内联 struct
db.Find(&users, map[string]interface{}{"age": 20}) // 内联 map
```

实测（场景 12）：

```
Find(&us, "age > ?", 18)          → WHERE age > 18      （3 条）
Find(&one, "name = ?", "alice")   → WHERE name = 'alice' （1 条）
```

### 4.6 Not 条件

```go
db.Not("name = ?", "jinzhu").First(&user)
// SELECT * FROM users WHERE NOT name = "jinzhu" ORDER BY id LIMIT 1;

db.Not(map[string]interface{}{"name": []string{"jinzhu", "jinzhu 2"}}).Find(&users)
// SELECT * FROM users WHERE name NOT IN ("jinzhu", "jinzhu 2");

db.Not(User{Name: "jinzhu", Age: 18}).First(&user)
// SELECT * FROM users WHERE name <> "jinzhu" AND age <> 18 ORDER BY id LIMIT 1;

db.Not([]int64{1,2,3}).First(&user)
// SELECT * FROM users WHERE id NOT IN (1,2,3) ORDER BY id LIMIT 1;
```

注意 `Not(struct)` 的语义是 **`<>`（不等于）**，不是 `NOT (name=... AND age=...)`，这个区别在多个字段时容易搞混。

### 4.7 Or 条件

```go
db.Where("role = ?", "admin").Or("role = ?", "super_admin").Find(&users)
// SELECT * FROM users WHERE role = 'admin' OR role = 'super_admin';

// Or + struct/map 会自动加括号分组
db.Where("name = 'jinzhu'").Or(User{Name: "jinzhu 2", Age: 18}).Find(&users)
// SELECT * FROM users WHERE name = 'jinzhu' OR (name = 'jinzhu 2' AND age = 18);

db.Where("name = 'jinzhu'").Or(map[string]interface{}{"name": "jinzhu 2", "age": 18}).Find(&users)
// SELECT * FROM users WHERE name = 'jinzhu' OR (name = 'jinzhu 2' AND age = 18);
```

实测（场景 11）：

```
Not("name = ?", "alice")               → WHERE NOT name = 'alice'   （4 条）
Where("name=?").Or("name=?")           → WHERE name = 'alice' OR name = 'bob'  （2 条）
```

> 更复杂的条件分组（嵌套括号）在高级查询文档 `advanced_query.html` 的 Group Conditions 一节。

---

## 5. 选择特定字段：Select

```go
db.Select("name", "age").Find(&users)
// SELECT name, age FROM users;

db.Select([]string{"name", "age"}).Find(&users)
// SELECT name, age FROM users;

db.Table("users").Select("COALESCE(age,?)", 42).Rows()
// SELECT COALESCE(age,'42') FROM users;
```

### 5.1 智能选择字段

官方原文说"`Select` 允许你指定要取的字段，否则 GORM 默认取所有字段"。但还有一层**隐藏的智能行为**（源码 `BuildQuerySQL` 第 79–100 行）：

**如果目标结构体不是模型类型本身（比如一个只有部分字段的小结构体），GORM 会只查这个小结构体里有的列。**

实测（场景 7），目标是小结构体 `NameOnly{ID, Name}`，即使不写 `Select`，也会只查 `id, name`：

```go
type NameOnly struct { ID int64; Name string }
var names []NameOnly
db.Model(&User{}).Find(&names)     // 只查 id, name，而不是 SELECT *
```

这在你只想取部分列、又懒得写 `Select` 时非常省事，还能减少 IO。不过为了代码可读性，**明确写 `Select` 是更推荐的做法**（意图清晰，不依赖隐式行为）。

---

## 6. 排序 / 分页：Order / Limit / Offset

### 6.1 Order

```go
db.Order("age desc, name").Find(&users)
// SELECT * FROM users ORDER BY age desc, name;

// 多次 Order 是"叠加"而不是"覆盖"
db.Order("age desc").Order("name").Find(&users)
// SELECT * FROM users ORDER BY age desc, name;

// 复杂排序（比如 FIELD 自定义顺序）
db.Clauses(clause.OrderBy{
	Expression: clause.Expr{SQL: "FIELD(id,?)", Vars: []interface{}{[]int{1, 2, 3}}, WithoutParentheses: true},
}).Find(&User{})
// SELECT * FROM users ORDER BY FIELD(id,1,2,3)
```

实测（场景 9）：

```
db.Order("age desc").Order("name")  → ORDER BY age desc,name
db.Clauses(clause.OrderBy{...FIELD(id,3,1,2)...}) → ORDER BY FIELD(id,3,1,2)
```

`FIELD(id, 3,1,2)` 是 MySQL 的"按指定顺序排"语法——把 id 为 3 的排最前、1 次之、2 最后。适合"置顶某几条"的场景。

### 6.2 Limit & Offset

```go
db.Limit(3).Find(&users)
// SELECT * FROM users LIMIT 3;

// Limit(-1) 取消 limit
db.Limit(10).Find(&users1).Limit(-1).Find(&users2)

db.Offset(3).Find(&users)
// SELECT * FROM users OFFSET 3;

db.Limit(10).Offset(5).Find(&users)
// SELECT * FROM users OFFSET 5 LIMIT 10;

// Offset(-1) 取消 offset
db.Offset(10).Find(&users1).Offset(-1).Find(&users2)
```

实测（场景 8）：

```
Limit(2)            → SELECT * FROM `demo_q_users` LIMIT 2       （2 条）
Limit(2).Limit(-1)  → SELECT * FROM `demo_q_users`               （5 条，limit 被取消）
```

**`Limit(-1)` / `Offset(-1)` 是"取消该限制"的魔法值**。这在封装"可选分页"的函数时很有用：

```go
func ListUsers(page, pageSize int) []User {
	tx := db
	if pageSize > 0 {
		tx = tx.Limit(pageSize).Offset((page - 1) * pageSize)
	}
	// pageSize <= 0 时就不加分页
	var users []User
	tx.Find(&users)
	return users
}
```

分页器的完整封装（总页数、上一页/下一页）官方指路到 `scopes.html` 的 Pagination 一节。

---

## 7. 分组聚合：Group By / Having / Distinct

### 7.1 Group By & Having

```go
type result struct {
	Date  time.Time
	Total int
}

db.Model(&User{}).Select("name, sum(age) as total").Where("name LIKE ?", "group%").Group("name").First(&result)
// SELECT name, sum(age) as total FROM `users` WHERE name LIKE "group%" GROUP BY `name` LIMIT 1

db.Model(&User{}).Select("name, sum(age) as total").Group("name").Having("name = ?", "group").Find(&result)
// SELECT name, sum(age) as total FROM `users` GROUP BY `name` HAVING name = "group"

// 用 Rows() 逐行读
rows, err := db.Table("orders").Select("date(created_at) as date, sum(amount) as total").Group("date(created_at)").Rows()
defer rows.Close()
for rows.Next() { ... }

// 直接 Scan 到结果结构体
type Result struct {
	Date  time.Time
	Total int64
}
db.Table("orders").Select("date(created_at) as date, sum(amount) as total").
	Group("date(created_at)").Having("sum(amount) > ?", 100).Scan(&results)
```

实测（场景 10），按 `age` 分组计数：

```
SELECT age, count(*) as total FROM `demo_q_users` GROUP BY `age` HAVING count(*) >= 1 ORDER BY age
→ age=0 有 1 人 | age=18 有 1 人 | age=20 有 2 人 | age=25 有 1 人
```

要点：

1. **聚合列要起别名，且别名要和结果结构体的字段名对上**——`count(*) as total` 对应 `Total int64`。
2. **`Having` 在 `Group` 之后**（和 SQL 语义一致），`Having` 过滤的是分组后的结果，`Where` 过滤的是分组前的行。
3. 结果不是模型本身（是聚合结果），所以用 `.Scan(&result)` 或 `.Rows()`，不用 `.Find(&Model{})`。

### 7.2 Distinct

```go
db.Distinct("name", "age").Order("name, age desc").Find(&results)
```

`Distinct` 也能和 `Pluck`、`Count` 配合（这两个在高级查询文档）。

---

## 8. Joins

### 8.1 传统写法

```go
type result struct {
	Name  string
	Email string
}

db.Model(&User{}).Select("users.name, emails.email").Joins("left join emails on emails.user_id = users.id").Scan(&result{})
// SELECT users.name, emails.email FROM `users` left join emails on emails.user_id = users.id

// 多个 join，带参数
db.Joins("JOIN emails ON emails.user_id = users.id AND emails.email = ?", "jinzhu@example.org").
	Joins("JOIN credit_cards ON credit_cards.user_id = users.id").
	Where("credit_cards.number = ?", "411111111111").Find(&user)
```

### 8.2 Joins 预加载（关联查询）

`Joins` 还能在**一条 SQL** 里预加载关联：

```go
db.Joins("Company").Find(&users)
// SELECT `users`.`id`,`users`.`name`,...,`Company`.`id` AS `Company__id`,... FROM `users` LEFT JOIN `companies` AS `Company` ON `users`.`company_id` = `Company`.`id`;

db.InnerJoins("Company").Find(&users)   // INNER JOIN 版本
db.Joins("Company", db.Where(&Company{Alive: true})).Find(&users)   // 带条件
```

> 关联查询（Preload / Joins 预加载）是 GORM 最庞大的一块，完整机制在 `preload.html`。这里先记住：**`Joins("关联名")` 是"一条 SQL 连表"，`Preload("关联名")` 是"先查主表、再发第二条 SQL 查关联"**——两者取舍（N+1 问题）到关联文档再展开。

### 8.3 Joins 一个衍生表（子查询 join）

```go
query := db.Table("order").Select("MAX(order.finished_at) as latest").
	Joins("left join user user on order.user_id = user.id").
	Where("user.age > ?", 18).Group("order.user_id")

db.Model(&Order{}).Joins("join (?) q on order.finished_at = q.latest", query).Scan(&results)
// SELECT ... FROM `order` join (SELECT MAX(...) ... GROUP BY ...) q on order.finished_at = q.latest
```

把整个 `query` 当参数塞进 `Joins("...(?)", query)`，`?` 会被替换成那个子查询的完整 SQL。

> Generics API 也有 `Joins`，且对关联查询支持更强（`clause.LeftJoin.Association` 等），但那是较新的用法。学习阶段以传统 API 为主。

---

## 9. Scan

`Scan` 把结果扫进一个**和模型无关**的结构体，类似 `Find` 但更自由：

```go
type Result struct {
	Name string
	Age  int
}

var result Result
db.Table("users").Select("name", "age").Where("name = ?", "Antonio").Scan(&result)

// 裸 SQL + Scan
db.Raw("SELECT name, age FROM users WHERE name = ?", "Antonio").Scan(&result)
```

实测（场景 13）：

```
db.Table("demo_q_users").Select("name", "age").Where("age > ?", 0).Scan(&rows)
→ alice 18 | bob 20 | charlie 20 | david 25
```

`Scan` 和 `Find` 的区别：**`Find` 要求目标是模型（或模型切片），会触发模型相关的行为（软删除过滤、`AfterFind` 钩子）；`Scan` 目标可以是任意结构体，绕开模型，直接映射列。** 聚合结果、跨表结果、裸 SQL 结果，用 `Scan`。

---

## 10. 坑点速查表

| # | 坑 | 现象 | 解法 |
|---|---|---|---|
| 1 | 复用目标变量查询 | 第二条查询凭空多出 `WHERE id = ?` | 每次 `var u User`，或查询后清主键 |
| 2 | struct 条件查零值 | `Where(&User{Age:0})` 查不到 age=0 | 用 `map[string]any{"age": 0}` |
| 3 | map 的 key 用大写 | 生成 `WHERE Age = 0`，PG 下报列不存在 | key 用小写字段名 `age` |
| 4 | `First` 查不到 vs `Find` 查不到 | 前者报 `ErrRecordNotFound`，后者静默 | 按语义选；用 `errors.Is(err, gorm.ErrRecordNotFound)` 判断 |
| 5 | 主键非零 + `Where` 同列 | `WHERE id = 10 AND id = 20` 查不到 | 查前清主键或用新变量 |
| 6 | 复用链式对象 | 条件跨查询累积 | 用 `.Session(&gorm.Session{})` 或新起 `db` |
| 7 | `db.Find(&user)` 取单条 | 全表扫描只取第一条，慢且不确定 | 用 `First`/`Take`/`Last` |
| 8 | 软删除模型查询 | 已删的记录"消失"了 | 需要时 `.Unscoped()` |
| 9 | 字符串条件手拼 SQL | SQL 注入 | 一律 `?` 占位符 |
| 10 | 聚合结果用 `Find(&Model{})` | 字段对不上、扫不进去 | 用 `.Scan(&result)` 或 `.Rows()` |
| 11 | `Order` 以为是覆盖 | 两次 `Order` 其实是叠加 | 要覆盖就重新起链 |

---

## 11. 动手练习

完整可运行代码在 **`GORM/查询Query/main.go`**（16 个场景，跑完自动删表，输出在 `运行输出.txt`）。以下是递进任务清单，建议在你自己的 `CRUD/main.go` 里逐条跑一遍，`logger.Info` 开着看 SQL：

- [ ] 用 `First` / `Take` / `Last` 各查一次，对比三条 SQL 的 `ORDER BY` 差异
- [ ] 查一个不存在的 id，确认 `First` 报 `ErrRecordNotFound`、`Find` 不报
- [ ] 故意复用同一个变量连查两次，观察第二次的 `WHERE id = ?` 是怎么来的
- [ ] 用 `Where(&User{Age: 0})` 和 `Where(map[string]any{"age": 0})` 各查一次，对比结果
- [ ] 试 `Where(&User{Name: "x"}, "Age")` 和 `Where(&User{Name: "x"}, "name", "Age")`，理解白名单语义
- [ ] 用 `Limit(2).Limit(-1)` 验证 limit 被取消
- [ ] 写一个 Group By + Having 的聚合查询，`Scan` 到结果结构体
- [ ] 给模型加 `gorm.DeletedAt`，对比 `Find` 和 `Unscoped().Find` 的结果
- [ ] 给模型加 `AfterFind`，打印日志确认"每条记录触发一次"
- [ ] 把 `Where` 条件换成 `?` 占位符形式，确认生成的 SQL 参数化（防注入）

---

> **下一步**：这一页是"查询"的地基。更深的查询能力在 `advanced_query.html`（子查询、Pluck、Count、FirstOrCreate、锁、Group Conditions）——建议先吃透本页的 16 个场景，再进高级查询。
