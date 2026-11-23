package redis

import (
	"errors"
	"math"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	oneWeekInSeconds = 7 * 24 * 3600
	scorePerVote     = 432 //每一票值多少数
)

var (
	ErrVoteTimeExpire = errors.New("投票时间已过")
	ErrVoteRepested   = errors.New("不允许重复投票")
)

func CreatPost(c *gin.Context, postID int64) error {
	//这两个要么同时成功要么同时不成功-->事务
	pipeline := client.TxPipeline()

	//帖子时间发布什么时候就什么时候
	pipeline.ZAdd(c, getRedisKey(KeyPostTimeZSet), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})

	//帖子分数,根据点赞发生变化
	pipeline.ZAdd(c, getRedisKey(KeyPostScoreZSet), redis.Z{
		Score:  float64(time.Now().Unix()),
		Member: postID,
	})
	_, err := pipeline.Exec(c)

	if err != nil {
		zap.L().Error("Redis.SAdd", zap.Error(err))
	}
	return err
}

func VoteForPost(c *gin.Context, userID, postID string, value float64) error {
	// 1.判断投票的限制
	//去redis 取帖子发布时hi见
	postTime := client.ZScore(c, getRedisKey(KeyPostTimeZSet), postID).Val()
	if float64(time.Now().Unix())-postTime > oneWeekInSeconds {
		return ErrVoteTimeExpire
	}
	//第二步和第三步也要放到事务里面去

	// 2.跟新帖子的分数
	// 先查当前用户给当前帖子的投票记录
	ov := client.ZScore(c, getRedisKey(KeyPostVotedZSetPF+postID), userID).Val()
	//  更新    如果这次投票和之前一致不允许
	if value == ov {
		return ErrVoteRepested
	}
	var op float64
	//现在的值大于
	if value > ov {
		op = 1
	} else {
		op = -1
	}
	diff := math.Abs(ov - value) //计算两次投票的差值
	pipeline := client.TxPipeline()
	pipeline.ZIncrBy(c, getRedisKey(KeyPostScoreZSet), op*diff*scorePerVote, postID)

	// 3.记录用户为该帖子投过票的数据

	if value == 0 {
		//移除投票项
		pipeline.ZRem(c, getRedisKey(KeyPostVotedZSetPF+postID), postID)
	} else {
		pipeline.ZAdd(c, getRedisKey(KeyPostVotedZSetPF+postID), redis.Z{
			Score:  value, //当前用户有投的是赞成票还是反对票
			Member: userID,
		})
	}
	_, err := pipeline.Exec(c)
	return err
}
