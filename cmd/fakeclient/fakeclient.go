package main

import (
	"goflysys/internal/server"
	"log"
)

func main() {
	//build server
	port := ":62451"
	newServer := server.NewUDPServer(port)
	go func() {
		for msg := range newServer.MsgChannel {
			log.Println(msg.Sender)
			log.Println(msg.Payload)
		}
	}()
	log.Fatal(newServer.Start())
}
