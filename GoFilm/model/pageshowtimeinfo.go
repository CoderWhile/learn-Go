package model

type PageShowtimeInfo struct {
	//所有上映中的电影
	movies []*Movie
	//本影院的影厅
	halls []*Hall
}
