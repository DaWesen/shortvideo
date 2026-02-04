package main

import (
	"log"

	"shortvideo/internal/live/handler"
	live "shortvideo/kitex_gen/live/liveservice"
)

func main() {
	svr := live.NewServer(handler.NewLiveService())

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
