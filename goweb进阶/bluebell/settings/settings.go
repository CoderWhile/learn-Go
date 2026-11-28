package settings

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// 全局变量,用来保存程序所有配置信息
var Conf = new(AppConfig)

type AppConfig struct {
	Name         string `mapstructure:"name"`
	Mode         string `mapstructure:"mode"`
	Version      string `mapstructure:"version"`
	Port         int    `mapstructure:"port"`
	StartTime    string `mapstructure:"start_time"`
	MachineID    int64  `mapstructure:"machine_id"`
	*LogConfig   `mapstructure:"log"`
	*MySQLConfig `mapstructure:"mysql"`
	*RedisConfig `mapstructure:"redis"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	Password     string `mapstructure:"password"`
	Port         string `mapstructure:"port"`
	User         string `mapstructure:"user"`
	DBName       string `mapstructure:"dbname"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Password     string `mapstructure:"password"`
	Port         string `mapstructure:"port"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

//func Init() (err error) {
//	//方式一:直接指定配置文件路径(相对路径或者绝对路径)
//	//相对路径:相对执行的可执行文件的相对路径
//	//绝对路径:系统中实际的文件路径
//	//viper.SetConfigFile("./config.yaml")
//
//	//方式二:指定配置文件名和配置文件的位置,viper自行查找可用的配置文件
//	//配置文件名不需要带后缀
//	//配置文件位置可配置多个
//	//viper.SetConfigName("config")//指定配置文件名(不带后缀)
//	//viper.AddConfigPath(".")//指定查找配置文件的路径(这里使用相对路径)
//	//viper.AddConfigPath("./conf")
//
//	//配合远程配置中心使用.告诉viper当前的数据使用什么格式去解析
//	//viper.SetConfigType("json")
//
//	viper.SetConfigName("config")
//	viper.SetConfigType("yaml")
//	viper.AddConfigPath("./")
//	//	viper.SetConfigFile("./conf/config.yaml") // 指定配置文件路径
//	err = viper.ReadInConfig() // 读取配置信息
//	if err != nil {            // 读取配置信息失败
//		fmt.Printf("viper.ReadInConfig() failed with %s\n", err)
//		panic(fmt.Errorf("Fatal error config file: %s \n", err))
//		return
//	}
//
//	//反序列化到结构体变量里
//	if err := viper.Unmarshal(&Conf); err != nil {
//		fmt.Printf("viper.Unmarshal() failed with %s\n", err)
//	}
//
//	// 监控配置文件变化
//	viper.WatchConfig()
//	viper.OnConfigChange(func(e fsnotify.Event) {
//		fmt.Println("配置文件有修改..")
//		if err := viper.Unmarshal(&Conf); err != nil {
//			fmt.Printf("viper.Unmarshal() failed with %s\n", err)
//		}
//	})
//	//
//	//r := gin.Default()
//	//// 访问/version的返回值会随配置文件的变化而变化
//	//r.GET("/version", func(c *gin.Context) {
//	//	c.String(http.StatusOK, viper.GetString("version"))
//	//})
//	//
//	//if err := r.Run(
//	//	fmt.Sprintf(":%d", viper.GetInt("port"))); err != nil {
//	//	panic(err)
//	//}
//	return
//}

func Init() (err error) {
	//方法1:读取指定的配置文件
	//viper.SetConfigFile("config.yaml") //指定读这个文件

	//方法2:在目录下搜索指定名字配置文件
	viper.AddConfigPath("./")     //查找路径(相对路径)
	viper.SetConfigName("config") //读取配置文件名
	viper.SetConfigType("yaml")   //配置文件类型,用来远程etcd获取配置信息格式

	//配置读取并加载
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			//文件未找到错误提示
			fmt.Println("viper.ReadInConfig failed. err :", err)
		} else {
			//文件找到,加载错误提示
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
	}

	//配置文件信息赋值给Conf结构变量中使用
	err = viper.Unmarshal(Conf)
	if err != nil {
		fmt.Println("load config with struct fail, err :", err)
		return
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("配置文件修改了")
		err := viper.Unmarshal(Conf)
		if err != nil {
			fmt.Println("热加载配置文件时出错: ", err)
			return
		}
		fmt.Println("新配置文件加载成功", Conf)
	})
	return
}
