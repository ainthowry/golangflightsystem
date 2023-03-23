package main

import (
	"fmt"
	"log"

	// shutdown "golangflightsystem/pkg"
	// storage "golangflightsystem/internal"
	"goflysys/internal/server"
)

func main() {
	port := ":8888"
	newServer := server.NewUDPServer(port)

	fmt.Printf("Server started on port %s\n", port)

	go func() {
		for msg := range newServer.MsgChannel {
			fmt.Printf("[%s] Request for: %s\n", msg.Sender, msg.Payload)
		}
	}()

	log.Fatal(newServer.Start())
}
