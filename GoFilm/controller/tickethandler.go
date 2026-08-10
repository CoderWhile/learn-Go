package controller

import (
	"GoFilm/dao"
	"encoding/json"
	"net/http"
	"strconv"
)

// 用户选购票
func TicketSelectHandler(w http.ResponseWriter, r *http.Request) {

}

func GetTicketDataJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	showtimeid := r.PostFormValue("showtimeId")
	ishowtimeid, err := strconv.Atoi(showtimeid)
	showtime, _ := dao.GetShowtimeById(ishowtimeid)
	hallID := showtime.HallID
	movie, err2 := dao.GetMovieById(showtime.MovieID)
	if err2 != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "电影不存在"})
		return
	}

	if hallID == "" {
		json.NewEncoder(w).Encode(map[string]string{"error": "缺少hallId"})
		return
	}

	//fmt.Println(hallID)
	hall, err := dao.GetHallById(hallID)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "影厅不存在"})
		return
	}

	seats, _ := dao.GetSeatsByHallId(hallID)
	//for _, seat := range seats {
	//	fmt.Printf("%+v\n", seat)
	//}
	//fmt.Printf("%+v\n", seats)
	// 返回前端需要的格式
	type SeatVO struct {
		Row      string `json:"row"`
		Col      int    `json:"col"`
		SeatType int    `json:"seattype"`
		Status   int    `json:"status"`
	}
	seatVOs := make([]SeatVO, len(seats))
	for i, s := range seats {
		seatVOs[i] = SeatVO{
			Row: s.Row, Col: s.Col,
			SeatType: int(s.SeatType), Status: int(s.Status),
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        hall.ID,
		"name":      hall.Name,
		"totalRows": hall.TotalRows,
		"totalCols": hall.TotalCols,
		"seats":     seatVOs,
	})
}
