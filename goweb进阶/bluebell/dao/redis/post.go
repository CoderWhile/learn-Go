package redis

import (
	"bluebell/models"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func GetPostIDsInOrder(c *gin.Context, p *models.ParamPostList) ([]string, error) {
	//  从redis获取id
	// 根据用户请求中携带的order参数确定要查询的redis KEy
	key := getRedisKey(KeyPostTimeZSet)
	if p.Order == models.OrderScore {
		key = getRedisKey(KeyPostScoreZSet)
	}
	// 2.确认查询的索引起始点
	start := (p.Page - 1) * p.Size
	end := start + p.Size - 1
	// 3.ZREVRANGE 按分数从大到小查询指定数量的元素
	return client.ZRevRange(c, key, start, end).Result()
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
