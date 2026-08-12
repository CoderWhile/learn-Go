package dao

import (
	"GoFilm/model"
	"GoFilm/utils"
	"fmt"
	"strings"
)

// 按照关键词获取电影
func GetMoviesByWord(word string) ([]*model.Movie, error) {
	//处理用户输入

	kw := strings.NewReplacer("%", "\\%", "_", "\\_").Replace(strings.TrimSpace(word))
	sql := `select id,title,genre,area,intro,imagePath,rating,duration,status
			from movies
			where title like ?
			order by
			    case
			    	when title=? then 0
			    	when title like concat(?,'%') then 1
			    	else 2
				end,
				rating desc
			LIMIT 20`
	likePattern := `%` + kw + `%`
	rows, err := utils.Db.Query(sql, likePattern, word, word)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	movies := []*model.Movie{}
	for rows.Next() {
		movie := &model.Movie{}
		err := rows.Scan(&movie.ID, &movie.Title, &movie.Genre, &movie.Area, &movie.Intro, &movie.ImagePath, &movie.Rating, &movie.Duration, &movie.Status)
		if err != nil {
			return nil, err
		}
		movies = append(movies, movie)
	}
	return movies, nil
}

// //按照标签和地区获取电影
func GetMoviesByReigonAndTag(reigon string, tag string) ([]*model.Movie, error) {
	if reigon != "" && tag != "" {
		fmt.Println("分类和地区")
		sql := `select id,title,genre,area,intro,imagePath,rating,duration,status from movies where genre like concat(?,'%') and area=?`
		rows, err := utils.Db.Query(sql, tag, reigon)
		if err != nil {
			return nil, err
		}
		var movies []*model.Movie
		for rows.Next() {
			var movie model.Movie
			rows.Scan(&movie.ID, &movie.Title, &movie.Genre, &movie.Area, &movie.Intro, &movie.ImagePath, &movie.Rating, &movie.Duration, &movie.Status)
			movies = append(movies, &movie)
		}
		return movies, nil
	} else if reigon == "" && tag != "" {
		fmt.Println("分类 tag:", tag)
		sql := `select id,title,genre,area,intro,imagePath,rating,duration,status from movies where genre like concat(?,'%')`
		rows, err := utils.Db.Query(sql, tag)
		if err != nil {
			return nil, err
		}
		var movies []*model.Movie
		for rows.Next() {
			var movie model.Movie
			rows.Scan(&movie.ID, &movie.Title, &movie.Genre, &movie.Area, &movie.Intro, &movie.ImagePath, &movie.Rating, &movie.Duration, &movie.Status)
			movies = append(movies, &movie)
		}
		return movies, nil
	} else if tag == "" && reigon != "" {
		fmt.Println("单独查地区")
		sql := `select id,title,genre,area,intro,imagePath,rating,duration,status from movies where area=?`
		rows, err := utils.Db.Query(sql, reigon)
		if err != nil {
			return nil, err
		}
		var movies []*model.Movie
		for rows.Next() {
			var movie model.Movie
			rows.Scan(&movie.ID, &movie.Title, &movie.Genre, &movie.Area, &movie.Intro, &movie.ImagePath, &movie.Rating, &movie.Duration, &movie.Status)
			movies = append(movies, &movie)
		}
		return movies, nil
	} else {
		fmt.Println("all")
		sql := `select id,title,genre,area,intro,imagePath,rating,duration,status from movies`
		rows, err := utils.Db.Query(sql)
		if err != nil {
			return nil, err
		}
		var movies []*model.Movie
		for rows.Next() {
			var movie model.Movie
			rows.Scan(&movie.ID, &movie.Title, &movie.Genre, &movie.Area, &movie.Intro, &movie.ImagePath, &movie.Rating, &movie.Duration, &movie.Status)
			movies = append(movies, &movie)
		}
		return movies, nil
	}

}

// 获取所有电影
func GetMovies() ([]*model.Movie, error) {
	sqlStr := `select id,title,genre,area,intro,imagePath,rating from movies`
	utils.Db.Query(sqlStr)
	rows, err := utils.Db.Query(sqlStr)
	if err != nil {
		return nil, err
	}
	var movies []*model.Movie
	for rows.Next() {
		var movie model.Movie
		rows.Scan(&movie.ID, &movie.Title, &movie.Genre, &movie.Area, &movie.Intro, &movie.ImagePath, &movie.Rating)
		movies = append(movies, &movie)
	}
	return movies, nil
}

// 获取前十条票房最高的电影
func GetMovieByBoxoffice() ([]*model.Movie, error) {
	sql := `select id,title,genre,area,intro,imagePath,rating,status,duration,boxoffice from movies order by boxoffice desc limit 0,10`
	rows, err := utils.Db.Query(sql)
	if err != nil {
		return nil, err
	}
	var movies []*model.Movie
	for rows.Next() {
		movie := &model.Movie{}
		rows.Scan(&movie.ID, &movie.Title, &movie.Genre, &movie.Area, &movie.Intro, &movie.ImagePath, &movie.Rating, &movie.Status, &movie.Duration, &movie.BoxOffice)
		movies = append(movies, movie)
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
	sqlStr := `select title,genre,area,intro,imagePath,rating,status,duration,boxoffice from movies where id=?`
	movie := &model.Movie{}
	row := utils.Db.QueryRow(sqlStr, id)
	movie.ID = id
	row.Scan(&movie.Title, &movie.Genre, &movie.Area, &movie.Intro, &movie.ImagePath, &movie.Rating, &movie.Status, &movie.Duration, &movie.BoxOffice)

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

// 更改电影票房数量
func UpdateBoxOffice(boxoffice float64, movieid int) error {
	sql := `update movies set boxoffice=boxoffice+? where id=?`
	_, err := utils.Db.Exec(sql, boxoffice, movieid)
	return err

}
