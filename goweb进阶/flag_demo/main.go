package main

import (
	"fmt"
	"os"
)

// os.Args demo
func main() {
	//os.Args是一个[]string
	fmt.Println(os.Args)
	if len(os.Args) > 0 {
		for index, arg := range os.Args {
			fmt.Printf("args[%d]=%v\n", index, arg)
		}
	}
}
