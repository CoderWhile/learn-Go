package model

type Page struct {
	Books       []*Book //每页里面的图书存放的切片
	PageNo      int64   //当前页
	PageSize    int64   //每页显示的条数
	TotalPageNo int64   //总页数,
	TotalRecord int64   //总记录数,
	MinPrice    string
	MaxPrice    string
	IsLogin     bool
	Username    string
}

// IsHasPrev 判断是否有上一页
func (p *Page) IsHasPrev() bool {
	return p.PageNo > 1
}

func (p *Page) IsHasNext() bool {
	return p.PageNo < p.TotalPageNo
}

// 获取上一页
func (p *Page) GetPrevPageNo() int64 {
	if p.IsHasPrev() {
		return p.PageNo - 1
	}
	return 1

}

// 获取下一页
func (p *Page) GetNextPageNo() int64 {
	if p.IsHasNext() {
		return p.PageNo + 1
	}
	return p.TotalPageNo

}
