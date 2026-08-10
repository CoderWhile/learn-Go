package model

type CartItem struct {
	CartItemID int64   //购物项的Id
	Ticket     *Ticket //购物项中票信息
	CartID     string  //当前购物项属于哪一个购物车

}
