package model

type MovieDetail struct {
	IsLogin        bool
	Username       string
	Movie          *Movie
	Isdelist       bool
	ShowtimeGroups []*ShowtimeGroup
}
type ShowtimeGroup struct {
	CinemaName string
	CinemaAddr string

	Showtimes []*Showtime
}
