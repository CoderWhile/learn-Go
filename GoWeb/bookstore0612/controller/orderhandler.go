package controller

import (
	"GoWeb/bookstore0612/dao"
	"GoWeb/bookstore0612/model"
	"GoWeb/bookstore0612/utils"
	"html/template"
	"net/http"
	"time"
)

// 去结账
func Checkout(w http.ResponseWriter, r *http.Request) {
	//获取session
	_, session := dao.IsLogin(r)
	//获取用户id
	userID := session.UserID
	cart, _ := dao.GetCartByUserID(userID)
	//生成订单号
	orderID := utils.CreateUUID()
	order := &model.Order{
		OrderID:     orderID,
		CreateTime:  time.Now().Format("2006-01-02 15:04:05"),
		TotalCount:  cart.TotalCount,
		TotalAmount: cart.TotalAmount,
		State:       0,
		UserID:      int64(userID),
	}
	dao.AddOrder(order)
	//保存订单项
	//获取购物车中的购物项
	cartItems := cart.CartItems
	for _, v := range cartItems {
		orderItem := &model.OrderItem{
			Count:   v.Count,
			Amount:  v.Amount,
			Title:   v.Book.Title,
			Author:  v.Book.Author,
			Price:   v.Book.Price,
			ImgPath: v.Book.ImagePath,
			OrderID: orderID,
		}
		//保存购物项
		dao.AddOrderItem(orderItem)
		//更新当前购物项中图书的库存和销量
		book := v.Book
		book.Sales = book.Sales + int(v.Count)
		book.Stock = book.Stock - int(v.Count)
		//更新图书的信息
		dao.UpdateBook(book)

	}
	//清空购物车
	dao.DeleteCartByCartID(cart.CartID)
	//将订单号设置到session中
	session.Order = order
	//解析模板
	t := template.Must(template.ParseFiles("views/pages/cart/checkout.html"))
	t.Execute(w, session)
}

// GetOrders 获取所有订单
func GetOrders(w http.ResponseWriter, r *http.Request) {
	orders, _ := dao.GetOrders()
	//解析模板
	t := template.Must(template.ParseFiles("views/pages/order/order_manager.html"))
	t.Execute(w, orders)
}

// 获取所有订单信息
func GetOrderInfo(w http.ResponseWriter, r *http.Request) {
	//获取订单号
	orderID := r.FormValue("orderId")
	//获取所有订单项
	orderItems, _ := dao.GetOrderItemsByOrderID(orderID)
	//解析模板
	t := template.Must(template.ParseFiles("views/pages/order/order_info.html"))
	t.Execute(w, orderItems)
}

// 获取我的订单
func GetMyOrders(w http.ResponseWriter, r *http.Request) {
	_, session := dao.IsLogin(r)
	userID := session.UserID
	orders, _ := dao.GetMyOrders(userID)
	//将订单设置到session中
	session.Orders = orders
	t := template.Must(template.ParseFiles("views/pages/order/order.html"))
	t.Execute(w, session)
}

// SendOrder 发货
func SendOrder(w http.ResponseWriter, r *http.Request) {
	//获取要发货的订单号
	orderID := r.FormValue("orderID")
	//发货
	dao.UpdateOrderState(orderID, 1)
	//再次获取所有订单
	GetOrders(w, r)

}

// Takeorder 收货
func TakeOrder(w http.ResponseWriter, r *http.Request) {
	//获取要发货的订单号
	orderID := r.FormValue("orderId")
	//发货
	dao.UpdateOrderState(orderID, 2)
	//调用霍夫区我的订单
	GetMyOrders(w, r)

}
