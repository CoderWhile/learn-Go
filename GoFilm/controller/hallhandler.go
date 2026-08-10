package controller

import (
	"GoFilm/dao"
	"GoFilm/model"
	"GoFilm/utils"
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
		})
	}
	//将影厅和座位存入数据库
	hall.ID = utils.CreateUUID()
	dao.AddHall(hall.CinemaID, hall)
	//把座位信息存入数据库

	dao.AddSeats(hall.ID, hall.Seats)
	//进入影厅添加成功界面
	t := template.Must(template.ParseFiles("views/pages/hall/hall_add_success.html"))
	t.Execute(w, hall)
}

// 去影厅更新页面
func ToUpdateHallHandler(w http.ResponseWriter, r *http.Request) {
	hallID := r.FormValue("hallId")
	//获取影厅
	hall, _ := dao.GetHallById(hallID)
	//根据影厅id获取座位
	hall.Seats, _ = dao.GetSeatsByHallId(hall.ID)

	t := template.Must(template.ParseFiles("views/pages/hall/hall_edit.html"))
	t.Execute(w, hall)
}

// GetHallJSON 返回影厅 + 座位 JSON（编辑页 Ajax 加载）
func GetHallJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	hallID := r.URL.Query().Get("hallId")
	hallID = r.FormValue("hallId")
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

func UpdateHallHandler(w http.ResponseWriter, r *http.Request) {
	hallID := r.PostFormValue("hallId")
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
		ID:        hallID,
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
		})
	}
	fmt.Printf("%+v\n", hall)
	//将影厅和座位更新数据库
	dao.UpdateHall(hall)

	//把座位信息存入数据库
	//批量更新座位信息
	//按照影厅id更新座位信息
	//根据影厅批量删除座位
	dao.DeleteSeatsByHallId(hall.ID)
	//根据影厅添加新座位
	dao.AddSeats(hall.ID, hall.Seats)
	//进入影厅添加成功界面
	t := template.Must(template.ParseFiles("views/pages/hall/hall_update_success.html"))
	t.Execute(w, hall)
}

// 删除影厅
func DeleteHallHandler(w http.ResponseWriter, r *http.Request) {
	hallID := r.PostFormValue("hallId")
	//先删除座位
	dao.DeleteSeatsByHallId(hallID)
	//在删除影厅
	dao.DeleteHall(hallID)
	//删除该影厅的座位
	dao.DeleteSeatsByHallId(hallID)
	//
}
