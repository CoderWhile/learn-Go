package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

func initClient() (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		PoolSize: 20,
	})
	//给ping这次操作设置一个5秒的超时,如果5秒内连不上redis,就自动放弃并返回超时错误
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = rdb.Ping(ctx).Result()
	return err
}

func redisExample() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = rdb.Set(ctx, "score", 100, 0).Err()
	if err != nil {
		fmt.Printf("set score failed, err:%v\n", err)
		return
	}

	val, err := rdb.Get(ctx, "score").Result()
	if err != nil {
		fmt.Printf("get score failed, err:%v\n", err)
		return
	}
	fmt.Printf("score:%v\n", val)

	val2, err := rdb.Get(ctx, "name").Result()
	if err == redis.Nil {
		fmt.Println("name does not exist")
	} else if err != nil {
		fmt.Printf("get name failed, err:%v\n", err)
	} else {
		fmt.Printf("name:%v\n", val2)
	}
	return
}

func hgetDemo() {
	v, err := rdb.HGetAll(context.Background(), "user").Result()
	if err != nil {
		fmt.Printf("hgettall failed, err:%v\n", err)
		return
	}
	fmt.Println(v)
	//根据指定的key
	v2 := rdb.HMGet(context.Background(), "user", "name", "age")
	fmt.Println(v2)

	v3 := rdb.HGet(context.Background(), "user", "age")
	fmt.Println(v3)

}

// zset示例
func redisExample2() {
	// key
	zsetKey := "language_rank"
	// value
	// 注意：v8版本使用[]*redis.Z；此处为v9版本使用[]redis.Z
	languages := []redis.Z{
		{Score: 90.0, Member: "Golang"},
		{Score: 98.0, Member: "Java"},
		{Score: 95.0, Member: "Python"},
		{Score: 97.0, Member: "JavaScript"},
		{Score: 99.0, Member: "C/C++"},
	}
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// ZADD
	err := rdb.ZAdd(ctx, zsetKey, languages...).Err()
	if err != nil {
		fmt.Printf("zadd failed, err:%v\n", err)
		return
	}
	fmt.Println("zadd success")

	// 把Golang的分数加10
	newScore, err := rdb.ZIncrBy(ctx, zsetKey, 10.0, "Golang").Result()
	if err != nil {
		fmt.Printf("zincrby failed, err:%v\n", err)
		return
	}
	fmt.Printf("Golang's score is %f now.\n", newScore)

	// 取分数最高的3个
	ret := rdb.ZRevRangeWithScores(ctx, zsetKey, 0, 2).Val()
	for _, z := range ret {
		fmt.Println(z.Member, z.Score)
	}

	// 取95~100分的
	op := &redis.ZRangeBy{
		Min: "95",
		Max: "100",
	}
	ret, err = rdb.ZRangeByScoreWithScores(ctx, zsetKey, op).Result()
	if err != nil {
		fmt.Printf("zrangebyscore failed, err:%v\n", err)
		return
	}
	for _, z := range ret {
		fmt.Println(z.Member, z.Score)
	}
}

func main() {
	if err := initClient(); err != nil {
		fmt.Printf("init redis client err:%v\n", err)
		return
	}
	fmt.Println("init redis client success")
	//rdb.Get()
	//程序退出时释放相关资源
	redisExample2()

	pipe := rdb.Pipeline()

	//给这个key"pipeline_counter"加1
	incr := pipe.Incr(context.Background(), "pipeline_counter")
	//设置一个过期的时间:一个小时
	pipe.Expire(context.Background(), "pipeline_counter", time.Hour)

	_, err := pipe.Exec(context.Background())
	fmt.Println(incr.Val(), err)

	rdb.Close()
}
