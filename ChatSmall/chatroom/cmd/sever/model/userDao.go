package model

import (
	"ChatSmall/chatroom/cmd/common/message"
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
)

// 在服务器启动后，就初始化一个userDao实例
// 做成全局变量，需要时直接使用

var MyUserDao *UserDao

// 定义一个UserDao结构体
// 完成对User结构体的各种操作
type UserDao struct {
	//连接池当作字段
	Db *sql.DB
}

// 使用工厂模式，创建一个UserDao实例
func NewUserDao(db *sql.DB) (userDao *UserDao) {
	userDao = &UserDao{
		Db: db,
	}
	return
}

func (this UserDao) intiDB() (err error) {
	// 构建连接字符串 (DSN)
	// 格式: 用户名:密码@tcp(地址:端口)/数据库名
	// 如果你没改过配置，通常地址是 127.0.0.1，端口是 3306
	// 这里的 'mysql' 是系统自带的默认库，用来测试连接最方便
	dsn := "root:bjy12345@tcp(127.0.0.1:3306)/chatroom"

	fmt.Println("正在尝试连接数据库...")

	// 打开数据库连接
	//不能写冒号，会导致空指针引用
	this.Db, err = sql.Open("mysql", dsn) //不会校验用户名和密码是否正确
	if err != nil {                       //dsn格式不正确报错
		return err
	}

	// 确保程序结束时关闭连接
	//defer db.Close()
	//尝试与数据库建立连接
	// Ping 一下数据库，验证连接是否真的通了
	err = this.Db.Ping()
	if err != nil {
		return err
	}
	//设置数据库连接池的最大连接数
	this.Db.SetMaxOpenConns(10)
	//最大空闲连接数
	this.Db.SetMaxIdleConns(5)
	return nil
}

// 应该提供哪些方法？
// 1根据用户id返回一个User实例
func (this *UserDao) getUserById(id int) (user *User, err error) {
	//通过给定的Id去mysql查询用户

	user = &User{}
	err = this.Db.QueryRow("select id,username,password from user where id=?", id).Scan(&user.UserId, &user.UserName, &user.UserPwd)
	if err == sql.ErrNoRows {
		err = ERROR_USER_NOTEXISTS
		return nil, err
	}
	if err != nil {
		fmt.Println("查询用户失败：%+v\n", err)
		return nil, fmt.Errorf("查询用户失败:%w", err)
	}
	return user, nil
}

func (this *UserDao) getUserByName(username string) (user *User, err error) {
	user = &User{}
	err = this.Db.QueryRow("select id,username,password from user where username=?", username).Scan(&user.UserId, &user.UserName, &user.UserPwd)
	if err == sql.ErrNoRows {
		err = ERROR_USER_NOTEXISTS
		return nil, err
	}
	if err != nil {
		fmt.Println("查询用户失败%+v\n", err)
		return nil, fmt.Errorf("查询用户失败：%w", err)
	}
	return user, nil
}

// 完成一个登录的校验
// 1.完成对用户的验证，
// 2.如果id和密码都正确，返回User实例
// 3.有错误，返回对应错误信息
func (this *UserDao) Login(userName string, userPwd string) (user *User, err error) {

	//defer this.db.Close()
	user, err = this.getUserByName(userName)
	if err != nil {
		return
	}
	fmt.Println("从数据库中读取的", user.UserId, userPwd)
	if user.UserPwd != userPwd {
		err = ERROR_USER_PWD
		return
	}
	return
}

// 完成对用户注册的校验
func (this *UserDao) Register(user *message.User) (err error) {
	//defer this.db.Close()
	readuser := &message.User{}
	err = this.Db.QueryRow("select username from user where username=?", user.UserName).Scan(&readuser.UserName)
	if err == nil && readuser.UserName == user.UserName {
		err = ERROR_USER_EXISTS
		return
	}
	//查不到数据,这个用户名还没有注册
	//完成注册
	//序列化
	//data, err := json.Marshal(user)
	//if err != nil {
	//	return
	//}
	//入库
	//插入操作
	sqlStr := `insert into user(username,password) values(?,?)`
	ret, err := this.Db.Exec(sqlStr, user.UserName, user.UserPwd)
	if err != nil {
		fmt.Println("newuser insert err:", err)
		return
	}
	id, err := ret.LastInsertId()
	if err != nil {
		fmt.Printf("get last insert id err:%v\n", err)
	}
	fmt.Println("所注册账号的id为：", id)
	return
}

// 密码修改
func (this *UserDao) ModifyPwd(newPwd string, user *message.User) (err error, newUser *message.User) {
	readuser := &message.User{}
	err = this.Db.QueryRow("select username from user where username=?", user.UserName).Scan(&readuser.UserName)
	//密码错误
	if err != nil {
		fmt.Println("密码修改查询出错")
		return
	}
	if readuser.UserPwd != user.UserPwd {
		err = ERROR_USER_PWD
		return
	}
	sqlStr := `update user set password=? where username=?`
	_, err = this.Db.Exec(sqlStr, newPwd, user.UserName)
	if err != nil {
		fmt.Println("newuser insert err:", err)
		return
	}
	newUser = readuser
	newUser.UserPwd = newPwd
	fmt.Printf("用户%s的密码修改成功", newUser.UserName)
	return

}

// 修改用户名
func (this *UserDao) UpdateUserName(userId int, newUserName, password string) (err error) {
	//1.验证用户身份
	user, err := this.getUserById(userId)
	if err != nil {
		return err
	}
	if user.UserPwd != password {
		err = ERROR_USER_PWD
		return
	}

	//2.检查新用户名是否已被占用
	_, err = this.getUserByName(newUserName)
	if err == nil {
		//查到了，用户名已存在
		err = ERROR_USER_EXISTS
		return
	}
	if err != ERROR_USER_NOTEXISTS {
		//其他数据库错误
		return
	}

	//3.更新用户名
	_, err = this.Db.Exec("update user set username=? where id=?", newUserName, userId)
	if err != nil {
		fmt.Println("update username err:", err)
		return
	}
	fmt.Printf("用户%d修改用户名为:%s\n", userId, newUserName)
	return
}
