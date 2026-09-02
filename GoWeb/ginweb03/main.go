package main

import (
	"fmt"
	"html/template"
	"net/http"
)

type User struct {
	Name   string
	Gender string
	Age    int
}

func sayHello(w http.ResponseWriter, r *http.Request) {
	//定义模板,解析模板,渲染模板
	t, err := template.ParseFiles("./hello.html")
	if err != nil {
		fmt.Println("template parse error", err)
		return
	}
	u1 := User{
		Name:   "Jack",
		Gender: "female",
		Age:    18,
	}
	m1 := map[string]interface{}{
		"Name":   "小龙女",
		"Gender": "女",
		"Age":    17,
	}
	t.Execute(w, map[string]interface{}{
		"u1": u1,
		"m1": m1,
	})
}

// 自定义模板函数
func f1(w http.ResponseWriter, r *http.Request) {
	//定义一个函数:
	kua := func(name string) (string, error) {
		return name + "夯爆了", nil
	}

	t := template.New("hello.html")
	//告诉模板引擎,现在多了一个自定义函数
	t.Funcs(template.FuncMap{
		"kua": kua,
	})
	_, err := t.ParseFiles("./hello.html")
	if err != nil {
		fmt.Println("template parse error", err)
		return
	}

	name := "小王子"
	t.Execute(w, name)
}

func main() {
	http.HandleFunc("/", f1)
	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
