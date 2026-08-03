package model

type Cart struct {
	CartID      string      //购物车id
	CartItems   []*CartItem //购物车中所有的购物项
	TotalCount  int64       //总数量，通过计算得到0
	TotalAmount float64     //购物车中总金额
	UserID      int         //当前购物车所属的用户
}

// GetTotalCount
func (cart *Cart) GetTotalCount() int64 {
	//遍历购物项切片
	var totalCount int64
	for _, v := range cart.CartItems {
		totalCount = totalCount + v.Count
	}
	return totalCount
}

// 获取购物车中图书总金额
func (cart *Cart) GetTotalAmount() float64 {
	var totalAmount float64

	for _, v := range cart.CartItems {
		totalAmount = totalAmount + v.GetAmount()
	}
	return totalAmount
}
