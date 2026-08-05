package model

type Page struct {
	Movies     []*Movie // 电影列表
	IsLogin    bool     // 是否已登录
	Username   string   // 用户名
	Identity   string   // 角色: customer / admin
	Keyword    string   // 搜索关键词
	Region     string   // 地区筛选
	CategoryID int      // 当前分类ID
	TagID      int      // 当前标签ID
	Categories []string // 分类列表
	Tags       []string // 标签列表
	// 分页相关 (保留以备后用)
	PageNo      int
	TotalPageNo int
	TotalRecord int
	IsHasPrev   bool
	IsHasNext   bool
}
