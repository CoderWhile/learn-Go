package main

import (
	"errors"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ---- 模型 ----

// 主模型：带 gorm.Model（含 CreatedAt/UpdatedAt/DeletedAt），用于验证
// ① Save 全字段更新 ② Updates 非零值 ③ UpdatedAt 自动追踪
type DemoUser struct {
	gorm.Model
	Name   string
	Age    int64
	Role   string
	Active bool
}

func (DemoUser) TableName() string { return "demo_u_users" }

// 跨表子查询的辅助表
type DemoMeta struct {
	ID   int64
	Name string
	Age  int64
}

func (DemoMeta) TableName() string { return "demo_u_metas" }

// 专门验证钩子调用顺序的模型
type DemoHook struct {
	gorm.Model
	Name string
	Age  int64
}

func (DemoHook) TableName() string { return "demo_u_hooks" }

func (u *DemoHook) BeforeSave(tx *gorm.DB) error   { fmt.Println("   [钩子] BeforeSave"); return nil }
func (u *DemoHook) BeforeUpdate(tx *gorm.DB) error { fmt.Println("   [钩子] BeforeUpdate"); return nil }
func (u *DemoHook) AfterUpdate(tx *gorm.DB) error  { fmt.Println("   [钩子] AfterUpdate"); return nil }
func (u *DemoHook) AfterSave(tx *gorm.DB) error    { fmt.Println("   [钩子] AfterSave"); return nil }

// 专门验证 Changed / SetColumn 的模型
type DemoTrack struct {
	gorm.Model
	Name string
	Age  int64
	Code string
}

func (DemoTrack) TableName() string { return "demo_u_tracks" }

func (u *DemoTrack) BeforeUpdate(tx *gorm.DB) error {
	if tx.Statement.Changed("Name") {
		fmt.Println("   [BeforeUpdate] Name 发生了变化")
		tx.Statement.SetColumn("Age", 18) // 顺带把 Age 强制改成 18
	}
	if tx.Statement.Changed("Code") {
		return errors.New("Code 字段不允许修改")
	}
	return nil
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
	if err := db.AutoMigrate(&DemoUser{}, &DemoMeta{}, &DemoHook{}, &DemoTrack{}); err != nil {
		panic(err)
	}
	defer func() {
		if !keepTables {
			db.Migrator().DropTable(&DemoUser{}, &DemoMeta{}, &DemoHook{}, &DemoTrack{})
		}
	}()

	// TRUNCATE 重置自增，保证每次 seed 后 id 稳定为 1,2,3（alice=1 bob=2 charlie=3）
	seedUsers := func() {
		db.Exec("TRUNCATE TABLE demo_u_users")
		db.Exec("TRUNCATE TABLE demo_u_metas")
		db.Create(&[]DemoUser{
			{Name: "alice", Age: 18, Role: "user", Active: true},
			{Name: "bob", Age: 20, Role: "user", Active: true},
			{Name: "charlie", Age: 25, Role: "admin", Active: false},
		})
		db.Create(&[]DemoMeta{
			{Name: "alice", Age: 99},
			{Name: "bob", Age: 88},
		})
	}

	// ============ 场景 1 · Save 有主键 = 更新所有字段（零值也写）============
	title("场景 1 · Save 有主键 = 更新所有字段（零值也写进去）")
	seedUsers()
	var u DemoUser
	db.First(&u) // 取 id=1 的 alice
	u.Name = "alice_v2"
	u.Age = 0    // 故意改成零值
	u.Role = ""  // 故意改成零值
	u.Active = false
	db.Save(&u)
	fmt.Println("  ↑ 看 UPDATE 的 SET：age=0、role=''、active=false 都被写进去了（Save=Select(*)）")

	// ============ 场景 2 · Save 无主键 = Create ============
	title("场景 2 · Save 无主键 = Create")
	db.Save(&DemoUser{Name: "newcomer", Age: 0})
	fmt.Println("  ↑ 无主键 → 走 INSERT（不是 UPDATE）")

	// ============ 场景 3 · Save 主键不存在 = 先 Update(0 行) 再回退 Create ============
	title("场景 3 · Save 主键不存在 = 先 Update(0 行) 再回退 Create")
	db.Save(&DemoUser{Model: gorm.Model{ID: 99999}, Name: "ghost", Age: 30})
	fmt.Println("  ↑ 一条 UPDATE（rows:0）+ 紧跟一条 INSERT ... ON DUPLICATE KEY UPDATE，这就是 Save 的 upsert 回退")

	// ============ 场景 4 · Update 单列 + Model 主键自动当条件 ============
	title("场景 4 · Update 单列 + Model 带主键")
	var u4 DemoUser
	db.Where("name = ?", "bob").First(&u4)
	db.Model(&u4).Update("role", "vip")
	fmt.Println("  ↑ WHERE id = ?（Model 的主键自动成为条件）")

	// ============ 场景 5 · 无条件 Update → ErrMissingWhereClause ============
	title("场景 5 · 无条件更新会被拦截（ErrMissingWhereClause）")
	err = db.Model(&DemoUser{}).Update("role", "hacker").Error
	fmt.Printf("  ❌ 无条件更新 → err = %v（errors.Is = %v）\n", err, errors.Is(err, gorm.ErrMissingWhereClause))
	db.Model(&DemoUser{}).Where("1 = 1").Update("role", "hacker")
	fmt.Println("  ✅ Where(\"1 = 1\") → 显式条件，放行")
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Model(&DemoUser{}).Update("role", "hacker")
	fmt.Println("  ✅ AllowGlobalUpdate → 会话级关闭保护，放行")

	// ============ 场景 6 · Updates(struct) 只更新非零值 ============
	title("场景 6 · Updates(struct) 只更新非零值（false/0 被跳过）")
	seedUsers()
	db.Model(&DemoUser{}).Where("id = ?", 1).
		Updates(DemoUser{Name: "alice_v3", Age: 0, Role: "", Active: false})
	fmt.Println("  ↑ 看 SET：只有 name 和 updated_at，age/role/active 的零值被跳过")

	// ============ 场景 7 · Updates(map) 会更新零值 ============
	title("场景 7 · Updates(map) 会更新零值")
	db.Model(&DemoUser{}).Where("id = ?", 2).
		Updates(map[string]any{"age": 0, "active": false})
	fmt.Println("  ↑ 看 SET：包含 age=0, active=false（map 不跳过零值）")

	// ============ 场景 8 · Select 指定字段 vs Select(*) 的坑 ============
	title("场景 8 · Select(\"Age\") 指定字段  vs  Select(\"*\") 全字段")
	seedUsers()
	db.Model(&DemoUser{}).Where("id = ?", 1).
		Select("Age").Updates(DemoUser{Name: "ignored", Age: 0})
	fmt.Println("  ↑ Select(\"Age\")：只更新 age=0，Name 被忽略（安全）")
	fmt.Println("\n  -- 下面是 Select(\"*\")：连主键 id、created_at 都写，created_at 变 '0000-00-00' 报错 --")
	db.Model(&DemoUser{}).Where("id = ?", 1).
		Select("*").Updates(DemoUser{Name: "alice_v4", Age: 0, Role: "", Active: false})
	fmt.Println("  ↑ Select(\"*\") 对「新构造的零值 struct」是危险的：id=0、created_at='0000-00-00' 都进 SQL 了")

	// ============ 场景 9 · SQL 表达式更新（原子自增）============
	title("场景 9 · SQL 表达式更新（原子自增）")
	seedUsers()
	db.Model(&DemoUser{}).Where("id = ?", 1).
		Update("age", gorm.Expr("age + ?", 10))
	db.Model(&DemoUser{}).Where("age > ?", 100).
		UpdateColumn("role", gorm.Expr("CONCAT(role, '_vip')"))
	fmt.Println("  ↑ 第一个 age = age + 10；第二个 role = CONCAT(...)（带条件）")

	// ============ 场景 10 · 子查询更新（跨表）============
	title("场景 10 · 子查询更新（跨表）")
	seedUsers()
	db.Model(&DemoUser{}).Where("name = ?", "alice").
		Update("age", db.Model(&DemoMeta{}).Select("age").
			Where("demo_u_metas.name = demo_u_users.name"))
	var u10 DemoUser
	db.Where("name = ?", "alice").First(&u10)
	fmt.Printf("  alice 的 age 被更新为 meta 里的 %d\n", u10.Age)

	// ============ 场景 11 · UpdateColumn 跳过钩子 + 不更新 updated_at ============
	title("场景 11 · UpdateColumn 跳过钩子 + 不更新 updated_at")
	db.Exec("TRUNCATE TABLE demo_u_hooks")
	db.Create(&DemoHook{Name: "a", Age: 1})
	var h DemoHook
	db.Where("name = ?", "a").First(&h)
	fmt.Println("\n  -- 用 Update（触发钩子 + 更新 updated_at）--")
	db.Model(&h).Update("age", 2)
	fmt.Println("\n  -- 用 UpdateColumn（跳过钩子 + 不动 updated_at）--")
	db.Model(&h).UpdateColumn("age", 3)
	fmt.Println("  ↑ 对比：Update 打印 4 行钩子，UpdateColumn 一行都没有")

	// ============ 场景 12 · 钩子调用顺序 ============
	title("场景 12 · 更新钩子的真实顺序")
	db.Exec("TRUNCATE TABLE demo_u_hooks")
	db.Create(&DemoHook{Name: "b", Age: 1})
	var h2 DemoHook
	db.Where("name = ?", "b").First(&h2)
	db.Model(&h2).Update("age", 9)
	fmt.Println("  ↑ 顺序：BeforeSave → BeforeUpdate → (UPDATE) → AfterUpdate → AfterSave")

	// ============ 场景 13 · Changed / SetColumn ============
	title("场景 13 · Changed 判断字段是否变更 + SetColumn 改值")
	db.Exec("TRUNCATE TABLE demo_u_tracks")
	db.Create(&DemoTrack{Name: "t1", Age: 1, Code: "c1"})
	var tr DemoTrack
	db.Where("name = ?", "t1").First(&tr)
	fmt.Println("\n  -- Name 变了：打印 'Name 发生了变化'，且 Age 被 SetColumn 强制改成 18 --")
	db.Model(&tr).Updates(map[string]any{"name": "t1_v2"})
	fmt.Println("\n  -- Name 没变（值相同）：Changed(\"Name\")=false，什么都不触发 --")
	db.Model(&DemoTrack{Model: gorm.Model{ID: tr.ID}, Name: "t1_v2"}).
		Updates(map[string]any{"name": "t1_v2"})
	fmt.Println("\n  -- 改 Code：Changed(\"Code\") 触发拦截，返回 error --")
	err = db.Model(&DemoTrack{Model: gorm.Model{ID: tr.ID}, Name: "t1_v2"}).
		Updates(map[string]any{"code": "c2"}).Error
	fmt.Printf("  Code 变更被拦截 → err = %v\n", err)

	// ============ 场景 14 · 批量更新的 struct false 陷阱 ============
	title("场景 14 · 批量更新：struct 的 false 是零值会被跳过")
	seedUsers()
	r1 := db.Model(&DemoUser{}).Where("role = ?", "user").
		Updates(DemoUser{Active: false})
	fmt.Printf("  ❌ struct 传 Active:false → SET 里没有 active，只更新了 updated_at，RowsAffected=%d\n", r1.RowsAffected)
	r2 := db.Model(&DemoUser{}).Where("role = ?", "user").
		Updates(map[string]any{"active": false})
	fmt.Printf("  ✅ map 传 active:false → SET 里有 active=false，RowsAffected=%d\n", r2.RowsAffected)

	// ============ 场景 15 · RowsAffected 语义 ============
	title("场景 15 · RowsAffected 语义：Update 带 updated_at vs UpdateColumn 不带")
	seedUsers()
	rA := db.Model(&DemoUser{}).Where("id = ?", 1).UpdateColumn("active", true)
	fmt.Printf("  UpdateColumn 把 active 设成 true（值本来就没变）→ RowsAffected=%d（MySQL 不计没变的行）\n", rA.RowsAffected)
	rB := db.Model(&DemoUser{}).Where("id = ?", 1).Update("active", true)
	fmt.Printf("  Update 同样把 active 设成 true → RowsAffected=%d（因为 updated_at 变了，也算 1 行）\n", rB.RowsAffected)

	fmt.Println("\n\n全部场景执行完毕。")
}
