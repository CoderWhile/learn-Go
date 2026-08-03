package model

type CartItem struct {
	CartItemID int64   //购物项的Id
	Book       *Book   //购物项中图书信息
	Count      int64   //购物项中图书数量
	Amount     float64 //购物项中图书的金额小计，通过计算得到
	CartID     string  //当前购物项属于哪一个购物车

}

// GetAmount获取购物项中图书的金额小计，有图书价格和图书数量
func (cartItem *CartItem) GetAmount() float64 {
	//获取当前购物项中图书的价格
	price := cartItem.Book.Price
	
	return float64(cartItem.Count) * price
}
