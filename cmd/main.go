package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"

	"goflysys/internal/api"
	"goflysys/internal/server"
	"goflysys/pkg/marshal"
	"goflysys/pkg/responsemanager"
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

	//build reqsponse cache
	responseCache := responsemanager.NewResponseManager()

	//build router
	router := api.NewFlightsRouter()

	//add handlers
	router.HandleFunc(uint32(1), api.GetFlightsHandler)
	router.HandleFunc(uint32(2), api.GetFlightByIdHandler)
	router.HandleFunc(uint32(3), api.ReserveFlightHandler)
	router.HandleFunc(uint32(4), api.SubscribeFlightByIdHandler)
	router.HandleFunc(uint32(5), api.GetSeatsByIdHandler)
	router.HandleFunc(uint32(6), api.RefundSeatBySeatNumHandler)

	fmt.Printf("Server started on port %s\n", port)

	go func() {
		for msg := range newServer.MsgChannel {
			reqId := marshal.UnmarshalUint32(msg.Payload[:4])
			resp := make([]byte, 0)

			hashKey := responseCache.GetHashKey(reqId, msg.Sender)
			cachedResponse, err := responseCache.GetCachedResponse(hashKey)
			if err == nil {
				resp = cachedResponse
			} else {
				path := marshal.UnmarshalUint32(msg.Payload[4:8])
				fmt.Printf("[%s] Request #%d for function %d chosen with payload: %s\n", msg.Sender, reqId, path, msg.Payload)
				fmt.Println("Intercepted payload of", msg.Payload)

				handler, ok := router.Routes[path]

				if !ok {
					fmt.Println("function cannot be handled")
					resp = bytes.Join([][]byte{msg.Payload[:4], marshal.MarshalString("BadRequestException\n")}, []byte{})
				} else {
					resp_pointer := msg.Payload[:4]
					resp = handler(&resp_pointer, msg.Payload[8:], db, msg.Sender)
					responseCache.SetCachedResponse(hashKey, resp)
				}
			}

			sendAddr, err := net.ResolveUDPAddr("udp", msg.Sender)
			if err != nil {
				log.Println(err)
				resp = bytes.Join([][]byte{msg.Payload[:4], marshal.MarshalUint32(400)}, []byte{})
			}
			fmt.Println("Sending", resp)
			newServer.Ln.WriteToUDP(resp, sendAddr)
		}
	}()

	log.Fatal(newServer.Start())

	shutdown.Gracefully()
}
