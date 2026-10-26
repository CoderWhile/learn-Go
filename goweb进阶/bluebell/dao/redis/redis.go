package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

var rdb *redis.Client

func Init() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", viper.GetString("redis.host"),
			viper.GetInt("redis.port"),
		),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
		PoolSize: viper.GetInt("redis.pool_size"),
	})
	//给ping这次操作设置一个5秒的超时,如果5秒内连不上redis,就自动放弃并返回超时错误
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = rdb.Ping(ctx).Result()
	return err
}

func Close() {
	_ = rdb.Close()
}

//func main() {
//	if err := initClient(); err != nil {
//		fmt.Printf("init redis client err:%v\n", err)
//		return
//	}
//	fmt.Println("init redis client success")
//	//rdb.Get()
//	//程序退出时释放相关资源
//	rdb.Close()
//}
