package dao

import (
	"GoWeb/bookstore0612/model"
	"GoWeb/bookstore0612/utils"
	"fmt"
)

// AddOrder 向数据库中插入订单
func AddOrder(order *model.Order) error {
	sql := `insert into orders(id,create_time,total_count,total_amount,state,user_id) values(?,?,?,?,?,?)`
	_, err := utils.Db.Exec(sql, order.OrderID, order.CreateTime, order.TotalCount, order.TotalAmount, order.State, order.UserID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// GetOrders获取数据库中所有的订单
func GetOrders() ([]*model.Order, error) {
	sql := `select id,create_time,total_count,total_amount,state,user_id from orders`
	rows, err := utils.Db.Query(sql)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var orders []*model.Order
	for rows.Next() {
		order := &model.Order{}
		rows.Scan(&order.OrderID, &order.CreateTime, &order.TotalCount, &order.TotalAmount, &order.State, &order.UserID)
		orders = append(orders, order)
	}
	return orders, nil
}

// GetMyOrders获取我的订单
func GetMyOrders(userID int) ([]*model.Order, error) {
	sql := `select id,create_time,total_count,total_amount,state,user_id from orders where user_id = ?`
	rows, err := utils.Db.Query(sql, userID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	var orders []*model.Order
	for rows.Next() {
		order := &model.Order{}
		rows.Scan(&order.OrderID, &order.CreateTime, &order.TotalCount, &order.TotalAmount, &order.State, &order.UserID)
		orders = append(orders, order)
	}
	return orders, nil
}

// UpdateOrderState 更新订单的状态,即发货和收获
func UpdateOrderState(orderID string, state int64) error {
	sql := `update orders set state=? where id=?`
	_, err := utils.Db.Exec(sql, state, orderID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
