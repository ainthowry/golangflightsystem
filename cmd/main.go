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
	"goflysys/pkg/cache"
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

	//build reqsponse cache
	resCache := make(map[cache.UserRequest]cache.ResponseData, 100)

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
			req := cache.UserRequest{ReqId: reqId, Sender: msg.Sender}
			cachedRes, reqExists := resCache[req]
			resp := []byte{}
			if reqExists {
				resp = cachedRes.Response
			} else {
				path := marshal.UnmarshalUint32(msg.Payload[4:8])
				fmt.Println("Intercepted ", msg.Payload)
				fmt.Printf("[%s] Request #%d for function %d chosen with: %s\n", msg.Sender, reqId, path, msg.Payload)
				if path == uint32(7) {
					fmt.Printf("Clearing cache for user %s\n", msg.Sender)
					//reset cache for the user
					for userReq, _ := range resCache {
						if userReq.Sender == msg.Sender {
							delete(resCache, userReq)
						}
					}
					resp = bytes.Join([][]byte{msg.Payload[:4], marshal.MarshalUint32(uint32(1))}, []byte{})
				} else {
					handler, ok := router.Routes[path]

					if !ok {
						fmt.Println("function cannot be handled")
						resp = bytes.Join([][]byte{msg.Payload[:4], marshal.MarshalString("BadRequestException\n")}, []byte{})
					} else {
						resp_pointer := msg.Payload[:4]
						resp = handler(&resp_pointer, msg.Payload[8:], db, msg.Sender)
						resCache[req] = cache.ResponseData{Response: resp}
					}
				}
			}
			sendAddr, err := net.ResolveUDPAddr("udp", msg.Sender)
			if err != nil {
				log.Print(err)
				resp = bytes.Join([][]byte{msg.Payload[:4], marshal.MarshalString("BadRequestException\n")}, []byte{})
			}
			fmt.Println("Sending", resp)
			newServer.Ln.WriteToUDP(resp, sendAddr)
		}
	}()

	log.Fatal(newServer.Start())

	shutdown.Gracefully()
}

func resetCache() {

}
