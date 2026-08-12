package controller

import (
	"GoFilm/dao"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
)

// 提交评分
func SubmitRatingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	_, sess := dao.IsLogin(r)
	if sess == nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "请先登录"})
		return
	}

	movieID, _ := strconv.Atoi(r.PostFormValue("movieId"))
	score, _ := strconv.Atoi(r.PostFormValue("score"))

	if score < 1 || score > 10 {
		json.NewEncoder(w).Encode(map[string]string{"error": "评分范围 1-10"})
		return
	}
	if err := dao.UpsertRating(sess.UserID, movieID, score); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "评分失败"})
		return
	}

	// 计算平均分
	scores := dao.GetMovieScores(movieID)
	count := len(scores)
	if count == 0 {
		dao.UpdateMovieRating(movieID, 0, 0)
	} else {
		sum := 0
		for _, s := range scores {
			sum += s
		}
		avg := math.Round(float64(sum)/float64(count)*10) / 10
		dao.UpdateMovieRating(movieID, avg, count)
	}

	// 返回新数据
	movie, _ := dao.GetMovieById(movieID)
	if movie == nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "服务器错误"})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"newRating": movie.Rating,
		"count":     movie.RatingCount,
	})
}
