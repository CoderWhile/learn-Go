package redis

import (
	"bluebell/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func getIDsFormKey(c *gin.Context, key string, page, size int64) ([]string, error) {
	start := (page - 1) * size
	end := start + size - 1
	// 	ZREVRANGE 按分数从小到大的顺序查询指定数量的元素
	return client.ZRevRange(c, key, start, end).Result()
}

func GetPostIDsInOrder(c *gin.Context, p *models.ParamPostList) ([]string, error) {
	//  从redis获取id
	// 根据用户请求中携带的order参数确定要查询的redis KEy
	key := getRedisKey(KeyPostTimeZSet)
	if p.Order == models.OrderScore {
		key = getRedisKey(KeyPostScoreZSet)
	}
	// 2.确认查询的索引起始点
	return getIDsFormKey(c, key, p.Page, p.Size)
}

// GetPostVoteDate 根据ids查询每篇帖子的投赞成票的数据
func GetPostVoteDate(c *gin.Context, ids []string) (data []int64, err error) {
	data = make([]int64, 0, len(ids))
	//for _, id := range ids {
	//	key := getRedisKey(KeyPostVotedZSetPF + id)
	//	// 查找key中分数是1的元素的数量->统计每篇帖子的赞成票数
	//	v := client.ZCount(c, key, "1", "1").Val()
	//	data = append(data, v)
	//
	//}
	// 使用pipeline一次发送多套命令,减少RTT
	pipeline := client.Pipeline()
	for _, id := range ids {
		key := getRedisKey(KeyPostVotedZSetPF + id)
		pipeline.ZCount(c, key, "1", "1")

	}
	cmders, err := pipeline.Exec(c)
	if err != nil {
		return nil, err
	}
	data = make([]int64, 0, len(cmders))

	for _, cmder := range cmders {
		v := cmder.(*redis.IntCmd).Val()
		data = append(data, v)
	}

	return
}

// GetCimmunity 按社区查询ids
func GetCommunityPostIDsInOrder(c *gin.Context, p *models.ParamPostList) ([]string, error) {
	// 使用zinterstore 把分区的帖子set与帖子分数的zset 生成一个新的zset
	orderkey := getRedisKey(KeyPostTimeZSet)
	if p.Order == models.OrderScore {
		orderkey = getRedisKey(KeyPostScoreZSet)
	}
	// 针对新的zset 按之前的逻辑取数据
	// 社区的key
	cKey := getRedisKey(KeyCommunitySetPF + strconv.Itoa(int(p.CommunityID)))
	// 利用缓存key减少zinterstore执行的次数
	key := orderkey + strconv.Itoa(int(p.CommunityID))
	if client.Exists(c, key).Val() < 1 {
		//不存在,需要计算
		pipeline := client.Pipeline()
		pipeline.ZInterStore(c, key, &redis.ZStore{
			Keys:      []string{cKey, orderkey},
			Aggregate: "MAX",
		}) /// 计算
		pipeline.Expire(c, key, 60*time.Second) // 设置超时时间
		_, err := pipeline.Exec(c)
		if err != nil {
			return nil, err
		}
	}
	return getIDsFormKey(c, key, p.Page, p.Size)
}
