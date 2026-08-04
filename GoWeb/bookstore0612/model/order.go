package model

type Order struct {
	OrderID     string
	CreateTime  string //生成订单时间
	TotalCount  int64  //订单中图书总数量
	TotalAmount float64
	State       int64 //0：未发货 1：已发货 2：交易完成
	UserID      int64 //订单所属用户

}

// 未发货
func (order *Order) NoSend() bool {
	return order.State == 0
}

// 已发货
func (order *Order) SendComplate() bool {
	return order.State == 1
}

// 是否交易完成
// 交易完成
func (order *Order) Complate() bool {
	return order.State == 2
}
