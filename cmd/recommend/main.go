package main

import (
	"log"

	"shortvideo/internal/recommend/handler"
	recommend "shortvideo/kitex_gen/recommend/recommendservice"
)

func main() {
	svr := recommend.NewServer(handler.NewRecommendService())

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
