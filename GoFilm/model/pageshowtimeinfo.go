package model

type PageShowtimeInfo struct {
	// 当前影院ID
	CinemaID string
	// 所有上映中的电影
	Movies []*Movie
	// 本影院的影厅
	Halls []*Hall
}
