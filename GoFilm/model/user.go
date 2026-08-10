package model

// User结构体
type User struct {
	ID       int
	Username string
	Password string
	Identity string
}

type CinemaList struct {
	IsLogin          bool
	Username         string
	TotalCinemaCount int
	Cinemas          []*Cinema
}
type CinemaShows struct {
	IsLogin     bool
	Username    string
	Cinema      *Cinema
	MovieGroups []*MovieGroup
}
type MovieGroup struct {
	MovieTitle string
	MovieGenre string
	IsShow     bool
	Showtimes  []*Showtime
}
