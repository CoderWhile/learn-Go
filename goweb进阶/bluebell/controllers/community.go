package controllers

import (
	"bluebell/logic"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

//与社区相关

func CommunityHandler(c *gin.Context) {
	//期望查询到所有社区(community_id,community_name)以列表的形式返回
	data, err := logic.GetCommunityList()
	if err != nil {
		zap.L().Error("logic.GetCommunityList failed", zap.Error(err))
		ResponseError(c, CodeServerBusy) //不轻易把服务端报错暴露给外面
		return
	}
	ResponseSuccess(c, data)
}

func CommunityDetailHandler(c *gin.Context) {
	//1.获取社区id
	idstr := c.Param("id") //获取url参数
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ResponseError(c, CodeInvalidParam) //参数错误
		return
	}
	//期望查询到所有社区(community_id,community_name)以列表的形式返回
	//根据id获取社区详情
	fmt.Printf("id:%d\n", id)
	data, err := logic.GetCommunityDetail(id)
	if err != nil {
		zap.L().Error("logic.GetCommunityDetail failed", zap.Error(err))
		ResponseError(c, CodeServerBusy) //不轻易把服务端报错暴露给外面
		return
	}
	ResponseSuccess(c, data)
}
