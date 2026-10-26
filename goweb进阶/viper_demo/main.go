package main

import (
	"github.com/spf13/viper"
)

func main() {
	viper.SetDefault("fileDir", "./")
}
