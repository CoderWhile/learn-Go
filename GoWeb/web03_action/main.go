package main

import (
	"GoWeb/web03_action/model"
	"html/template"
	"net/http"
)

func testIf(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("index.html"))

	age := 17

	t.Execute(w, age > 18)
}

// 测试Range
func testRange(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("range.html"))
	var emps []*model.Employee
	//emps := make([]*model.Employee, 0)
	emp := &model.Employee{
		ID:       1,
		LastName: "王",
		Email:    "2398@sdf",
	}
	emps = append(emps, emp)
	emp2 := &model.Employee{
		ID:       2,
		LastName: "白",
		Email:    "3928@sdf",
	}
	emps = append(emps, emp2)
	emp3 := &model.Employee{
		ID:       3,
		LastName: "流",
		Email:    "4323@sdf",
	}
	emps = append(emps, emp3)
	t.Execute(w, emps)
}

func testWith(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("with.html"))

	t.Execute(w, "狸猫")
}
func testTemplate(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("hello.html", "hello2.html"))

	t.Execute(w, "XXX")
}

func testDefine(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.ParseFiles("define.html"))

	t.ExecuteTemplate(w, "model", "")
}

func main() {
	http.HandleFunc("/testIf", testIf)
	http.HandleFunc("/testRange", testRange)
	http.HandleFunc("/testWith", testWith)
	http.HandleFunc("/testTemplate", testTemplate)
	http.HandleFunc("/testDefine", testDefine)

	http.ListenAndServe(":8080", nil)

}
