package dao

import (
	"GoFilm/model"
	"GoFilm/utils"
)

// 获取所有电影
func GetMovies() ([]*model.Movie, error) {
	sqlStr := `select id,title,genre,area,intro,imagePath from movies`
	utils.Db.Query(sqlStr)
	rows, err := utils.Db.Query(sqlStr)
	if err != nil {
		return nil, err
	}
	var movies []*model.Movie
	for rows.Next() {
		var movie model.Movie
		rows.Scan(&movie.ID, &movie.Title, &movie.Genre, &movie.Area, &movie.Intro, &movie.ImagePath)
		movies = append(movies, &movie)
	}
	return movies, nil
}
