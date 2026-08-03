package controller

import (
	"GoWeb/bookstore0612/dao"
	"GoWeb/bookstore0612/model"
	"GoWeb/bookstore0612/utils"
	"fmt"
	"html/template"
	"net/http"
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
			fmt.Println(err)
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
						dao.UpdateBookCount(v.Count, v.Book.ID, cart.CartID)
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
	if cart != nil {
		//有购物车,解析模板文件
		t := template.Must(template.ParseFiles("views/pages/cart/cart.html"))
		t.Execute(w, cart)

	}
}
