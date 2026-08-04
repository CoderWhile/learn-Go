package main

import (
	"GoWeb/bookstore0612/controller"
	"net/http"
)

// 去首页的处理器
//func IndexHandler(w http.ResponseWriter, r *http.Request) {
//	//解析模板
//	//去首页是通过引擎
//	t := template.Must(template.ParseFiles("views/index.html"))
//	t.Execute(w, "")
//}

func main() {

	//设置处理静态资源
	//直接去页面
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("views/static/"))))
	http.Handle("/pages/", http.StripPrefix("/pages/", http.FileServer(http.Dir("views/pages/"))))

	http.HandleFunc("/main", controller.GetPageBooksByPrice)

	//去登陆:
	http.HandleFunc("/login", controller.Login)
	http.HandleFunc("/logout", controller.Logout)

	//去注册
	http.HandleFunc("/regist", controller.Regist)

	//通过Ajax请求验证用户名是否可用
	http.HandleFunc("/checkUserName", controller.CheckUserName)

	//获取所有图书
	//http.HandleFunc("/getBooks", controller.GetBooks)
	//获取带分页的图书信息
	http.HandleFunc("/getPageBooks", controller.GetPageBooks)

	http.HandleFunc("/getPageBooksByPrice", controller.GetPageBooksByPrice)

	//添加图书
	//http.HandleFunc("/addBook", controller.AddBook)

	//删除图书
	http.HandleFunc("/deleteBook", controller.DeleteBook)

	//去更新图书的页面
	http.HandleFunc("/toUpdateBookPage", controller.ToUpdateBookPage)

	//更新图书
	http.HandleFunc("/updateOrAddBook", controller.UpdateOrADdBook)

	//添加图书到购物车
	http.HandleFunc("/addBook2Cart", controller.AddBook2Cart)

	//获取购物车信息
	http.HandleFunc("/getCartInfo", controller.GetCartInfo)

	//清空购物车
	http.HandleFunc("/deleteCart", controller.DeleteCart)

	//删除购物项
	http.HandleFunc("/deleteCartItem", controller.DeleteCartItem)

	///updateCartItem 更新购物项
	http.HandleFunc("/updateCartItem", controller.UpdateCartItem)

	//去结账
	http.HandleFunc("/checkout", controller.Checkout)

	//获取所有订单
	http.HandleFunc("/getOrders", controller.GetOrders)

	//获取订单详情,以及所对应的所有订单项
	http.HandleFunc("/getOrderInfo", controller.GetOrderInfo)

	//获取我的订单
	http.HandleFunc("/getMyOrders", controller.GetMyOrders)

	//发货
	http.HandleFunc("/sendOrder", controller.SendOrder)

	//确认收货
	http.HandleFunc("/takeOrder", controller.TakeOrder)
	http.ListenAndServe(":8080", nil)
}
