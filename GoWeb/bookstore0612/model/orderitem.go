package model

// 订单项
type OrderItem struct {
	OrderItemID int64
	Count       int64   //订单项中图书数量
	Amount      float64 //金额小计
	Title       string
	Author      string
	Price       float64
	ImgPath     string //订单项中图书的封面
	OrderID     string //订单项所属的订单
}
