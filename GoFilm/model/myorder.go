package model

import "time"

type OrderItem struct {
	TicketID     int
	MovieTitle   string
	MovieImage   string
	CinemaName   string
	HallName     string
	ShowTime     string
	Row          string
	Col          int
	SeatLabel    string // "A3"
	Price        float64
	Status       string // "已放映" / "未放映"
	IsPassed     bool   // 是否已过放映时间
	OrderTime    time.Time
}

type MyOrderPage struct {
	Orders   []OrderItem
	Username string
	IsLogin  bool
}
