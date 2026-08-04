package controller

import (
	"GoWeb/bookstore0612/dao"
	"GoWeb/bookstore0612/model"
	"GoWeb/bookstore0612/utils"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

// 添加图书到购物车
func AddBook2Cart(w http.ResponseWriter, r *http.Request) {

	//判断是否登录
	flag, session := dao.IsLogin(r)
	if flag {
		//已经登录
		//获取要添加的图书的Id
		bookID := r.FormValue("bookId")
		//根据图书id获取图书信息
		book, _ := dao.GetBookByID(bookID)

		//获取用户id
		userID := session.UserID

		//判断数据库中书否有当前用户的购物车
		cart, err := dao.GetCartByUserID(userID)
		if err != nil {
			//fmt.Println(err)
		}
		if cart != nil {
			//当前用户已经有购物车
			//判断购物车中是否有当前这本图书
			cartItem, _ := dao.GetCartItemByBookIDAndCartID(bookID, cart.CartID)
			if cartItem != nil {
				//购物车中已经有该图书，只需将对应的项加一就可以
				//获取购物车切片中所有的购物项
				cts := cart.CartItems
				//遍历得到每一个购物项
				for _, v := range cts {
					//找到对应的购物项
					if v.Book.ID == cartItem.Book.ID {
						//	fmt.Println(1)
						v.Count = v.Count + 1
						//fmt.Printf("%+v\n", v)
						//更新数据路
						dao.UpdateBookCount(v)
					}
				}

			} else {
				fmt.Println("当前购物车中还没有该图书")
				//购物项中还没有该图书，创建一个购物项并添加到数据库中
				cartItem := &model.CartItem{
					Book:   book,
					Count:  1,
					CartID: cart.CartID,
				}
				//将购物项添加到切片中
				cart.CartItems = append(cart.CartItems, cartItem)
				//将新创建的购物项添加到数据库中
				dao.AddCartItem(cartItem)
			}
			//不管之前购物车中是否有当前图书对应的购物项，都需要更新购物项和中的图书总数量和总金额
			dao.UpdateCart(cart)
		} else {

			//当前用户还没有购物车,创建购物车并添加到数据库
			//创建购物车
			//生成购物车id
			cartID := utils.CreateUUID()
			cart := &model.Cart{
				CartID: cartID,
				UserID: userID,
			}
			//创建购物项
			var cartItems []*model.CartItem
			cartItem := &model.CartItem{
				Book:   book,
				Count:  1,
				CartID: cartID,
			}
			//将购物项添加到切片中
			cartItems = append(cartItems, cartItem)
			//将切片设置到cart中
			cart.CartItems = cartItems
			//将购物车保存到数据库
			dao.AddCart(cart)

		}
		w.Write([]byte("刚刚将" + book.Title + "添加到购物车"))
	} else {
		//没有登录
		w.Write([]byte("请先登录！"))
	}
}

// 获取购物车信息
func GetCartInfo(w http.ResponseWriter, r *http.Request) {
	_, session := dao.IsLogin(r)
	userID := session.UserID
	//根据用户id获取对应购物车
	cart, _ := dao.GetCartByUserID(userID)
	//设置用户名

	if cart != nil {
		cart.UserName = session.UserName
		session.Cart = cart
		//有购物车,解析模板文件
		t := template.Must(template.ParseFiles("views/pages/cart/cart.html"))
		t.Execute(w, session)

	} else {
		//该用户还没有购物车
		t := template.Must(template.ParseFiles("views/pages/cart/cart.html"))

		t.Execute(w, session)
	}
}

// Deletecart 清空购物车
func DeleteCart(w http.ResponseWriter, r *http.Request) {
	//获取要删除的购物车的Id
	cartID := r.FormValue("cartId")
	//清空购物车
	dao.DeleteCartByCartID(cartID)
	//调用GetCartInfo函数再次查询购物车信息
	GetCartInfo(w, r)
}

// DeleteCartItem 删除购物项
func DeleteCartItem(w http.ResponseWriter, r *http.Request) {
	//获取要删除的购物项的id
	cartItemID := r.FormValue("cartItemId")
	//将购物项id转为int64
	iCartItemID, _ := strconv.ParseInt(cartItemID, 10, 64)
	//获取购物车session
	_, session := dao.IsLogin(r)
	userID := session.UserID
	//获取该用户购物车
	cart, _ := dao.GetCartByUserID(userID)
	//获取购物车中的购物项
	cartItems := cart.CartItems
	//遍历所有购物项
	for k, v := range cartItems {
		if v.CartItemID == iCartItemID {
			//删除购物项
			//将当前购物项从切片中移除
			cartItems = append(cartItems[:k], cartItems[k+1:]...)
			//
			cart.CartItems = cartItems
			//数据库层面
			dao.DeleteCartItemByID(cartItemID)
		}
	}

	//更新购物车中的金额
	dao.UpdateCart(cart)
	//获取购物车函数查询
	GetCartInfo(w, r)

}

// UpdateCartItem 更新购物项
func UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	//获取要更新的购物项的id
	cartItemID := r.FormValue("cartItemId")
	//将购物项id转为int64
	iCartItemID, _ := strconv.ParseInt(cartItemID, 10, 64)
	bookCount := r.FormValue("bookCount")
	ibookCount, _ := strconv.ParseInt(bookCount, 10, 64)

	//获取购物车session
	_, session := dao.IsLogin(r)
	userID := session.UserID
	//获取该用户购物车
	cart, _ := dao.GetCartByUserID(userID)
	//获取购物车中的购物项
	cartItems := cart.CartItems
	//遍历所有购物项
	for _, v := range cartItems {
		if v.CartItemID == iCartItemID {
			//更新购物项
			v.Count = ibookCount
			//更新数据库中购物想图书
			dao.UpdateBookCount(v)
		}
	}

	//更新购物车中的金额
	dao.UpdateCart(cart)
	//获取购物车函数查询
	cart, _ = dao.GetCartByUserID(userID)
	//GetCartInfo(w, r)
	//获取购物车中图书的总数量
	totalCount := cart.TotalCount
	//获取购物车中图书的总金额
	totalAmount := cart.TotalAmount
	var amount float64
	//获取购物车中更新的购物项中的金额小计
	cIs := cart.CartItems
	for _, v := range cIs {
		if v.CartItemID == iCartItemID {
			//索要找的购物项,获取当前购物项中的金额小计
			amount = v.Amount
		}
	}
	//创建data结构
	data := model.Data{
		Amount:      amount,
		TotalAmount: totalAmount,
		TotalCount:  totalCount,
	}
	//将data转换为jison字符串
	json, _ := json.Marshal(data)

	//相应到客户端浏览器
	w.Write(json)
}
