package main

import (
	"errors"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

// ---- 模型 ----

type User struct {
	ID   int64
	Name string
	Age  int64
}

func (User) TableName() string { return "demo_q_users" }

// 软删除模型：带 gorm.DeletedAt，查询时自动追加 deleted_at IS NULL
type SoftRecord struct {
	ID        int64
	Name      string
	DeletedAt gorm.DeletedAt
}

func (SoftRecord) TableName() string { return "demo_q_soft" }

// 专门用来演示 AfterFind 钩子的模型
type HookUser struct {
	ID   int64
	Name string
}

func (HookUser) TableName() string { return "demo_q_hooks" }

func (h *HookUser) AfterFind(tx *gorm.DB) error {
	fmt.Printf("      [AfterFind 钩子] 读到 id=%d name=%s\n", h.ID, h.Name)
	return nil
}

// 聚合结果结构体（Group By / Having / Scan 用）
type GroupResult struct {
	Age   int64
	Total int64
}

var db *gorm.DB

func title(s string) {
	fmt.Printf("\n\n===== %s =====\n", s)
}

func main() {
	dsn := "root:bjy12345@(127.0.0.1:3306)/sql_test?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}

	keepTables := false
	if err := db.AutoMigrate(&User{}, &SoftRecord{}, &HookUser{}); err != nil {
		panic(err)
	}
	defer func() {
		if !keepTables {
			db.Migrator().DropTable(&User{}, &SoftRecord{}, &HookUser{})
		}
	}()

	// TRUNCATE 会重置自增，保证硬编码 id 每次一致
	seedUsers := func() {
		db.Exec("TRUNCATE TABLE demo_q_users")
		db.Create(&[]User{
			{Name: "alice", Age: 18},
			{Name: "bob", Age: 20},
			{Name: "charlie", Age: 20},
			{Name: "david", Age: 25},
			{Name: "eve", Age: 0}, // 零值，用来对比 struct vs map 条件
		})
	}

	// ============ 场景 1 · First / Take / Last 生成的 SQL 有何不同 ============
	title("场景 1 · First / Take / Last 生成的 SQL 有何不同（各用独立变量）")
	seedUsers()
	var f, t, l User
	db.First(&f) // ORDER BY id LIMIT 1
	db.Take(&t)  // LIMIT 1（无排序）
	db.Last(&l)  // ORDER BY id DESC LIMIT 1

	// ============ 场景 1b · 复用同一变量会带主键残留 ============
	title("场景 1b · 复用同一个目标变量会带着上次回写的主键")
	var u1b User
	db.First(&u1b) // u1b.ID 被回写为 1
	db.Take(&u1b)  // 意外带上 WHERE id = 1！
	fmt.Printf("  ↑ 第二条 Take 为什么多了 WHERE id = 1？因为 First 已把主键回写进 u1b\n")

	// ============ 场景 2 · 查不到记录：First 报错 vs Find 不报错 ============
	title("场景 2 · 查不到记录：First 报错 vs Find 不报错")
	db.Exec("TRUNCATE TABLE demo_q_users")
	var one User
	err = db.First(&one).Error
	fmt.Printf("  First 空表 → errors.Is(err, ErrRecordNotFound) = %v | err = %v\n",
		errors.Is(err, gorm.ErrRecordNotFound), err)
	var many []User
	err = db.Find(&many).Error
	fmt.Printf("  Find  空表 → err = %v | len = %d（不报错，返回空切片）\n", err, len(many))

	// ============ 场景 3 · 目标结构体带主键值 → 自动加 WHERE ============
	title("场景 3 · 目标结构体带非零主键 → 自动构建 WHERE id = ?")
	seedUsers()
	var u3 User
	u3.ID = 3
	db.First(&u3) // WHERE id = 3
	fmt.Printf("  取回：id=%d name=%s\n", u3.ID, u3.Name)

	// ============ 场景 4 · Where + 非零主键 = AND 叠加（官方提醒的坑）============
	title("场景 4 · Where + 非零主键 = AND 叠加（官方提醒的坑）")
	var u4 User
	u4.ID = 3 // id=3 这条记录真实存在（charlie）
	err = db.Where("id = ?", 20).First(&u4).Error
	fmt.Printf("  WHERE id = 20 AND id = 3 → err = %v（id=3 存在却查不到，因为叠加了 id=20）\n", err)

	// ============ 场景 5 · struct 条件跳过零值 vs map 条件包含零值 ============
	title("场景 5 · struct 条件跳过零值 vs map 条件包含零值")
	var us []User
	db.Where(&User{Age: 0}).Find(&us)
	fmt.Printf("  Where(&User{Age:0}) → 零值被跳过，等价无条件，查到 %d 条\n", len(us))
	us = us[:0]
	db.Where(map[string]any{"age": 0}).Find(&us)
	fmt.Printf("  Where(map{age:0})  → 零值被保留，WHERE age = 0，查到 %d 条\n", len(us))

	// ============ 场景 6 · 指定结构体查询字段 ============
	title("场景 6 · 指定结构体查询字段")
	us = us[:0]
	db.Where(&User{Name: "alice"}, "Age").Find(&us)
	fmt.Printf("  Where(&User{Name:alice}, \"Age\")         → 只查 Age 字段，查到 %d 条\n", len(us))
	us = us[:0]
	db.Where(&User{Name: "alice"}, "name", "Age").Find(&us)
	fmt.Printf("  Where(&User{Name:alice}, \"name\", \"Age\") → 查 name + Age 两字段，查到 %d 条\n", len(us))

	// ============ 场景 7 · Select 字段 + 扫描到小结构体（智能选择）============
	title("场景 7 · Select 字段 + 扫描到小结构体（智能选择）")
	type NameOnly struct {
		ID   int64
		Name string
	}
	var names []NameOnly
	db.Model(&User{}).Select("id", "name").Find(&names)
	fmt.Printf("  Select(id,name) → 查到 %d 条，每条只含 id/name\n", len(names))

	// ============ 场景 8 · Limit(-1) / Offset(-1) 取消限制 ============
	title("场景 8 · Limit(-1) / Offset(-1) 取消限制")
	us = us[:0]
	db.Limit(2).Find(&us)
	fmt.Printf("  Limit(2)           → 查到 %d 条\n", len(us))
	us = us[:0]
	db.Limit(2).Limit(-1).Find(&us)
	fmt.Printf("  Limit(2).Limit(-1) → 查到 %d 条（limit 被取消）\n", len(us))

	// ============ 场景 9 · Order 多字段 + FIELD() ============
	title("场景 9 · Order 多字段 + FIELD()")
	us = us[:0]
	db.Order("age desc").Order("name").Find(&us)
	us = us[:0]
	db.Clauses(clause.OrderBy{
		Expression: clause.Expr{SQL: "FIELD(id,?)", Vars: []any{[]int{3, 1, 2}}, WithoutParentheses: true},
	}).Find(&us)
	fmt.Printf("  FIELD(id,3,1,2) → 按 3,1,2 顺序返回\n")

	// ============ 场景 10 · Group By + Having + Scan 聚合 ============
	title("场景 10 · Group By + Having + Scan 聚合结果")
	var gr []GroupResult
	db.Model(&User{}).Select("age, count(*) as total").Group("age").
		Having("count(*) >= ?", 1).Order("age").Scan(&gr)
	for _, g := range gr {
		fmt.Printf("  age=%d 有 %d 人\n", g.Age, g.Total)
	}

	// ============ 场景 11 · Not 与 Or 条件 ============
	title("场景 11 · Not 与 Or 条件")
	us = us[:0]
	db.Not("name = ?", "alice").Find(&us)
	fmt.Printf("  Not(name=alice) → 查到 %d 条\n", len(us))
	us = us[:0]
	db.Where("name = ?", "alice").Or("name = ?", "bob").Find(&us)
	fmt.Printf("  name=alice OR name=bob → 查到 %d 条\n", len(us))

	// ============ 场景 12 · 内联条件 ============
	title("场景 12 · 内联条件")
	us = us[:0]
	db.Find(&us, "age > ?", 18)
	fmt.Printf("  Find(&us, \"age > ?\", 18) → 查到 %d 条\n", len(us))
	var u12 User
	db.Find(&u12, "name = ?", "alice")
	fmt.Printf("  Find(&one, \"name = ?\", \"alice\") → %s\n", u12.Name)

	// ============ 场景 13 · db.Table + Scan 到非模型 struct ============
	title("场景 13 · db.Table + Scan 到非模型 struct")
	type Row struct {
		Name string
		Age  int64
	}
	var rows []Row
	db.Table("demo_q_users").Select("name", "age").Where("age > ?", 0).Scan(&rows)
	for _, r := range rows {
		fmt.Printf("  %-8s %d\n", r.Name, r.Age)
	}

	// ============ 场景 14 · 软删除：gorm.DeletedAt 自动加 deleted_at IS NULL ============
	title("场景 14 · 软删除：gorm.DeletedAt 自动加 deleted_at IS NULL")
	db.Exec("TRUNCATE TABLE demo_q_soft")
	db.Create(&SoftRecord{Name: "keep"})
	db.Create(&SoftRecord{Name: "trash"})
	db.Delete(&SoftRecord{ID: 2}) // 软删除，只更新 deleted_at
	var softs []SoftRecord
	db.Find(&softs)
	fmt.Printf("  Find()            → 查到 %d 条（自动过滤已删）\n", len(softs))
	softs = softs[:0]
	db.Unscoped().Find(&softs)
	fmt.Printf("  Unscoped().Find() → 查到 %d 条（含已删）\n", len(softs))

	// ============ 场景 15 · AfterFind 钩子：每条记录触发一次 ============
	title("场景 15 · AfterFind 钩子：每条记录触发一次")
	db.Exec("TRUNCATE TABLE demo_q_hooks")
	db.Create(&[]HookUser{{Name: "a"}, {Name: "b"}, {Name: "c"}})
	var hs []HookUser
	db.Find(&hs)
	fmt.Println("  ↑ 注意：3 条记录触发了 3 次 AfterFind")

	fmt.Println("\n\n全部场景执行完毕。")
}
