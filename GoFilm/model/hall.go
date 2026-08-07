package model

import "strings"

// 影厅
type Hall struct {
	ID         string
	Name       string
	TotalRows  int
	TotalCols  int
	Seats      []*Seat
	SeatMatrix [][]*Seat // 运行时索引 [row][col] → *Seat
	Version    int64
	CinemaID   string
}

func (h *Hall) BuildMatrix() {
	h.SeatMatrix = make([][]*Seat, h.TotalRows+1)
	for i := range h.SeatMatrix {
		h.SeatMatrix[i] = make([]*Seat, h.TotalCols+1)
	}
	for idx := range h.Seats {
		s := &h.Seats[idx]
		rowIdx := rowToInt((*s).Row)
		h.SeatMatrix[rowIdx][(*s).Col] = *s

	}
}

func (h *Hall) GetSeat(row string, col int) *Seat {
	r := rowToInt(row)
	if r < 1 || r >= h.TotalRows || col < 1 || col >= h.TotalCols {
		return nil
	}
	return h.SeatMatrix[r][col]
}

func rowToInt(row string) int {
	var n int
	for _, c := range strings.ToUpper(row) {
		n = n*26 + int(c-'A'+1)
	}
	return n
}
