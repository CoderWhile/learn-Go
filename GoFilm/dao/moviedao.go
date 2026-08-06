package dao

import (
	"GoFilm/model"
	"GoFilm/utils"
	"fmt"
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

// 向数据库中添加一部电影
func AddMovie(movie *model.Movie) error {
	sqlStr := "insert into movies(title,genre,area,intro,imagePath,rating,status,duration) values(?,?,?,?,?,?,?,?)"
	_, err := utils.Db.Exec(sqlStr, movie.Title, movie.Genre, movie.Area, movie.Intro, movie.ImagePath, movie.Rating, movie.Status, movie.Duration)

	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// 更新电影信息
func UpdateMovie(movie *model.Movie) error {
	sqlStr := `update movies set title=?,genre=? ,intro=? , imagePath=? ,rating=? ,status=? , duration=?  where id=?`

	_, err := utils.Db.Exec(sqlStr, movie.Title, movie.Genre, movie.Intro, movie.ImagePath, movie.Rating, movie.Status, movie.Duration, movie.ID)
	if err != nil {
		return err
	}
	return nil
}

// 查询电影数量
func GetMovieCount() int {
	sqlStr := `select count(*) from movies`
	utils.Db.Query(sqlStr)
	var count int
	utils.Db.QueryRow(sqlStr).Scan(&count)
	return count
}

// 根据电影id查询电影
func GetMovieById(id int) (*model.Movie, error) {
	sqlStr := `select title,genre,area,intro,imagePath,rating,status,duration from movies where id=?`
	movie := &model.Movie{}
	row := utils.Db.QueryRow(sqlStr, id)
	movie.ID = id
	row.Scan(&movie.Title, &movie.Genre, &movie.Area, &movie.Intro, &movie.ImagePath, &movie.Rating, &movie.Status, &movie.Duration)

	return movie, nil
}

// 根据电影id删除电影
func DeleteMovieById(id int) error {
	sqlStr := `delete from movies where id=?`
	_, err := utils.Db.Exec(sqlStr, id)
	if err != nil {
		return err
	}
	return nil
}
