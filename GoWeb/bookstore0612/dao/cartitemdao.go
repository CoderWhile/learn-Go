package dao

import (
	"GoWeb/bookstore0612/model"
	"GoWeb/bookstore0612/utils"
)

// 向购物项表中插入购物项
func AddCartItem(cartItem *model.CartItem) error {
	sqlStr := `insert into cart_items(count,amount,book_id,cart_id) values(?,?,?,?)`
	_, err := utils.Db.Exec(sqlStr, cartItem.Count, cartItem.GetAmount(), cartItem.Book.ID, cartItem.CartID)
	if err != nil {
		return err
	}
	return nil
}

// GetCartItemByBookID根据图书的id获取对应的购物项
func GetCartItemByID(bookID string) (*model.CartItem, error) {
	sqlStr := `select id,count,amount,cart_id from cart_items where book_id=?`
	row := utils.Db.QueryRow(sqlStr, bookID)
	cartItem := &model.CartItem{}
	err := row.Scan(&cartItem.CartItemID, &cartItem.Count, &cartItem.Amount, &cartItem.CartID)
	if err != nil {
		return nil, err
	}
	return cartItem, nil
}

// UpdateBookCount根据购物项更新购物项中图书的数量和金额小计
func UpdateBookCount(cartItem *model.CartItem) error {
	sql := `update cart_items set count= ? , amount = ? where book_id= ?  and cart_id = ? `
	//fmt.Println(bookCount, cartID)
	_, err := utils.Db.Exec(sql, cartItem.Count, cartItem.GetAmount(), cartItem.Book.ID, cartItem.CartID)
	if err != nil {
		return err
	}
	return nil
}

// 根据图书的id和购物车的id获取对应的购物项
func GetCartItemByBookIDAndCartID(bookID string, cartID string) (*model.CartItem, error) {
	sqlStr := `select id,count,amount,cart_id from cart_items where book_id=? and cart_id=?`
	row := utils.Db.QueryRow(sqlStr, bookID, cartID)
	cartItem := &model.CartItem{}
	err := row.Scan(&cartItem.CartItemID, &cartItem.Count, &cartItem.Amount, &cartItem.CartID)
	if err != nil {
		return nil, err
	}
	book, _ := GetBookByID(bookID)
	cartItem.Book = book
	return cartItem, nil
}

// GetCartItemsByCartID 根据购物车的id获取购物车中所有的购物项
func GetCartItemsByCartID(cartID string) ([]*model.CartItem, error) {
	sqlStr := `select id,count,amount,book_id,cart_id from cart_items where cart_id=?`
	rows, err := utils.Db.Query(sqlStr, cartID)
	if err != nil {
		return nil, err
	}
	var cartItems []*model.CartItem

	for rows.Next() {
		var bookID string
		cartItem := &model.CartItem{}
		err2 := rows.Scan(&cartItem.CartItemID, &cartItem.Count, &cartItem.Amount, &bookID, &cartItem.CartID)
		if err != nil {
			return nil, err2
		}
		//根据bookID获取图书信息
		book, _ := GetBookByID(bookID)
		cartItem.Book = book
		cartItems = append(cartItems, cartItem)
	}

	return cartItems, nil
}

// DeleteCartItemsByCartID 根据购物车的Id删除所有购物项
func DeleteCartItemsByCartID(cartID string) error {
	sql := `delete from cart_items where cart_id=?`
	_, err := utils.Db.Exec(sql, cartID)
	if err != nil {
		return err
	}
	return nil
}

// DeleteCartItemByID根据购物项的id删除购物项
func DeleteCartItemByID(cartItemID string) error {
	sql := `delete from cart_items where id=?`
	_, err := utils.Db.Exec(sql, cartItemID)
	if err != nil {
		return err
	}
	return nil
}
