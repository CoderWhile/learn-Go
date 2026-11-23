package redis

import (
	"bluebell/settings"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
	Nil    = redis.Nil
)

// rdb *redis.Client

func Init(cfg *settings.RedisConfig) (err error) {
	client = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Host,
			cfg.Port,
		),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})
	//给ping这次操作设置一个5秒的超时,如果5秒内连不上redis,就自动放弃并返回超时错误
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = client.Ping(ctx).Result()
	return err
}

func Close() {
	_ = client.Close()
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
