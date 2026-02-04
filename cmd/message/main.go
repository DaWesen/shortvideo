package main

import (
	"log"

	"shortvideo/internal/message/handler"
	message "shortvideo/kitex_gen/message/messageservice"
)

func main() {
	svr := message.NewServer(handler.NewMessageService())

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
