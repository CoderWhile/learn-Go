// GORM 事务 / 会话 演示程序
//
// 目的：用真实 SQL 日志 + 连接 ID 证明：
//   钩子（以及任何回调）里如果用「外层全局 db」而不是「参数 tx」，
//   这条写操作会跑在「另一条连接」上、立即自动提交，脱离外层事务。
//   外层事务一旦回滚，它不会跟着回滚 —— 这就是「数据不一致」。
//
// 运行：go run .
// 说明：使用独立的 demo_ 前缀表，跑完自动删除，不影响业务表。
package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB

// 开关：是否让钩子故意使用外层全局 db
var hookUseGlobalDB bool

// 钩子调用次数（验证 SkipHooks 的作用范围）
var hookCalls int

// 连接 ID：用来判断「是不是同一条连接 / 同一个事务」
var hookConnID, outerConnID int64

// 跑完是否保留 demo 表
const keepTables = false

// ---------- 模型：表名自动是 demo_users / demo_counters / demo_items / demo_zeros ----------
type DemoUser struct {
	ID   uint
	Name string
}

type DemoCounter struct {
	ID    uint
	Total int64
}

type DemoItem struct {
	ID     uint
	Name   string
	Amount int64
}

type DemoZero struct {
	ID    uint
	Name  string
	Count int64
}

// 用来测 default tag 与零值的边界
type DemoDefault struct {
	ID     uint
	Name   string `gorm:"default:anon"`
	Age    int64  `gorm:"default:18"`
	AgePtr *int64 `gorm:"default:18"`
}

// ---------- 取连接 ID ----------
// Session(&gorm.Session{}) 让这次查询独立执行，不污染传入的 d
func connID(d *gorm.DB) int64 {
	var id int64
	d.Session(&gorm.Session{}).Raw("SELECT CONNECTION_ID()").Scan(&id)
	return id
}

// ---------- 钩子 ----------

// DemoUser 的 BeforeCreate：往计数器 +1
//
//	✅ target = tx → 同一连接、同一事务
//	❌ target = db → 另一条连接，语句级自动提交
func (u *DemoUser) BeforeCreate(tx *gorm.DB) error {
	hookCalls++
	target := tx
	if hookUseGlobalDB {
		target = db
	}
	hookConnID = connID(target)
	return target.Model(&DemoCounter{}).Where("id = ?", 1).
		UpdateColumn("total", gorm.Expr("total + 1")).Error
}

// DemoItem 的 AfterCreate：把本条元素的金额写进它对应的计数器
// 注意条件用的是「元素自身的 ID」——如果 tx 上的条件会跨元素累积，这里必然出错
func (it *DemoItem) AfterCreate(tx *gorm.DB) error {
	return tx.Model(&DemoCounter{}).Where("id = ?", it.ID).
		UpdateColumn("total", it.Amount).Error
}

// ---------- 小工具 ----------

// 静默会话：读数据时不打 SQL 日志，顺便演示 Session 的 Logger 也是「会话级」
func quiet() *gorm.DB {
	return db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)})
}

func readCounter() int64 {
	var c DemoCounter
	quiet().First(&c, 1)
	return c.Total
}

func readCounterByID(id uint) int64 {
	var c DemoCounter
	quiet().First(&c, id)
	return c.Total
}

func countUsers() int64 {
	var n int64
	quiet().Model(&DemoUser{}).Count(&n)
	return n
}

func reset() {
	for _, m := range []any{&DemoUser{}, &DemoItem{}, &DemoCounter{}, &DemoZero{}, &DemoDefault{}} {
		quiet().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(m)
	}
}

func seedCounters(ids ...uint) {
	for _, id := range ids {
		quiet().Create(&DemoCounter{ID: id})
	}
}

func title(s string) {
	fmt.Printf("\n\n%s\n%s\n%s\n", strings.Repeat("=", 72), s, strings.Repeat("=", 72))
}

func ok(label string, good bool, bad string) {
	if good {
		fmt.Printf("   ✅ %s\n", label)
	} else {
		fmt.Printf("   ❌ %s\n", bad)
	}
}

func main() {
	dsn := "root:bjy12345@(127.0.0.1:3306)/sql_test?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	if err := db.AutoMigrate(&DemoUser{}, &DemoCounter{}, &DemoItem{}, &DemoZero{}, &DemoDefault{}); err != nil {
		log.Fatal(err)
	}
	defer func() {
		if !keepTables {
			db.Migrator().DropTable(&DemoUser{}, &DemoCounter{}, &DemoItem{}, &DemoZero{}, &DemoDefault{})
			fmt.Println("\n（demo 表已删除；想保留请把 keepTables 改成 true）")
		}
		if sqlDB, e := db.DB(); e == nil {
			sqlDB.Close()
		}
	}()

	// ==========================================================
	title("场景 1 · 钩子里用 tx（正确写法）")
	// ==========================================================
	reset()
	seedCounters(1)
	hookUseGlobalDB = false

	fmt.Println("\n-- 事务开始前：计数器 = 0，用户数 = 0")
	fmt.Println("-- 事务内容：① Create 用户（钩子给计数器 +1） ② 返回 error 触发回滚")
	fmt.Printf("-- 事务外的连接 ID = %d\n", connID(db))

	err = db.Transaction(func(tx *gorm.DB) error {
		outerConnID = connID(tx)
		fmt.Printf("-- 事务内的连接 ID = %d\n", outerConnID)
		if e := tx.Create(&DemoUser{Name: "订单A"}).Error; e != nil {
			return e
		}
		return errors.New("后续步骤失败，整个事务回滚")
	})
	fmt.Printf("\n-- 事务返回：%v\n", err)
	fmt.Printf("-- 钩子里的连接 ID = %d\n", hookConnID)
	ok(fmt.Sprintf("钩子与事务是同一条连接（%d == %d）→ 同一事务", hookConnID, outerConnID),
		hookConnID == outerConnID, "钩子和事务不在同一条连接上")
	fmt.Printf("-- 事务结束后：计数器 = %d，用户数 = %d\n", readCounter(), countUsers())
	ok("用户被回滚了（=0）", countUsers() == 0, "用户没回滚")
	ok("计数器被回滚了（=0）", readCounter() == 0, "计数器没回滚")
	fmt.Println("   👉 两处修改在同一事务里，一起回滚，数据一致。")

	// ==========================================================
	title("场景 2 · 钩子里用外层 db（错误写法 = 你看到的那句话）")
	// ==========================================================
	reset()
	seedCounters(1)
	hookUseGlobalDB = true

	fmt.Println("\n-- 事务开始前：计数器 = 0，用户数 = 0")
	fmt.Println("-- 事务内容与场景 1 完全相同，只有钩子里的 target 换成了全局 db")
	fmt.Printf("-- 事务外的连接 ID = %d\n", connID(db))

	err = db.Transaction(func(tx *gorm.DB) error {
		outerConnID = connID(tx)
		fmt.Printf("-- 事务内的连接 ID = %d\n", outerConnID)
		if e := tx.Create(&DemoUser{Name: "订单A"}).Error; e != nil {
			return e
		}
		return errors.New("后续步骤失败，整个事务回滚")
	})
	fmt.Printf("\n-- 事务返回：%v\n", err)
	fmt.Printf("-- 钩子里的连接 ID = %d\n", hookConnID)
	ok(fmt.Sprintf("钩子与事务是同一条连接（%d == %d）", hookConnID, outerConnID),
		hookConnID == outerConnID,
		fmt.Sprintf("钩子用了另一条连接（%d != %d）→ 已经脱离事务", hookConnID, outerConnID))
	fmt.Printf("-- 事务结束后：计数器 = %d，用户数 = %d\n", readCounter(), countUsers())
	ok("用户被回滚了（=0）", countUsers() == 0, "用户没回滚")
	ok("计数器被回滚了（=0）", readCounter() == 0,
		fmt.Sprintf("计数器没回滚（= %d）—— 这就是「数据不一致」", readCounter()))
	hookUseGlobalDB = false

	// ==========================================================
	title("场景 3 · 「会话级」到底影响谁")
	// ==========================================================
	reset()
	seedCounters(1)

	fmt.Println("\n-- 3.1 SkipHooks：只影响这一条链，不影响 db 本身")
	hookCalls = 0
	db.Session(&gorm.Session{SkipHooks: true}).Create(&DemoUser{Name: "跳过钩子"})
	fmt.Printf("   (a) Session{SkipHooks:true}.Create → 钩子调用次数 = %d（期望 0）\n", hookCalls)
	db.Create(&DemoUser{Name: "正常钩子"})
	fmt.Printf("   (b) 紧接着 db.Create            → 钩子调用次数 = %d（期望 1）\n", hookCalls)

	fmt.Println("\n-- 3.2 不安全复用：一次使用会「永久污染」这个对象（官方 Bad case）")
	reset()
	db.Session(&gorm.Session{SkipHooks: true}).Create(&[]DemoUser{{Name: "会话A"}, {Name: "会话B"}})

	stmt := db.Where("name = ?", "会话A")
	var n1, n2, n3 int64
	stmt.Model(&DemoUser{}).Where("name = ?", "会话B").Count(&n1) // 会话A AND 会话B
	stmt.Model(&DemoUser{}).Count(&n2)                            // 本该只剩会话A，实际仍带着会话B
	quiet().Model(&DemoUser{}).Count(&n3)                         // 对照：全新链
	fmt.Printf("   第一次(会话A AND 会话B) = %d（期望 0）\n", n1)
	fmt.Printf("   第二次(本以为只剩会话A) = %d（期望 1）\n", n2)
	ok("复用后仍能查到会话A", n2 == 1, "复用后查不到了 —— stmt 被上一次的条件污染了")
	fmt.Printf("   对照：一条全新链      = %d\n", n3)

	fmt.Println("\n-- 3.3 安全复用：Session(&gorm.Session{}) 让每次派生都独立（官方 Good case）")
	base := db.Where("name = ?", "会话A").Session(&gorm.Session{})
	var f1, f2, f3 int64
	base.Model(&DemoUser{}).Where("name = ?", "会话B").Count(&f1) // 会话A AND 会话B
	base.Model(&DemoUser{}).Count(&f2)                            // 应该只剩会话A
	base.Model(&DemoUser{}).Count(&f3)                            // 再查一次，仍然只剩会话A
	fmt.Printf("   第一次(会话A AND 会话B) = %d（期望 0）\n", f1)
	fmt.Printf("   第二次(只剩会话A)       = %d（期望 1）\n", f2)
	fmt.Printf("   第三次(只剩会话A)       = %d（期望 1）\n", f3)
	ok("base 没被后续调用污染", f2 == 1 && f3 == 1, "base 被污染了")

	fmt.Println("\n-- 3.4 Session{NewDB: true} → 丢弃之前累积的条件（官方给钩子的 tx 就是这么造的）")
	dirty := db.Where("name = ?", "会话A")
	var g1, g2 int64
	dirty.Model(&DemoUser{}).Count(&g1)
	dirty.Session(&gorm.Session{NewDB: true}).Model(&DemoUser{}).Count(&g2)
	fmt.Printf("   原链          = %d（带 name='会话A'）\n", g1)
	fmt.Printf("   NewDB 之后    = %d（条件被清空，全表 2 条）\n", g2)

	// ==========================================================
	title("场景 4 · 嵌套 Transaction 会变成 SAVEPOINT")
	// ==========================================================
	reset()
	seedCounters(1)
	fmt.Println("\n-- 内层返回 nil（看起来提交了），外层返回 error，看内层是否也回滚")

	err = db.Transaction(func(tx *gorm.DB) error {
		tx.Create(&DemoUser{Name: "外层"})
		inner := tx.Transaction(func(tx2 *gorm.DB) error { // 注意：起点必须是 tx，不是 db
			return tx2.Create(&DemoUser{Name: "内层"}).Error
		})
		fmt.Printf("   -- 内层事务返回：%v\n", inner)
		fmt.Println("   -- 外层事务返回 error，触发整体回滚")
		return errors.New("外层失败")
	})
	fmt.Printf("-- 最终用户数 = %d\n", countUsers())
	ok("内外层一起回滚了（=0）", countUsers() == 0, "内层没跟着回滚")

	// ==========================================================
	title("场景 5 · 批量插入时，钩子里的 tx 会不会跨元素累积条件？")
	// ==========================================================
	reset()
	seedCounters(1, 2, 3)
	fmt.Println("\n-- 批量插入 3 条，每条在 AfterCreate 里更新「自己那个」计数器")
	fmt.Println("-- 如果 tx 的条件会累积，第 2 条的 WHERE 就会变成 id=1 AND id=2 → 更新 0 行")

	db.Create(&[]DemoItem{
		{Name: "a", Amount: 10},
		{Name: "b", Amount: 20},
		{Name: "c", Amount: 30},
	})
	fmt.Println()
	for _, id := range []uint{1, 2, 3} {
		fmt.Printf("   counter %d = %d\n", id, readCounterByID(id))
	}
	allOK := readCounterByID(1) == 10 && readCounterByID(2) == 20 && readCounterByID(3) == 30
	ok("三条都各更新了自己的计数器 → 没有累积", allOK, "有元素的条件被累积，更新丢失了")

	// ==========================================================
	title("场景 6 · struct 的零值字段，Create 时会不会被写进 SQL？")
	// ==========================================================
	reset()
	fmt.Println("\n-- db.Create(&DemoZero{Name: \"z\"})：Count 未赋值（零值 0）")
	db.Create(&DemoZero{Name: "z"})
	var z DemoZero
	quiet().First(&z)
	fmt.Printf("   读回来 Count = %d\n", z.Count)
	fmt.Println("   👆 看上面那条 INSERT 的列清单里有没有 `count`")

	// ==========================================================
	title("场景 7 · 零值到底什么时候被跳过？（default tag 的边界）")
	// ==========================================================
	reset()
	fmt.Println("\n-- (a) 无 tag 的 int64 + 零值：db.Create(&DemoZero{Name: \"a\"})")
	db.Create(&DemoZero{Name: "a"})

	fmt.Println("\n-- (b) 带 default tag 的字段全部留零值：db.Create(&DemoDefault{})")
	b := DemoDefault{}
	db.Create(&b)
	ptrB := "nil"
	if b.AgePtr != nil {
		ptrB = fmt.Sprint(*b.AgePtr)
	}
	fmt.Printf("   Create 之后，内存里的结构体被改成了：name=%q age=%d age_ptr=%s\n", b.Name, b.Age, ptrB)

	fmt.Println("\n-- (c) 带 default tag，但显式传 0；指针指向 0")
	zero := int64(0)
	db.Create(&DemoDefault{Name: "", Age: 0, AgePtr: &zero})

	fmt.Println("\n-- (d) 用 Select 强制包含默认值字段：db.Select(\"Name\",\"Age\").Create(&DemoDefault{})")
	db.Select("Name", "Age").Create(&DemoDefault{})

	fmt.Println("\n-- 读回来看真实存了什么：")
	var ds []DemoDefault
	quiet().Order("id").Find(&ds)
	for _, d := range ds {
		ptr := "nil"
		if d.AgePtr != nil {
			ptr = fmt.Sprint(*d.AgePtr)
		}
		fmt.Printf("   id=%d name=%-8q age=%-3d age_ptr=%s\n", d.ID, d.Name, d.Age, ptr)
	}

	// ==========================================================
	title("场景 8 · 用 map 创建时，default tag 还会生效吗？")
	// ==========================================================
	reset()
	fmt.Println("\n-- db.Model(&DemoDefault{}).Create(map[string]any{\"Name\": \"\", \"Age\": 0})")
	db.Model(&DemoDefault{}).Create(map[string]any{"Name": "", "Age": 0})

	fmt.Println("\n-- 对照：struct 同样传零值")
	db.Create(&DemoDefault{Name: "", Age: 0})

	fmt.Println("\n\n全部场景执行完毕。")
}
