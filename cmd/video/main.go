package main

import (
	"log"

	"shortvideo/internal/video/handler"
	video "shortvideo/kitex_gen/video/videoservice"
)

func main() {
	svr := video.NewServer(handler.NewVideoService())

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
