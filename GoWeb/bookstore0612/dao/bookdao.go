package dao

import (
	"GoWeb/bookstore0612/model"
	"GoWeb/bookstore0612/utils"
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
