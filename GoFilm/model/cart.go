package model

type Cart struct {
	CartID    string      //购物车id
	CartItems []*CartItem //购物车中所有的购物项

	TotalAmount float64 //购物车中总金额
	UserID      int     //当前购物车所属的用户
	UserName    string
}

//
//// 获取购物车中图书总金额
//func (cart *Cart) GetTotalAmount() float64 {
//	var totalAmount float64
//
//	for _, v := range cart.CartItems {
//		totalAmount = totalAmount + v.Ticket.
//	}
//	return totalAmount
//}
