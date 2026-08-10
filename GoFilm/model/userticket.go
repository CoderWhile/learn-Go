package model

type UserTicket struct {
	UserId   int
	Username string
	Tickets  []*Ticket
}
