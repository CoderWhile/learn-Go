package controllers

import (
	"bluebell/logic"
	"bluebell/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// CreatePostHandler 创建帖子的处理函数
func CreatePostHandler(c *gin.Context) {
	//1.获取参数及参数的校验
	//c.ShouldBindJSON()嵌入了一个validator-->binding tag去做一个数据的绑定
	p := new(models.Post)
	//要知道提交的数据都有哪些规则通过binding这个标签规定
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Debug("c.ShouldBindJSON(p) error", zap.Any("err", err))
		zap.L().Error("create post with invalid param")
		ResponseError(c, CodeInvalidParam)
		return
	}

	//从c取到当前发请求的用户的ID
	userID, err := GetCurrentUserID(c)
	if err != nil {
		ResponseError(c, CodeNeedLogin)
		return
	}
	p.AuthorID = userID
	//2.创建帖子
	if err := logic.CreatePost(c, p); err != nil {
		zap.L().Error("logic.CreatePost(p)  failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	//3.返回响应
	ResponseSuccess(c, nil)
}

// GetPostDetailHandler 获取帖子详情的处理函数
func GetPostDetailHandler(c *gin.Context) {

	//1.获取参数(帖子的idURL中获取
	pidStr := c.Param("id")
	pid, err := strconv.ParseInt(pidStr, 10, 64)
	if err != nil {
		zap.L().Error("get post detail with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	//2.根据Id获取帖子数据(查数据库)
	data, err := logic.GetPostById(pid)
	if err != nil {
		zap.L().Error("logic.GetPostById(pid)  failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	//3. 返回响应
	ResponseSuccess(c, data)
}

// GetPostListDetail 获取帖子列个表处理函数
func GetPostListHandler(c *gin.Context) {
	//获取分页参数
	page, size := GetPageInfo(c)
	// 获取数据
	data, err := logic.GetPostList(page, size)
	if err != nil {
		zap.L().Error("logic.GetPostList() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	//返回响应
	ResponseSuccess(c, data)
}

// GetPostListHandler2 升级版帖子列表接口
// @Summary 升级版帖子列表接口
// @Description 可按社区按时间或分数排序查询帖子列表接口
// @Tags 帖子相关就i恶口(api分组展示使用的)
// @Accept application/json
// @Produce application/json
// @Param authorizaton header string true "Bearer JWT"
// @Security ApiKeyAuth
// @Seccess 200 {object} _ResponsePostList
// @Router /posts2 [get]
func GetPostListHandler2(c *gin.Context) {

	// 1.获取参数query string参数
	// 2.去redis查询id列表
	// 3.根据id去数据库查询帖子详细信息
	//	GET请求参数(query string): /api/v1/posts2?page=1&sizse=10&order=time
	// 初始化结构体时指定初始参数
	// c.ShouldBind() 根据请求的数据类型选择响应的方法去获取数据
	//c.ShouldBindJSON()如果请求中携带的是json格式的数据,才能用这个方法获取

	// 指定上默认的值最好事改成配置文件的
	p := &models.ParamPostList{
		Page:  1,
		Size:  10,
		Order: models.OrderTime,
	}
	if err := c.ShouldBindQuery(p); err != nil {
		zap.L().Error("get postslistHandler2 list with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	data, err := logic.GetPostListNew(c, p) //更新合二为一
	// 获取数据
	if err != nil {
		zap.L().Error("logic.GetPostListNew() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	//返回响应
	ResponseSuccess(c, data)
}

//// 根据社区去查询帖子列表
//func GetCommunityPostListHandler(c *gin.Context) {
//	// 初始化结构体时指定参数
//	p := &models.ParamCommunityPostList{
//		ParamPostList: &models.ParamPostList{
//			Page:  1,
//			Size:  10,
//			Order: models.OrderTime,
//		},
//	}
//	if err := c.ShouldBindQuery(p); err != nil {
//		zap.L().Error("get CommunitypostslistHandler list with invalid param", zap.Error(err))
//		ResponseError(c, CodeInvalidParam)
//		return
//	}
//	// 获取数据
//	if err != nil {
//		zap.L().Error("logic.GetCommunityPostList() failed", zap.Error(err))
//		ResponseError(c, CodeServerBusy)
//		return
//	}
//	//返回响应
//	ResponseSuccess(c, data)
//}
