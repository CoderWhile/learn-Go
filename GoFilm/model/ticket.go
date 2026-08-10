package model

import "time"

const (
	TicketStatusPaid   = "paid"
	TicketStatusLocked = "locked"
)

type Ticket struct {
	ID         int
	ShowtimeID int
	UserID     int
	SeatID     int64
	Status     string
	Price      float64
	LockTime   time.Time
	CreatedAt  time.Time
}
