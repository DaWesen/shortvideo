package main

import (
	"log"

	"shortvideo/internal/social/handler"
	social "shortvideo/kitex_gen/social/socialservice"
)

func main() {
	svr := social.NewServer(handler.NewSocialService())

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
