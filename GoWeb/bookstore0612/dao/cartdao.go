package dao

import (
	"GoWeb/bookstore0612/model"
	"GoWeb/bookstore0612/utils"
)

// AddCart享购物车表中插入购物车
func AddCart(cart *model.Cart) error {
	sqlStr := `insert into carts(id,total_count,total_amount,user_id) values(?,?,?,?)`
	_, err := utils.Db.Exec(sqlStr, cart.CartID, cart.GetTotalCount(), cart.GetTotalAmount(), cart.UserID)
	if err != nil {
		return err
	}
	//获取购物车中所有购物项
	cartItems := cart.CartItems
	for _, cartItem := range cartItems {
		//保存购物想
		AddCartItem(cartItem)
	}
	return nil
}

// GetCartByUserID 根据用户的id从数据库中查询对应的购物车
func GetCartByUserID(userID int) (*model.Cart, error) {
	sql := `select id,total_count,total_amount,user_id from carts where user_id = ?`
	row := utils.Db.QueryRow(sql, userID)
	//创建一个购物车
	cart := &model.Cart{}
	err := row.Scan(&cart.CartID, &cart.TotalCount, &cart.TotalAmount, &cart.UserID)
	if err != nil {
		return nil, err
	}
	//获取当前购物车中所有购物想
	cartItems, _ := GetCartItemsByCartID(cart.CartID)
	cart.CartItems = cartItems
	return cart, nil
}

// UpdateCart 更新购物车中图书的总数量和总金额
func UpdateCart(cart *model.Cart) error {
	sqlStr := `update carts set total_count=?,total_amount=? where id=?`

	_, err := utils.Db.Exec(sqlStr, cart.GetTotalCount(), cart.GetTotalAmount(), cart.CartID)
	if err != nil {
		return err
	}
	return nil
}
