package controllers

import "bluebell/models"

// 专门放接口文档用到的Model
type _ResponsePostList struct {
	Code    ResCode                 `json:"code"` //业务响应状态码
	Message string                  `json:"msg"`  //提示信息
	Data    []*models.ApiPostDetail `json:"data"` //数据
}
