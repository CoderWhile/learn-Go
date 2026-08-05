package model

// SeatType 座位类型
type SeatType uint8

const (
	SeatTypeRegular SeatType = iota // 普通座

	SeatTypeCouple // 情侣座

)

// SeatStatus 座位状态
type SeatStatus uint8

const (
	SeatStatusAvailable SeatStatus = iota // 可用
	SeatStatusDisabled                    //维修
)

// 座位结构体
type Seat struct {
	ID       int64
	HallID   string
	Row      string
	Col      int
	SeatType SeatType
	Status   SeatStatus
}
