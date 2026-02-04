package main

import (
	"log"

	"shortvideo/internal/danmu/handler"
	danmu "shortvideo/kitex_gen/danmu/danmuservice"
)

func main() {
	svr := danmu.NewServer(handler.NewDanmuService())

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
