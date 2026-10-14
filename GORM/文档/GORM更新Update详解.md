# GORM 更新 Update 详解

> **官方原文**：https://gorm.io/zh_CN/docs/update.html
> **本地环境**：`gorm.io/gorm v1.31.2` + `gorm.io/driver/mysql v1.6.0`（MySQL 8.0，库 `sql_test`）
> **本文档在原页基础上额外做了**：① 逐条核对源码（`finisher_api.go` / `callbacks/update.go` / `callbacks/helper.go` / `statement.go`）；② 写了一个 15 场景的可运行实测程序（`GORM/更新Update/main.go`），把每个结论都用**真实 SQL 日志**钉死；③ 补充了官方页面没讲的 5 个坑。
> **配套文件**：`GORM/更新Update/main.go`（可运行）、`GORM/更新Update/运行输出.txt`（真实日志）

---

## 目录

- [0. 全局视角：更新的回调链](#0-全局视角更新的回调链)
- [1. Save —— 保存所有字段](#1-save--保存所有字段)
- [2. Update —— 更新单个列](#2-update--更新单个列)
- [3. Updates —— 更新多个列](#3-updates--更新多个列)
- [4. Select / Omit —— 更新选定字段](#4-select--omit--更新选定字段)
- [5. 更新钩子](#5-更新钩子)
- [6. 批量更新与阻止全局更新](#6-批量更新与阻止全局更新)
- [7. RowsAffected —— 更新的记录数](#7-rowsaffected--更新的记录数)
- [8. 高级选项](#8-高级选项)
- [9. 坑点速查表](#9-坑点速查表)
- [10. 动手练习](#10-动手练习)

---

## 0. 全局视角：更新的回调链

和 Create 一样，Update 也是一条**固定顺序的回调链**。先把它背下来，后面所有小节（Save 的 upsert、钩子、UpdateColumn 跳过）都能各归其位：

```
BEGIN（写操作默认开事务）
 ├─ setup_reflect_value        ← 确定「更新的目标模型」是谁
 ├─ before_update              ← 触发 BeforeSave → BeforeUpdate
 ├─ save_before_associations   ← 关联保存
 ├─ update                     ← 真正执行 UPDATE（内部含「无条件保护」检查）
 ├─ save_after_associations
 ├─ after_update               ← 触发 AfterUpdate → AfterSave
 └─ COMMIT / ROLLBACK
```

源码依据 `callbacks/callbacks.go:66-75`。**更新钩子的真实顺序是**：

```
BeforeSave → BeforeUpdate → UPDATE → AfterUpdate → AfterSave
```

注意：**前钩子 Save 在前，后钩子 Update 在前**——这和 Create 的 `BeforeSave→BeforeCreate→INSERT→AfterCreate→AfterSave` 完全对称（后钩子都是"具体操作"在前、"Save"在后）。

> ⚠️ 官方页面「更新 Hook」一节把钩子列成 `BeforeSave, BeforeUpdate, AfterSave, AfterUpdate`——那是"有哪些钩子"的清单，**不是执行顺序**。别被它误导。

---

## 1. Save —— 保存所有字段

### 1.1 官方原文

```go
db.First(&user)

user.Name = "jinzhu 2"
user.Age = 100
db.Save(&user)
// UPDATE users SET name='jinzhu 2', age=100, birthday='2016-01-01', updated_at = '2013-11-17 21:34:10' WHERE id=111;
```

`Save` 是一个 **upsert** 函数：

- 如果值**不含主键**，执行 `Create`；
- 如果值**有主键**，先执行 `Update`（通过 `Select(*)` 更新**所有字段**）；
- 如果 `Update` 之后 `rows affected = 0`，自动**回退到 `Create`**。

```go
db.Save(&User{Name: "jinzhu", Age: 100})
// INSERT INTO `users` (`name`,`age`,`birthday`,`update_at`) VALUES ("jinzhu",100,"0000-00-00 00:00:00","0000-00-00 00:00:00")

db.Save(&User{ID: 1, Name: "jinzhu", Age: 100})
// UPDATE `users` SET `name`="jinzhu",`age`=100,`birthday`="0000-00-00 00:00:00",`update_at`="0000-00-00 00:00:00" WHERE `id` = 1
```

> **NOTE** 不要将 `Save` 和 `Model` 一同使用，这是**未定义的行为**。

> **NOTE** `Save` 方法被有意从 Generics API 中移除，以避免歧义和并发问题。请改用 `Create` 或 `Updates` 方法。

### 1.2 源码：Save 的真实三分支

`finisher_api.go:85-127` 的 `Save` 逻辑，简化后长这样：

```go
func (db *DB) Save(value interface{}) (tx *DB) {
    // ① 切片 → 走 Create（带 OnConflict UpdateAll）
    if reflectValue.Kind() == Slice/Array {
        return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(value)
    }

    // ② struct 且主键全为零 → 直接 Create
    for _, pf := range Schema.PrimaryFields {
        if isZero { return tx.callbacks.Create().Execute(tx) }
    }

    // ③ struct 有主键 → Select("*") 全字段 Update
    if !selectedUpdate {
        tx.Statement.Selects = append(tx.Statement.Selects, "*")   // ← 关键：Save 强制 Select(*)
    }
    updateTx := tx.callbacks.Update().Execute(...)

    // ④ Update 影响 0 行 → 回退 Create（SkipHooks + OnConflict UpdateAll）
    if updateTx.Error == nil && updateTx.RowsAffected == 0 {
        return tx.Session(&Session{SkipHooks: true}).
            Clauses(clause.OnConflict{UpdateAll: true}).Create(value)
    }
    return updateTx
}
```

### 1.3 实测（真实 SQL）

**场景 1 · Save 有主键 = 更新所有字段（零值也写）**

```go
var u DemoUser
db.First(&u)        // 取出 id=1 的 alice（created_at 有值）
u.Name = "alice_v2"
u.Age = 0           // 故意零值
u.Role = ""
u.Active = false
db.Save(&u)
```

```sql
UPDATE `demo_u_users` SET
  `created_at`='2026-09-19 21:31:33.883', `updated_at`='2026-09-19 21:31:33.894',
  `deleted_at`=NULL, `name`='alice_v2', `age`=0, `role`='', `active`=false
WHERE `demo_u_users`.`deleted_at` IS NULL AND `id` = 1
```

`age=0`、`role=''`、`active=false` 全写进去了——**这就是 Save 和 Updates(struct) 的本质区别**（后者会跳过零值）。

**场景 2 · Save 无主键 = Create**

```go
db.Save(&DemoUser{Name: "newcomer", Age: 0})
```

```sql
INSERT INTO `demo_u_users` (`created_at`,`updated_at`,`deleted_at`,`name`,`age`,`role`,`active`)
VALUES ('2026-09-19 21:31:33.899','...',NULL,'newcomer',0,'',false)
```

**场景 3 · Save 主键不存在 = 先 Update(0 行) 再回退 Create**

```go
db.Save(&DemoUser{Model: gorm.Model{ID: 99999}, Name: "ghost", Age: 30})
```

```sql
-- 第一条：UPDATE 影响 0 行（id=99999 不存在）
UPDATE `demo_u_users` SET `created_at`='0000-00-00 00:00:00', ..., `name`='ghost', `age`=30
WHERE `demo_u_users`.`deleted_at` IS NULL AND `id` = 99999   -- [rows:0]

-- 第二条：自动回退为 INSERT ... ON DUPLICATE KEY UPDATE
INSERT INTO `demo_u_users` (...,`id`) VALUES (...,99999)
ON DUPLICATE KEY UPDATE `updated_at`=..., `name`=VALUES(`name`), `age`=VALUES(`age`), ...
```

### 1.4 三个必须知道的点

| 点 | 说明 |
|---|---|
| **Save 会覆盖零值** | 因为它内部 `Select(*)`。想"只更新非零值"就用 `Updates` |
| **Save 传"新构造的零值 struct"很危险** | 见场景 3：`created_at` 被写成 `'0000-00-00 00:00:00'`（零值）。若主键**恰好存在**，这条 UPDATE 会真的执行并报错（MySQL 拒绝 `0000-00-00`）——详见 [第 4 节](#4-select--omit--更新选定字段) |
| **Save 和 Model 别混用** | 官方 NOTE 说是"未定义行为"。原因是 `Save` 内部会 `tx.Statement.Parse(value)` 并把 `Dest = value`，会**覆盖**你 `Model()` 传进去的模型 |

> **Generics API 里没有 `Save`**：`gorm.G[User](db)` 上不提供 `Save` 方法。官方理由是避免"upsert 语义"带来的歧义和并发问题。要做就显式 `Create` 或 `Updates`。

---

## 2. Update —— 更新单个列

### 2.1 官方原文

```go
// 根据条件更新
db.Model(&User{}).Where("active = ?", true).Update("name", "hello")
// UPDATE users SET name='hello', updated_at='...' WHERE active=true;

// User 的 ID 是 111
db.Model(&user).Update("name", "hello")
// UPDATE users SET name='hello', updated_at='...' WHERE id=111;

// 根据条件和 model 的值进行更新
db.Model(&user).Where("active = ?", true).Update("name", "hello")
// UPDATE users SET name='hello', updated_at='...' WHERE id=111 AND active=true;
```

### 2.2 机制与实测

`Update` 的源码（`finisher_api.go:413`）就两行：

```go
func (db *DB) Update(column string, value interface{}) (tx *DB) {
    tx = db.getInstance()
    tx.Statement.Dest = map[string]interface{}{column: value}   // 单列被包装成 map
    return tx.callbacks.Update().Execute(tx)
}
```

**`Model` 带主键时，主键自动成为 WHERE 条件**（实测场景 4）：

```go
var u4 DemoUser
db.Where("name = ?", "bob").First(&u4)   // u4.ID = 2
db.Model(&u4).Update("role", "vip")
```

```sql
UPDATE `demo_u_users` SET `role`='vip',`updated_at`='...' WHERE `demo_u_users`.`deleted_at` IS NULL AND `id` = 2
```

**注意 `updated_at` 被自动追加了**——`Update`/`Updates` 会追踪 `UpdatedAt` 字段（`gorm.Model` 里的 `UpdatedAt`），每次更新都自动刷新。这个细节在第 7 节会引出大坑。

> Generics API 对应写法：`gorm.G[User](db).Where("id = ?", 111).Update(ctx, "name", "hello")`。原文里的 `gorm.GUser` 是 `gorm.G[User]` 的示意简写。

---

## 3. Updates —— 更新多个列

### 3.1 官方原文

```go
// 根据 struct 更新，只会更新非零值字段
db.Model(&user).Updates(User{Name: "hello", Age: 18, Active: false})
// UPDATE users SET name='hello', age=18, updated_at = '...' WHERE id = 111;

// 根据 map 更新
db.Model(&user).Updates(map[string]interface{}{"name": "hello", "age": 18, "active": false})
// UPDATE users SET name='hello', age=18, active=false, updated_at='...' WHERE id=111;
```

> **注意** 使用 struct 更新时，GORM 只更新非零值字段。你可能想用 `map` 来更新属性，或者使用 `Select` 声明字段来更新。

### 3.2 实测：struct 跳过零值 vs map 保留零值

**场景 6 · Updates(struct)**

```go
db.Model(&DemoUser{}).Where("id = ?", 1).
    Updates(DemoUser{Name: "alice_v3", Age: 0, Role: "", Active: false})
```

```sql
UPDATE `demo_u_users` SET `updated_at`='...', `name`='alice_v3'
WHERE id = 1 AND `demo_u_users`.`deleted_at` IS NULL
```

`age=0`、`role=''`、`active=false` **全被跳过**——SET 里只剩 `name` 和 `updated_at`。

**场景 7 · Updates(map)**

```go
db.Model(&DemoUser{}).Where("id = ?", 2).
    Updates(map[string]any{"age": 0, "active": false})
```

```sql
UPDATE `demo_u_users` SET `active`=false, `age`=0, `updated_at`='...'
WHERE id = 2 AND `demo_u_users`.`deleted_at` IS NULL
```

`age=0`、`active=false` **全写进去**——map 的每个 key 都是"显式存在的"，零值也照写。

### 3.3 一句话记忆

| 参数类型 | 零值字段（0 / "" / false） | 适用场景 |
|---|---|---|
| `struct` | **跳过**（不更新） | 只想改"确实有值"的字段 |
| `map[string]any` | **保留**（更新为 0） | 想把字段真正清零 / 置 false |

> 和 Create 的规则对比：Create 里**只有带 `default` tag 的字段**零值才被替换；而 Update 的 `struct` 参数是**所有零值字段无条件跳过**。两处的"零值规则"不完全一样，别混。

---

## 4. Select / Omit —— 更新选定字段

### 4.1 官方原文

```go
// Select with Map —— 只更新 name
db.Model(&user).Select("name").Updates(map[string]interface{}{"name": "hello", "age": 18, "active": false})
// UPDATE users SET name='hello' WHERE id=111;

// Omit with Map —— 除了 name 都更新
db.Model(&user).Omit("name").Updates(map[string]interface{}{"name": "hello", "age": 18, "active": false})
// UPDATE users SET age=18, active=false, updated_at='...' WHERE id=111;

// Select with Struct —— 强制更新零值字段
db.Model(&user).Select("Name", "Age").Updates(User{Name: "new_name", Age: 0})
// UPDATE users SET name='new_name', age=0 WHERE id=111;

// Select all fields —— 更新所有字段（包括零值）
db.Model(&user).Select("*").Updates(User{Name: "jinzhu", Role: "admin", Age: 0})
```

### 4.2 实测：Select("Age") 安全 vs Select("*") 危险

**场景 8 上半 · `Select("Age")`（安全）**

```go
db.Model(&DemoUser{}).Where("id = ?", 1).
    Select("Age").Updates(DemoUser{Name: "ignored", Age: 0})
```

```sql
UPDATE `demo_u_users` SET `updated_at`='...', `age`=0 WHERE id = 1 AND ...
```

只更新了 `age=0`，`Name: "ignored"` 被忽略——`Select("Age")` 限定了字段白名单。

**场景 8 下半 · `Select("*")`（危险）**

```go
db.Model(&DemoUser{}).Where("id = ?", 1).
    Select("*").Updates(DemoUser{Name: "alice_v4", Age: 0, Role: "", Active: false})
```

```sql
-- 报错：Error 1292 (22007): Incorrect datetime value: '0000-00-00' for column 'created_at' at row 1
UPDATE `demo_u_users` SET `id`=0, `created_at`='0000-00-00 00:00:00', `updated_at`='...',
  `deleted_at`=NULL, `name`='alice_v4', `age`=0, `role`='', `active`=false
WHERE id = 1 AND ...
```

### 4.3 ⚠️ `Select("*")` 的真实含义

官方页面只说"更新所有字段包括零值"，**但没警告它连主键和审计字段也一起更新**。看上面那条 SQL：

- `id=0` —— 主键被写成了 0；
- `created_at='0000-00-00 00:00:00'` —— 因为 struct 的 `CreatedAt` 是零值，被写成了非法日期，MySQL 直接报错；
- `deleted_at=NULL` —— 软删除字段也进了 SET。

**正确姿势**：`Select("*")` 只该搭配"从数据库查出来的完整对象"用（此时 `created_at` 有真实值）。对**新构造的零值 struct** 用 `Select("*")` 是灾难。要"更新多个指定字段（含零值）"，用 `Select("A", "B")` 明确列出，而不是 `Select("*")`。

> 这也解释了为什么 `Save` 场景 1 没报错、场景 3/8 会报错：场景 1 的 `u` 是 `First` 查出来的，`created_at` 有值；场景 3/8 是 `&DemoUser{...}` 新构造的，`created_at` 是零值。

---

## 5. 更新钩子

### 5.1 可用接口与顺序

```go
func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
    if u.Role == "admin" {
        return errors.New("admin user not allowed to update")
    }
    return
}
```

四个钩子：`BeforeSave`、`BeforeUpdate`、`AfterSave`、`AfterUpdate`。真实顺序（实测场景 12）：

```
BeforeSave → BeforeUpdate → (UPDATE) → AfterUpdate → AfterSave
```

- 接收者是指针，返回值 `error` 会**中止更新并回滚**；
- 钩子里做 DB 操作要用参数 `tx`（否则脱离事务，详见答疑记录 Q1）。

### 5.2 Changed —— 判断字段是否真的变了

`Changed` 只在 `BeforeUpdate` 钩子里用，只对 `Update`/`Updates` 有效。它判断"要更新的值"是否和"模型当前值"不同，返回 `true` 表示字段确实变了（且没被 Omit）。

```go
func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
    if tx.Statement.Changed("Role") {                  // Role 变了
        return errors.New("role not allowed to change")
    }
    if tx.Statement.Changed("Name", "Admin") {         // Name 或 Admin 变了
        tx.Statement.SetColumn("Age", 18)
    }
    if tx.Statement.Changed() {                        // 任意字段变了
        tx.Statement.SetColumn("RefreshedAt", time.Now())
    }
    return nil
}
```

**实测场景 13**（模型带 `BeforeUpdate`，判断 `Changed("Name")` / `Changed("Code")`）：

```go
db.Model(&tr).Updates(map[string]any{"name": "t1_v2"})   // Name 从 t1 → t1_v2，变了
// [BeforeUpdate] Name 发生了变化
// UPDATE ... SET age=18, name='t1_v2', updated_at='...'   ← age 被 SetColumn 强制改成 18

db.Model(&DemoTrack{Model: gorm.Model{ID: tr.ID}, Name: "t1_v2"}).
    Updates(map[string]any{"name": "t1_v2"})              // 值相同，Name 没变
// UPDATE ... SET name='t1_v2', updated_at='...'           ← 什么都没触发

err := db.Model(&DemoTrack{Model: gorm.Model{ID: tr.ID}, Name: "t1_v2"}).
    Updates(map[string]any{"code": "c2"}).Error           // Code 变了 → 拦截
// err = "Code 字段不允许修改"                               ← UPDATE 根本没执行
```

### 5.3 SetColumn —— 在钩子里改要更新的值

在 `Before` 钩子里改更新值，用 `tx.Statement.SetColumn(...)`，而不是直接改结构体字段（除非是 `Save` 的全量更新）：

```go
func (user *User) BeforeSave(tx *gorm.DB) (err error) {
    if pw, err := bcrypt.GenerateFromPassword(user.Password, 0); err == nil {
        tx.Statement.SetColumn("EncryptedPassword", pw)
    }
    if tx.Statement.Changed("Code") {
        user.Age += 20
        tx.Statement.SetColumn("Age", user.Age)
    }
}
```

> `SetColumn` 的签名是 `SetColumn(name string, value interface{}, fromCallbacks ...bool)`——第三个参数区分"钩子里调用"和"回调内部调用"。钩子里正常用两参数即可。

---

## 6. 批量更新与阻止全局更新

### 6.1 官方原文

```go
// 用 struct 批量更新
db.Model(User{}).Where("role = ?", "admin").Updates(User{Name: "hello", Age: 18})
// UPDATE users SET name='hello', age=18 WHERE role = 'admin';

// 用 map 批量更新
db.Table("users").Where("id IN ?", []int{10, 11}).Updates(map[string]interface{}{"name": "hello", "age": 18})
// UPDATE users SET name='hello', age=18 WHERE id IN (10, 11);
```

**阻止全局更新**：

```go
db.Model(&User{}).Update("name", "jinzhu").Error // gorm.ErrMissingWhereClause

db.Model(&User{}).Where("1 = 1").Update("name", "jinzhu")     // ✅ 显式条件
db.Exec("UPDATE users SET name = ?", "jinzhu")                // ✅ 原生 SQL
db.Session(&gorm.Session{AllowGlobalUpdate: true}).
    Model(&User{}).Update("name", "jinzhu")                   // ✅ 会话级放行
```

### 6.2 实测场景 5：无条件更新被拦（连软删除条件都不算）

```go
err := db.Model(&DemoUser{}).Update("role", "hacker").Error
// err = "WHERE conditions required"（errors.Is(err, gorm.ErrMissingWhereClause) = true）
```

看它尝试生成的 SQL：

```sql
UPDATE `demo_u_users` SET `role`='hacker',`updated_at`='...' WHERE `demo_u_users`.`deleted_at` IS NULL
```

**注意**：SQL 里明明有 `WHERE deleted_at IS NULL`，但 GORM 依然拦截。因为源码 `helper.go:109-123` 的 `checkMissingWhereConditions` 知道——`deleted_at IS NULL` 是软删除**自动追加**的，不算"用户提供的条件"。只有当用户自己写了 `Where(...)` 才放行。

两种放行方式都实测通过：

```sql
UPDATE ... WHERE 1 = 1 AND `demo_u_users`.`deleted_at` IS NULL          -- Where("1 = 1")
UPDATE ... WHERE `demo_u_users`.`deleted_at` IS NULL                   -- AllowGlobalUpdate
```

### 6.3 实测场景 14：批量更新的 struct false 陷阱

```go
r1 := db.Model(&DemoUser{}).Where("role = ?", "user").
    Updates(DemoUser{Active: false})     // ❌ false 是零值，被跳过！
// UPDATE ... SET updated_at='...' WHERE role = 'user'  ← SET 里没有 active！

r2 := db.Model(&DemoUser{}).Where("role = ?", "user").
    Updates(map[string]any{"active": false})   // ✅ map 不跳过
// UPDATE ... SET active=false, updated_at='...' WHERE role = 'user'
```

**想批量把 `Active` 改成 `false`，用 struct 会静默失败**（`false` 是零值，被跳过，只更新了 `updated_at`）。批量"清零/置 false"必须用 `map` 或 `Select`。

---

## 7. RowsAffected —— 更新的记录数

### 7.1 官方原文

```go
result := db.Model(User{}).Where("role = ?", "admin").Updates(User{Name: "hello", Age: 18})
result.RowsAffected // 返回更新的记录数
result.Error        // 返回错误
```

### 7.2 ⚠️ 一个官方没讲的坑：`updated_at` 会污染计数

`Update`/`Updates` 会自动更新 `updated_at`。所以即使你的业务字段值**根本没变**，只要 `updated_at` 变了，MySQL 也把这行算作"已修改"。实测场景 15：

```go
rA := db.Model(&DemoUser{}).Where("id = ?", 1).UpdateColumn("active", true) // active 本来就是 true
fmt.Println(rA.RowsAffected) // 0 ← 值没变，MySQL 不计

rB := db.Model(&DemoUser{}).Where("id = ?", 1).Update("active", true)       // active 本来也是 true
fmt.Println(rB.RowsAffected) // 1 ← 但 updated_at 变了，计 1 行
```

```sql
UPDATE `demo_u_users` SET `active`=true WHERE id = 1 ...                    -- [rows:0]（UpdateColumn，无 updated_at）
UPDATE `demo_u_users` SET `active`=true,`updated_at`='...' WHERE id = 1 ... -- [rows:1]（Update，带 updated_at）
```

**结论**：`RowsAffected` 对 `Update`/`Updates` 来说，反映的是"`updated_at` 变了的行数"，不完全是"业务字段真的变了的行数"。要精确判断"某字段是否真的被改"，用 `Changed`（第 5.2 节）更可靠。

---

## 8. 高级选项

### 8.1 使用 SQL 表达式更新

```go
db.Model(&product).Update("price", gorm.Expr("price * ? + ?", 2, 100))
// UPDATE products SET price = price * 2 + 100, updated_at = '...' WHERE id = 3;

db.Model(&product).UpdateColumn("quantity", gorm.Expr("quantity - ?", 1))
// UPDATE products SET quantity = quantity - 1 WHERE id = 3;
```

**实测场景 9**：

```go
db.Model(&DemoUser{}).Where("id = ?", 1).Update("age", gorm.Expr("age + ?", 10))
// UPDATE ... SET `age`=age + 10, `updated_at`='...' WHERE id = 1 ...
```

`gorm.Expr` 是最常用的"读-改-写"原子手段（库存扣减、计数器自增）。**一定要配合 WHERE 条件**，否则又是全局更新。

### 8.2 根据子查询更新

```go
db.Model(&user).Update("company_name", db.Model(&Company{}).Select("name").Where("companies.id = users.company_id"))
// UPDATE users SET company_name = (SELECT name FROM companies WHERE companies.id = users.company_id);
```

**实测场景 10**（跨表）：

```go
db.Model(&DemoUser{}).Where("name = ?", "alice").
    Update("age", db.Model(&DemoMeta{}).Select("age").
        Where("demo_u_metas.name = demo_u_users.name"))
```

```sql
UPDATE `demo_u_users` SET `age`=(SELECT `age` FROM `demo_u_metas`
  WHERE demo_u_metas.name = demo_u_users.name), `updated_at`='...'
WHERE name = 'alice' AND ...
```

> ⚠️ MySQL 有个限制：**子查询不能直接引用正在更新的那张表**（报 `You can't specify target table for update in FROM clause`）。所以同表自更新子查询要套一层派生表（`SELECT * FROM (SELECT ...) AS x`），跨表则没问题（官方示例和上面实测都是跨表）。

### 8.3 UpdateColumn / UpdateColumns —— 跳过钩子和时间追踪

```go
db.Model(&user).UpdateColumn("name", "hello")
// UPDATE users SET name='hello' WHERE id = 111;          ← 没有 updated_at

db.Model(&user).UpdateColumns(User{Name: "hello", Age: 18})
// UPDATE users SET name='hello', age=18 WHERE id = 111;
```

**源码**（`finisher_api.go:426-438`）：`UpdateColumn` 和 `Update` 唯一区别是多设了 `tx.Statement.SkipHooks = true`。这一个标志带来两个效果：

1. **不触发钩子**（`BeforeUpdate` 回调里有 `!db.Statement.SkipHooks` 判断）；
2. **不更新 `updated_at`**（`ConvertToAssignments` 里 `!stmt.SkipHooks && field.AutoUpdateTime > 0`）。

**实测场景 11**：

```go
db.Model(&h).Update("age", 2)
// [钩子] BeforeSave → BeforeUpdate → AfterUpdate → AfterSave
// UPDATE ... SET `age`=2, `updated_at`='...' ...

db.Model(&h).UpdateColumn("age", 3)
// （无任何钩子输出）
// UPDATE ... SET `age`=3 ...   ← 没有 updated_at
```

**适用场景**：纯字段修正、批量刷数据、不想污染 `updated_at` 和触发业务钩子。

### 8.4 Returning —— 返回被修改的行（PG / SQLite 专属）

```go
var users []User
db.Model(&users).Clauses(clause.Returning{}).Where("role = ?", "admin").
    Update("salary", gorm.Expr("salary * ?", 2))
// UPDATE users SET salary=salary*2 ... RETURNING *
```

> ⚠️ **MySQL 不支持 `RETURNING`**。你用的是 MySQL，这一节跳过即可；只有 PostgreSQL / SQLite 才有此能力。

---

## 9. 坑点速查表

| # | 坑 | 现象 | 解法 |
|---|---|---|---|
| 1 | `Save` 更新所有字段（含零值） | 只想改 name，结果 age 被清零 | 用 `Updates`（struct 跳过零值） |
| 2 | `Save` 传"新构造零值 struct" | `created_at='0000-00-00'` 报错 / 主键被写 0 | `Save` 只传"查出来的完整对象" |
| 3 | `Save` 和 `Model` 混用 | 未定义行为 | 二选一，别同时用 |
| 4 | `Select("*")` 更新所有字段 | 连 `id`/`created_at`/`deleted_at` 都写，报错 | 用 `Select("A","B")` 明确列字段 |
| 5 | `Updates(struct)` 的 `false`/`0` 零值 | 批量置 false 静默失败 | 用 `map` 或 `Select` |
| 6 | 无条件 `Update` | `ErrMissingWhereClause` | 显式 `Where`，或 `AllowGlobalUpdate` |
| 7 | 软删除模型的"无条件更新" | 有 `deleted_at IS NULL` 仍被拦 | 这是保护，不是 bug；加真实条件 |
| 8 | `RowsAffected` 被 `updated_at` 污染 | 值没变却返回 1 | 用 `Changed` 判断，或 `UpdateColumn` |
| 9 | 钩子顺序记错 | 写成 AfterSave→AfterUpdate | 正解：`BeforeSave→BeforeUpdate→UPDATE→AfterUpdate→AfterSave` |
| 10 | 钩子里用外层 `db` | 脱离事务，回滚不一致 | 用参数 `tx`（详见答疑记录 Q1） |
| 11 | `Changed` 在 `AfterUpdate` 里用 | 返回结果不可靠 | `Changed` 只在 `BeforeUpdate` 有意义 |
| 12 | 子查询更新同表 | MySQL 报 `You can't specify target table` | 跨表，或同表套派生表 |
| 13 | MySQL 用 `clause.Returning` | 报语法错 | 换 PG/SQLite，MySQL 不支持 |

---

## 10. 动手练习

配套程序在 `GORM/更新Update/main.go`（15 个场景，跑完自动删表）。建议你**自己改几个地方**，边跑边看 SQL 变化：

- [ ] 把场景 1 的 `db.Save(&u)` 换成 `db.Updates(&u)`，观察 `age=0` 是否还在 SET 里
- [ ] 给 `DemoUser` 加一个带 `default:0` tag 的字段，用 `Updates(struct)` 传零值，对比"更新"和"创建"的零值规则差异
- [ ] 在 `BeforeUpdate` 里返回一个 error，观察整条 UPDATE 是否被中止、事务是否回滚
- [ ] 用 `db.Model(&DemoUser{}).Where("1 = 1").Update(...)` 做一次全表更新，确认 `RowsAffected` 等于总行数
- [ ] 试试 `Omit("name").Updates(map[...])`，确认 SET 里确实没有 name
- [ ] 把场景 10 的子查询改成"同表自更新"（`Update("age", db.Model(&DemoUser{}).Select("age").Where(...))`），复现 MySQL 的报错，再用派生表绕过

### 速记卡

```
更新三兄弟：
  Save       → 全字段 upsert（零值也写，危险）
  Update     → 单列，必须带条件（否则 ErrMissingWhereClause）
  Updates    → 多列：struct 跳零值 / map 保留零值

字段控制：
  Select("A","B")  → 只更新指定字段（含零值）
  Omit("A")        → 除 A 外都更新
  Select("*")      → 全字段（⚠️ 连 id/created_at 都写，慎用）

跳过钩子：
  UpdateColumn / UpdateColumns  → SkipHooks=true，不触发钩子、不刷新 updated_at

钩子顺序：
  BeforeSave → BeforeUpdate → UPDATE → AfterUpdate → AfterSave
```
