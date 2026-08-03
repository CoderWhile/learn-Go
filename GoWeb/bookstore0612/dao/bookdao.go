package dao

import (
	"GoWeb/bookstore0612/model"
	"GoWeb/bookstore0612/utils"
	"strconv"
)

// 获取所有图书
func GetBooks() ([]*model.Book, error) {
	sqlStr := `select id,title,author,price,sales,stock,img_path from books`
	utils.Db.Query(sqlStr)
	rows, err := utils.Db.Query(sqlStr)
	if err != nil {
		return nil, err
	}
	var books []*model.Book
	for rows.Next() {
		var book model.Book
		rows.Scan(&book.ID, &book.Title, &book.Author, &book.Price, &book.Sales, &book.Stock, &book.ImagePath)
		books = append(books, &book)
	}
	return books, nil
}

// 向数据库中添加一本书
func AddBook(book *model.Book) error {
	sqlStr := "insert into books(title,author,price,sales,stock,img_path) values(?,?,?,?,?,?)"
	_, err := utils.Db.Exec(sqlStr, book.Title, book.Author, book.Price, book.Sales, book.Stock, book.ImagePath)
	if err != nil {
		return err
	}
	return nil

}

// 删除图书,根据图书id
func DeleteBook(bookID string) error {
	sqlStr := "delete from books where id=?"
	_, err := utils.Db.Exec(sqlStr, bookID)
	if err != nil {
		return err
	}
	return nil
}

// 根据屠苏Id从数据库中查询一本图书
func GetBookByID(bookID string) (*model.Book, error) {
	sqlStr := `select id,title,author,price,sales,stock,img_path from books where id=?`
	row := utils.Db.QueryRow(sqlStr, bookID)
	book := &model.Book{}
	row.Scan(&book.ID, &book.Title, &book.Author, &book.Price, &book.Sales, &book.Stock, &book.ImagePath)

	return book, nil

}

// 根据图书Id更新图书id
func UpdateBook(book *model.Book) error {
	sqlStr := `update books set title=? ,author=? ,price=? ,sales=? ,stock=? ,img_path=? where id=?`
	_, err := utils.Db.Exec(sqlStr, book.Title, book.Author, book.Price, book.Sales, book.Stock, book.ImagePath, book.ID)
	if err != nil {
		return err
	}
	return nil

}

// GetPageBooks获取带分页的图书信息
func GetPageBooks(pageNo string) (*model.Page, error) {
	//页码装欢
	iPageNo, _ := strconv.ParseInt(pageNo, 10, 64)
	//获取总数
	sqlStr := `select count(*) from books`
	var totalRecord int64
	row := utils.Db.QueryRow(sqlStr)
	row.Scan(&totalRecord)
	//设置每页显示4条记录
	var pageSize int64
	//获取总页数
	var totalPageNo int64
	pageSize = 4
	if totalRecord%pageSize == 0 {
		totalPageNo = totalRecord / pageSize
	} else {
		totalPageNo = totalRecord/pageSize + 1
	}
	var books []*model.Book
	//获取当前页中的图书
	sqlStr2 := `select id,title,author,price,sales,stock,img_path from books limit ?,? `
	rows, err := utils.Db.Query(sqlStr2, (iPageNo-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		book := &model.Book{}
		rows.Scan(&book.ID, &book.Title, &book.Author, &book.Price, &book.Sales, &book.Stock, &book.ImagePath)
		books = append(books, book)
	}
	//创建page
	page := &model.Page{
		Books:       books,
		PageNo:      iPageNo,
		PageSize:    pageSize,
		TotalPageNo: totalPageNo,
		TotalRecord: totalRecord,
	}
	return page, nil
}

// GetPageBooksByPrice获取带分页和价格的图书信息
func GetPageBooksByPrice(pageNo string, minPrice string, maxPrice string) (*model.Page, error) {
	//页码装欢
	iPageNo, _ := strconv.ParseInt(pageNo, 10, 64)
	//获取总数
	sqlStr := `select count(*) from books where price between ? and ? `
	var totalRecord int64
	row := utils.Db.QueryRow(sqlStr, minPrice, maxPrice)
	row.Scan(&totalRecord)
	//设置每页显示4条记录
	var pageSize int64
	//获取总页数
	var totalPageNo int64
	pageSize = 4
	if totalRecord%pageSize == 0 {
		totalPageNo = totalRecord / pageSize
	} else {
		totalPageNo = totalRecord/pageSize + 1
	}
	var books []*model.Book
	//获取当前页中的图书
	sqlStr2 := `select id,title,author,price,sales,stock,img_path from books where price between ? and ? limit ?,? `
	rows, err := utils.Db.Query(sqlStr2, minPrice, maxPrice, (iPageNo-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		book := &model.Book{}
		rows.Scan(&book.ID, &book.Title, &book.Author, &book.Price, &book.Sales, &book.Stock, &book.ImagePath)
		books = append(books, book)
	}
	//创建page
	page := &model.Page{
		Books:       books,
		PageNo:      iPageNo,
		PageSize:    pageSize,
		TotalPageNo: totalPageNo,
		TotalRecord: totalRecord,
	}
	return page, nil
}
