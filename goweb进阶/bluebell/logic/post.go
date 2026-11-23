package logic

import (
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/models"
	"bluebell/pkg/snowflake"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func CreatePost(c *gin.Context, p *models.Post) (err error) {
	//1.生成postid
	p.ID = snowflake.GenID()

	//2.保存到数据库
	err = mysql.CreatePost(p)
	if err != nil {
		return err
	}
	err = redis.CreatPost(c, p.ID)
	return
	//3.返回
}

func GetPostById(pid int64) (data *models.ApiPostDetail, err error) {
	//查询并组合我们接口想用的数据
	post, err := mysql.GetPostById(pid)
	if err != nil {
		zap.L().Error("mysql.GetPostById(pid) failed", zap.Int64("pid", pid), zap.Error(err))
		return
	}

	//根据作者id查询作者信息
	user, err := mysql.GetUseById(post.AuthorID)
	if err != nil {
		zap.L().Error("mysql.GetUseById(post.AuthorID) failed", zap.Int64("author_id", post.AuthorID), zap.Error(err))
		return
	}
	//根据社区id查询社区详细信息
	community, err := mysql.GetCommunityDetailByID(post.CommunityID)
	if err != nil {
		zap.L().Error("mysql.GetCommunityDetailByID(post.CommunityID) failed", zap.Int64("community_id", post.CommunityID), zap.Error(err))
		return
	}
	//接口数据拼接
	data = &models.ApiPostDetail{
		AuthorName:      user.Username,
		Post:            post,
		CommunityDetail: community,
	}

	return
}

// GetPostList //获取帖子列表
func GetPostList(page, size int64) (data []*models.ApiPostDetail, err error) {
	//查询并组合我们接口想用的数据
	posts, err := mysql.GetPostList(page, size)
	if err != nil {
		zap.L().Error("mysql.GetPostList(pid) failed", zap.Error(err))
		return
	}
	data = make([]*models.ApiPostDetail, 0, len(posts))
	//根据作者id查询作者信息
	for _, post := range posts {

		user, err := mysql.GetUseById(post.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUseById(post.AuthorID) failed", zap.Int64("author_id", post.AuthorID), zap.Error(err))
			continue
		}
		//根据社区id查询社区详细信息
		community, err := mysql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetCommunityDetailByID(post.CommunityID) failed", zap.Int64("community_id", post.CommunityID), zap.Error(err))
			continue
		}
		//接口数据拼接
		postdetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			Post:            post,
			CommunityDetail: community,
		}
		data = append(data, postdetail)
	}

	return
}

func GetPostList2(c *gin.Context, p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	// 2.去redis查询id列表

	ids, err := redis.GetPostIDsInOrder(c, p)
	if err != nil {
		return
	}
	if len(ids) == 0 {
		zap.L().Warn("redis.GEtPostIDsInOrder(p) return 0 data ")
		return
	}
	zap.L().Debug("GetPostList2", zap.Any("ids", ids))
	// 3.根据Id去MYsql数据库查询帖子详细信息
	// 返回的数据还要按照给定的id的顺序返回
	// 在mysql查询的时候用到了Order by FIND_IN_SET函数
	// 或者按照之前的顺序排序
	posts, err := mysql.GetPostListByIDs(ids)
	if err != nil {
		return
	}
	zap.L().Debug("GetPostList2", zap.Any("posts", posts))
	// 提前查询好每篇帖子的投票数
	voteData, err := redis.GetPostVoteDate(c, ids)
	if err != nil {
		return
	}
	// 将帖子的作者及分区信息查询出来填充的奥帖子中
	for idx, post := range posts {

		user, err := mysql.GetUseById(post.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUseById(post.AuthorID) failed", zap.Int64("author_id", post.AuthorID), zap.Error(err))
			continue
		}
		//根据社区id查询社区详细信息
		community, err := mysql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetCommunityDetailByID(post.CommunityID) failed", zap.Int64("community_id", post.CommunityID), zap.Error(err))
			continue
		}
		//接口数据拼接
		postdetail := &models.ApiPostDetail{
			VoteNum:         voteData[idx],
			AuthorName:      user.Username,
			Post:            post,
			CommunityDetail: community,
		}
		data = append(data, postdetail)
	}
	return
}
