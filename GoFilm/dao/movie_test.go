package dao

import (
	"fmt"
	"testing"
)

func TestGetMovies(t *testing.T) {
	fmt.Println("GetMovies")
	t.Run("GetMovies", testGetMovies)

}
func testGetMovies(t *testing.T) {
	movies, err := GetMovies()
	if err != nil {
		fmt.Println(err)
	}

	for _, movie := range movies {
		fmt.Printf("%+v\n", movie)
	}
}

// 测试添加座位
func TestAddSeats(t *testing.T) {
	fmt.Println("测试根据影厅Id批量添加座位")
	t.Run()
}
func testAddSeats(t *testing.T) {

}
