package main

import (
	"GoFilm/controller"
	"context"
	"log"
	"net/http"
	"sync"
	"time"
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

	//去添加场次界面
	http.HandleFunc("/toaddshowtime", controller.ToAddShowTimeHandler)

	//添加场次
	http.HandleFunc("/addshowtime", controller.AddShowTimeHandler)

	//场次更新
	http.HandleFunc("/updateshowtime", controller.UpdateShowTimeHandler)

	//场次删除
	http.HandleFunc("/deleteshowtime", controller.DeleteShowTimeHandler)

	//用户进入电影的详情页面
	http.HandleFunc("/movie/detail", controller.MovieDetailHandler)

	//用户端影院列表
	http.HandleFunc("/cinema/list", controller.UserCinemaListHandler) // 用户端影院列表

	//影院排片页
	http.HandleFunc("/cinema/shows", controller.CinemaShowsHandler) // 影院排片页

	////用户选购票页
	//http.HandleFunc("/ticket/select", controller.TicketSelectHandler) // 选座购票页

	// 购票页 Ajax 数据
	http.HandleFunc("/api/ticket/showtime", controller.GetTicketDataJSON)

	//购买票
	http.HandleFunc("/ticket/buy", controller.BuyTicketHandler)
	//锁定座位
	http.HandleFunc("/ticket/lock", controller.LockSeatsHandler)
	//释放锁定
	http.HandleFunc("/ticket/unlock", controller.UnlockSeatsHandler)

	//用户的订单
	http.HandleFunc("/getMyOrders", controller.UserOrdersHandler)
	//退票
	http.HandleFunc("/ticket/refund", controller.RefundTicketHandler)

	//关键词搜索电影（待实现）
	// http.HandleFunc("/Search", controller.SearchHandler)

	//添加评论
	http.HandleFunc("/addcomment", controller.AddComment)

	//获取某电影的所有评论
	http.HandleFunc("/getcomment", controller.GetCommentTree)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go StartShowtimeChecker(ctx)

	http.ListenAndServe(":8080", nil)
}
func StartShowtimeChecker(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// ★ 防重叠：上一轮没跑完就跳过
	var mu sync.Mutex

	// 启动时立即执行一次
	go func() {
		if mu.TryLock() {
			controller.CheckStatusShowtime(ctx)
			mu.Unlock()
		}
	}()

	for {
		select {
		case <-ticker.C:
			if !mu.TryLock() {
				log.Println("上一轮场次检查未完成，跳过本轮")
				continue
			}
			go func() {
				defer mu.Unlock()
				controller.CheckStatusShowtime(ctx)
			}()
		case <-ctx.Done():
			//被取消
			return
		}
	}
}
