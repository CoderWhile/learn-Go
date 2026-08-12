package model

type MovieDetail struct {
	IsLogin        bool
	Username       string
	Movie          *Movie
	Isdelist       bool
	ShowtimeGroups []*ShowtimeGroup
	MyScore        int
}
type ShowtimeGroup struct {
	CinemaName string
	CinemaAddr string

	Showtimes []*Showtime
}
