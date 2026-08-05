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
