package controller

import (
	"GoFilm/dao"
	"GoFilm/model"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 格式化时间差
func timeAgo(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	diff := time.Since(t)
	minutes := int(math.Abs(diff.Minutes()))
	switch {
	case minutes < 1:
		return "刚刚"
	case minutes < 60:
		return fmt.Sprintf("%d分钟前", minutes)
	case minutes < 1440:
		return fmt.Sprintf("%d小时前", minutes/60)
	default:
		return fmt.Sprintf("%d天前", minutes/1440)
	}
}

// 建树用一个roots数组存根节点,有父结点的就放在父结点的Children节点数组里
func buildTree(list []*model.Comment) []*model.CommentNode {
	tree := make(map[int]*model.CommentNode)
	var roots []*model.CommentNode

	for _, v := range list {
		node := &model.CommentNode{
			ID:        v.ID,
			UserName:  v.UserName,
			Content:   v.Content,
			CreatedAt: timeAgo(v.CreatedAt),
			Children:  []*model.CommentNode{},
		}
		tree[v.ID] = node
	}
	for _, v := range list {
		if v.ParentID == 0 {
			roots = append(roots, tree[v.ID])
		} else {
			parent := tree[v.ParentID]
			if parent != nil {
				parent.Children = append(parent.Children, tree[v.ID])
			}
		}
	}
	if roots == nil {
		roots = []*model.CommentNode{} // 保证返回 [], 不是 null
	}
	return roots
}

// 获得所有评论
func GetCommentTree(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	movieID, _ := strconv.Atoi(r.URL.Query().Get("movieId"))
	comments, _ := dao.GetCommentsByMovieId(movieID)
	//for _, comment := range comments {
	//	fmt.Printf("%+v\n", comment)
	//}
	//fmt.Println("获得所有评论")
	roots := buildTree(comments)
	json.NewEncoder(w).Encode(roots)
}

// 添加评论
func AddComment(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	flag, sess := dao.IsLogin(r)
	if !flag {
		json.NewEncoder(w).Encode(map[string]string{"error": "请先登录"})
		return
	}
	movieID, _ := strconv.Atoi(r.PostFormValue("movieId"))
	parentID, _ := strconv.Atoi(r.PostFormValue("parentId"))
	content := r.PostFormValue("content")
	if strings.TrimSpace(content) == "" {
		json.NewEncoder(w).Encode(map[string]string{"error": "内容不能为空"})
		return
	}
	comment := &model.Comment{
		MovieID:   movieID,
		ParentID:  parentID,
		Content:   content,
		UserName:  sess.UserName,
		UserID:    sess.UserID,
		CreatedAt: time.Now(),
	}
	newcom, _ := dao.AddComment(comment)
	newcom.CreatedAt = time.Now()
	fmt.Printf("%+v\n", newcom)
	json.NewEncoder(w).Encode(newcom)
}
