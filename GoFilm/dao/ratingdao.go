package dao

import "GoFilm/utils"

// 新增评分
func UpsertRating(userID, movieID, score int) error {
	sql := `INSERT INTO ratings (user_id, movie_id, score) VALUES (?,?,?)
            ON DUPLICATE KEY UPDATE score=?`
	_, err := utils.Db.Exec(sql, userID, movieID, score, score)
	return err
}

// 获取某电影的所有评分
func GetMovieScores(movieID int) []int {
	rows, err := utils.Db.Query("SELECT score FROM ratings WHERE movie_id=?", movieID)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var scores []int
	for rows.Next() {
		var s int
		rows.Scan(&s)
		scores = append(scores, s)
	}
	return scores
}

// UpdateMovieRating 更新电影的平均分和评分人数
func UpdateMovieRating(movieID int, avg float64, count int) error {
	_, err := utils.Db.Exec("UPDATE movies SET rating=?, rating_count=? WHERE id=?", avg, count, movieID)
	return err
}

// GetUserRating 获取某用户对某电影的评分
func GetUserRating(userID, movieID int) int {
	var score int
	err := utils.Db.QueryRow(
		"SELECT score FROM ratings WHERE user_id=? AND movie_id=?", userID, movieID,
	).Scan(&score)
	if err != nil {
		return 0 // 未评
	}
	return score
}
