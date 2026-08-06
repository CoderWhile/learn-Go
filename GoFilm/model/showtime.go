package model

type Showtime struct {
	ID        int
	MovieID   int    //这场电影的id
	StartTime string //电影开始时间
	HallID    string //影厅Id
	Hall      *Hall
	CinemaID  string //电影院id
	Status    string //场次的状态：预售/已放映
	Price     float64
}
