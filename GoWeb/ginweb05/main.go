package main //TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {
	r := gin.Default()
	r.GET("/json", func(c *gin.Context) {
		//方法1:使用map
		//data := map[string]interface{}{
		//	"name":    "小王子",
		//	"message": "hello go",
		//	"age":     12,
		//}

		data := gin.H{
			"name":    "小王子",
			"message": "hello go",
			"age":     12,
		}

		c.JSON(http.StatusOK, data)
	})

	//方法2结构体
	type msg struct {
		Name string `json:"name"`
		Age  int    `json:"int"`
	}
	r.GET("/another_json", func(c *gin.Context) {
		data := msg{
			"jack",
			12,
		}
		c.JSON(http.StatusOK, data)
	})
	r.Run(":8080")
}
