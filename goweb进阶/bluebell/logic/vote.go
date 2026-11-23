package logic

import (
	"bluebell/dao/redis"
	"bluebell/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 投票功能
// 1.用户投票得数据

//投一票就加432分
//随时间的走势
//用户投票帖子分数就增加

//投票的几种情况
/*
direction=1时,有两种情况
	1.之前没有投过票,现在头赞成票 --> 更新分数和投票记录 差值的绝对值:1 +432
	2.之前投反对票,现在改成赞成的 --> 更新分数和投票记录 差值绝对值:2 +432*2
direction=0时,有两种情况
	1.之前投赞成票,现在取消投票 -->1 -432
	2.之前投反对票,现在取消投票 -->1 +432
direction=-1时,有两种情况:
	1.之前没有投过票,现在图反对票  --> 1 -432
	2.之前投赞成票,现在改投反对票  -->2 -432*2

投票的限制:
每个帖子自发表之日起一个星期内允许用户投票,超过一个星期就不允许再投票了
	1.到期之后将redis中保存的赞成票数及反对票数存储到mysql中
	2.到期之后删除那个 KeyPostVotedZSetPF
*/

// VoteForPost 为帖子投票的函数
func VoteForPost(c *gin.Context, userID int64, p *models.ParamVoteData) error {
	// 1.判断投票的限制
	zap.L().Debug("VoteForPost",
		zap.Int64("userID", userID),
		zap.String("postID", p.PostID),
		zap.Int8("direction", p.Direction))
	return redis.VoteForPost(c, strconv.Itoa(int(userID)), p.PostID, float64(p.Direction))

}
