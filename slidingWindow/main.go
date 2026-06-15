package main

import "fmt"

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {
	arrivals := make([]int, 0, 20)
	arrivals = append(arrivals, 7, 3, 9, 9, 7, 3, 5, 9, 7, 2, 6, 10, 9, 7, 9, 1, 3, 6, 2, 4, 6, 2, 6, 8, 4, 8, 2, 7, 5, 6)
	a := minArrivalsToDiscard(arrivals, 10, 1)
	fmt.Println(a)
}
func minArrivalsToDiscard(arrivals []int, w int, m int) int {
	map1 := make(map[int]int)
	rea := 0
	for v, value := range arrivals {
		map1[value]++
		if map1[value] > m {
			re++
			map1[value]--
		}
		left := v - w + 1
		if left >= 0 {
			map1[arrivals[left]]--
		}
	}
	return rea
}
