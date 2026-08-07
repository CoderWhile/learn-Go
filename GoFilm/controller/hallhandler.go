package controller

import (
	"GoFilm/dao"
	"GoFilm/model"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
)

func AddHallHandler(w http.ResponseWriter, r *http.Request) {
	cinemaID := r.PostFormValue("cinemaId")
	name := r.PostFormValue("name")
	totalRows, err := strconv.Atoi(r.PostFormValue("totalRows"))
	if err != nil {
		http.Error(w, "排数必须输入合法数字", http.StatusBadRequest)
		return
	}
	// 业务校验：影厅排数不能小于1
	if totalRows <= 0 {
		http.Error(w, "排数必须大于0", http.StatusBadRequest)
		return
	}
	totalCols, err := strconv.Atoi(r.PostFormValue("totalCols"))
	if err != nil {
		http.Error(w, "排数必须输入合法数字", http.StatusBadRequest)
		return
	}
	// 业务校验：影厅排数不能小于1
	if totalRows <= 0 {
		http.Error(w, "排数必须大于0", http.StatusBadRequest)
		return
	}
	seatJSON := r.PostFormValue("seats")
	fmt.Println(seatJSON)
	//座位解析
	type SeatInput struct {
		Row      string `json:"row"`
		Col      int    `json:"col"`
		SeatType uint8  `json:"seatType"`
		Status   uint8  `json:"status"`
	}
	var seats []SeatInput
	err = json.Unmarshal([]byte(seatJSON), &seats)
	if err != nil {
		fmt.Println("作为信息反序列化错误", err)
	}
	hall := &model.Hall{
		CinemaID:  cinemaID,
		Name:      name,
		TotalRows: totalRows,
		TotalCols: totalCols,
	}
	for _, s := range seats {
		hall.Seats = append(hall.Seats, &model.Seat{
			Row:      s.Row,
			Col:      s.Col,
			SeatType: model.SeatType(s.SeatType),
			Status:   model.SeatStatus(s.Status),
			HallID:   hall.ID,
		})
	}
	//将影厅和座位存入数据库
	dao.AddHall(hall.CinemaID, hall)
	//把座位信息存入数据库
	dao.AddSeats(hall.ID, hall.Seats)
	//进入影厅添加成功界面
	t := template.Must(template.ParseFiles("views/pages/hall/hall_add_success.html"))
	t.Execute(w, hall)
}
