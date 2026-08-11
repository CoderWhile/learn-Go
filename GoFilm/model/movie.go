package model

type Movie struct {
	ID        int     // id
	Title     string  // 名称
	Genre     string  // 类型
	Area      string  // 地区
	Intro     string  // 简介
	ImagePath string  // 图片保存地址
	Rating    float64 // 评分
	Status    string  //状态  未上映，上映中，下架
	Duration  int     // 时长(分钟)
	Count     int     //票房数量
}
