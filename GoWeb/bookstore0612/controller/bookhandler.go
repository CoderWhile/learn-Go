package controller

import (
	"GoWeb/bookstore0612/dao"
	"GoWeb/bookstore0612/model"
	"html/template"
	"net/http"
	"strconv"
)

// 获取所有图书
func GetBooks(w http.ResponseWriter, r *http.Request) {
	//调用bookdao中的
	books, _ := dao.GetBooks()
	t := template.Must(template.ParseFiles("views/pages/manager/book_manager.html"))
	t.Execute(w, books)
}

// 添加图书
//func AddBook(w http.ResponseWriter, r *http.Request) {
//	//获取图书信息
//	title := r.PostFormValue("title")
//	author := r.PostFormValue("author")
//	price := r.PostFormValue("price")
//	sales := r.PostFormValue("sales")
//	stock := r.PostFormValue("stock")
//	fPrice, _ := strconv.ParseFloat(price, 64)
//	iSales, _ := strconv.ParseInt(sales, 10, 0)
//	iStock, _ := strconv.ParseInt(stock, 10, 0)
//
//	book := &model.Book{Title: title, Author: author, Price: fPrice, Sales: int(iSales), Stock: int(iStock), ImagePath: "/static/img/default.jpg"}
//	dao.AddBook(book)
//	//调用GetBooks处理器函数再查询一次数据库
//	GetBooks(w, r)
//}

// 删除图书
func DeleteBook(w http.ResponseWriter, r *http.Request) {
	//获取图书id
	bookID := r.FormValue("bookId")
	dao.DeleteBook(bookID)
	GetBooks(w, r)
}

// 去更新或添加图书的页面
func ToUpdateBookPage(w http.ResponseWriter, r *http.Request) {
	//获取要更新图书id
	bookID := r.FormValue("bookId")
	book, _ := dao.GetBookByID(bookID)
	if book.ID > 0 {
		//在更新图书
		t := template.Must(template.ParseFiles("views/pages/manager/book_edit.html"))

		t.Execute(w, book)
	} else {
		//在添加图书
		t := template.Must(template.ParseFiles("views/pages/manager/book_edit.html"))

		t.Execute(w, "")
	}

}

// 更新或添加图书
func UpdateOrADdBook(w http.ResponseWriter, r *http.Request) {
	bookID := r.PostFormValue("bookId")
	title := r.PostFormValue("title")
	author := r.PostFormValue("author")
	price := r.PostFormValue("price")
	sales := r.PostFormValue("sales")
	stock := r.PostFormValue("stock")
	ibookID, _ := strconv.ParseInt(bookID, 10, 0)
	fPrice, _ := strconv.ParseFloat(price, 64)
	iSales, _ := strconv.ParseInt(sales, 10, 0)
	iStock, _ := strconv.ParseInt(stock, 10, 0)

	book := &model.Book{ID: int(ibookID), Title: title, Author: author, Price: fPrice, Sales: int(iSales), Stock: int(iStock), ImagePath: "/static/img/default.jpg"}
	if book.ID > 0 {
		//更i性能图书
		dao.UpdateBook(book)
	} else {
		//添加图书
		dao.AddBook(book)
	}

	//调用GetBooks处理器函数再查询一次数据库
	GetBooks(w, r)
}
