package model

type TicketPage struct {
	Showtime    *Showtime
	Hall        *Hall
	Seats       []*Seat
	SoldSeatIDs map[int64]bool
	IsLogin     bool
	Username    string
}
