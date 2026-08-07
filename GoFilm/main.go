package main

import (
	"GoFilm/controller"
	"net/http"
)

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {
	//设置处理静态资源
	//直接去页面
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("views/static/"))))
	http.Handle("/pages/", http.StripPrefix("/pages/", http.FileServer(http.Dir("views/pages/"))))
	http.HandleFunc("/main", controller.FirstPage)

	//去登陆
	http.HandleFunc("/login", controller.Login)

	//去注册
	http.HandleFunc("/regist", controller.Regist)

	//通过Ajax请求验证用户名是否可用
	http.HandleFunc("/checkUserName", controller.CheckUserName)

	//用户注销
	http.HandleFunc("/logout", controller.Logout)

	//进入管理员首页
	http.HandleFunc("/manager", controller.FirstPageManager)

	//添加电影
	http.HandleFunc("/addmovie", controller.AddMovieHandler)

	//添加影院
	http.HandleFunc("/addcinema", controller.AddCinemaHandler)

	//进入查看影院界面
	http.HandleFunc("/cinemalist", controller.CinemaHandler)

	//去电影更新界面
	http.HandleFunc("/tomovieupdate", controller.ToMovieUpdateHandler)

	//电影删除
	http.HandleFunc("/deletemovie", controller.DeleteMovieHandler)

	//电影信息更新
	http.HandleFunc("/movieupdate", controller.MovieUpdateHandler)

	//电影院详细信息展示
	http.HandleFunc("/cinemainfo", controller.CinemaInfoHandler)

	//影院删除
	http.HandleFunc("/deletecinema", controller.DeleteCinemaHandler)

	//添加影厅
	http.HandleFunc("/addhall", controller.AddHallHandler)

	//去影厅更新界面
	http.HandleFunc("/toupdatehall", controller.GetHallJSON)

	//影厅更新
	http.HandleFunc("/updatehall", controller.UpdateHallHandler)

	//影厅删除
	http.HandleFunc("/deletehall", controller.DeleteHallHandler)

	//影厅编辑,json接口
	http.HandleFunc("/api/hall/get", controller.GetHallJSON)

	http.ListenAndServe(":8080", nil)
}
