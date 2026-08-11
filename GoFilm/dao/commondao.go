package dao

import (
	"GoFilm/model"
	"GoFilm/utils"
)

// 获取某电影的所有评论（含用户名）
func GetCommentsByMovieId(movieID int) ([]*model.Comment, error) {
	sql := `SELECT c.id, c.movie_id, c.user_id, u.username, c.parent_id, c.content, c.create_at
	        FROM comments c
	        JOIN users u ON c.user_id = u.id
	        WHERE c.movie_id = ?
	        ORDER BY c.create_at ASC`
	rows, err := utils.Db.Query(sql, movieID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.Comment
	for rows.Next() {
		r := &model.Comment{}
		rows.Scan(&r.ID, &r.MovieID, &r.UserID, &r.UserName, &r.ParentID, &r.Content, &r.CreatedAt)
		list = append(list, r)
	}
	return list, nil
}

// 添加评论（回填自增 ID）
func AddComment(com *model.Comment) (*model.Comment, error) {
	sql := `INSERT INTO comments (movie_id, user_id, parent_id, content, create_at) VALUES (?,?,?,?,?)`
	result, err := utils.Db.Exec(sql, com.MovieID, com.UserID, com.ParentID, com.Content, com.CreatedAt)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	com.ID = int(id)
	return com, nil
}
