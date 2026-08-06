package model

type PageCinemaInfo struct {
	Cinema    *Cinema
	Halls     []*Hall //该电影院的影厅
	Showtimes []*Showtime
}
