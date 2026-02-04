package main

import (
	"log"

	"shortvideo/internal/interaction/handler"
	interaction "shortvideo/kitex_gen/interaction/interactionservice"
)

func main() {
	svr := interaction.NewServer(handler.NewInteractionService())

	err := svr.Run()

	if err != nil {
		log.Println(err.Error())
	}
}
