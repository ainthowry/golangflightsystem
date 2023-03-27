package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"

	// shutdown "golangflightsystem/pkg"
	// storage "golangflightsystem/internal"
	"goflysys/internal/api"
	"goflysys/internal/server"
	"goflysys/pkg/marshal"
	"goflysys/pkg/shutdown"
)

func main() {
	//init the storage
	db, err := api.NewDatabase(10 * time.Second)
	if err != nil {
		log.Fatal(err)
	}

	//build server
	port := ":8888"
	newServer := server.NewUDPServer(port)

	//build router
	router := api.NewFlightsRouter()

	//add handlers
	router.HandleFunc(uint32(1), api.GetFlightsHandler)
	router.HandleFunc(uint32(2), api.GetFlightByIdHandler)
	router.HandleFunc(uint32(3), api.ReserveFlightHandler)
	router.HandleFunc(uint32(4), api.SubscribeFlightByIdHandler)

	fmt.Printf("Server started on port %s\n", port)

	go func() {
		for msg := range newServer.MsgChannel {
			fmt.Printf("[%s] Request for: %s\n", msg.Sender, msg.Payload)
			// reqId := marshal.UnmarshalUint32(msg.Payload[:4])
			path := marshal.UnmarshalUint32(msg.Payload[4:8])
			fmt.Printf("Function %d chosen \n", path)
			handler, ok := router.Routes[path]
			sendAddr, err := net.ResolveUDPAddr("udp", msg.Sender)
			if err != nil {
				log.Fatal(err)
			}

			if !ok {
				fmt.Println("function cannot be handled")
				resp := bytes.Join([][]byte{msg.Payload[:4], marshal.MarshalString("BadRequestException\n")}, []byte{})
				newServer.Ln.WriteToUDP(resp, sendAddr)
			} else {
				resp := msg.Payload[:4]
				handler(resp, msg.Payload[8:], db, msg.Sender)
				newServer.Ln.WriteToUDP(resp, sendAddr)
			}
		}
	}()

	log.Fatal(newServer.Start())

	shutdown.Gracefully()
}
