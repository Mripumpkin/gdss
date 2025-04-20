package main

import (
	"fmt"

	"github.com/jekki/gdss/config"
	"github.com/jekki/gdss/log"
)

func main() {
	conf, err := config.LoadConfigProvider()
	if err == nil {
		fmt.Println("config init failed!")
		return
	}
	log := log.NewLogger(conf)
	log.Info("success")
	// return
}
