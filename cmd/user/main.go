package main

import (
	"log"

	"shortvideo/internal/user/handler"
	user "shortvideo/kitex_gen/user/userservice"
)

func main() {
	svr := user.NewServer(handler.NewUserService())

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
