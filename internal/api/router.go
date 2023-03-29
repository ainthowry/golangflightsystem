package api

import (
	"bytes"
	"fmt"
	"goflysys/pkg/marshal"
	"log"
	"time"
)

type FlightsRouter struct {
	Routes map[uint32]func(Response *[]byte, Request []byte, fdb *FlightDatabase, user string) []byte
}

func NewFlightsRouter() *FlightsRouter {
	return &FlightsRouter{
		Routes: make(map[uint32]func(Response *[]byte, Request []byte, fdb *FlightDatabase, user string) []byte),
	}
}

func (r *FlightsRouter) HandleFunc(path uint32, handler func(Response *[]byte, Request []byte, fdb *FlightDatabase, user string) []byte) {
	r.Routes[path] = handler
}

func GetFlightsHandler(res_pointer *[]byte, data []byte, fdb *FlightDatabase, user string) []byte {
	res := *res_pointer

	source := marshal.UnmarshalString(data)
	destination := marshal.UnmarshalString(data[4+len(source):])
	flights, err := fdb.GetFlights(source, destination)
	if err != nil {
		log.Printf("[SERVICE ERROR] %v", err)
	}
	var flightids []byte
	count := 0
	for _, flight := range flights {
		count++
		flightids = append(flightids, marshal.MarshalUint32(flight.id)...)
	}

	fmt.Println("flightids:", flightids)
	res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(count)), flightids}, []byte{})
	fmt.Println("processed:", res)
	return res

}

func GetFlightByIdHandler(res_pointer *[]byte, data []byte, fdb *FlightDatabase, user string) []byte {
	res := *res_pointer

	id := marshal.UnmarshalUint32(data[:4])
	flight, err := fdb.GetFlightById(id)
	if err != nil {
		log.Printf("[SERVICE ERROR] %v", err)
	}

	res = bytes.Join([][]byte{res, marshal.MarshalInt64(flight.departureTime.Unix()), marshal.MarshalFloat64(flight.price), marshal.MarshalUint32(flight.seatsLeft)}, []byte{})
	return res
}

func ReserveFlightHandler(res_pointer *[]byte, data []byte, fdb *FlightDatabase, user string) []byte {
	res := *res_pointer

	id := marshal.UnmarshalUint32(data[:4])
	numSeats := marshal.UnmarshalUint32(data[4:8])

	seatsReserved, err := fdb.ReserveFlight(id, numSeats, user)
	if err != nil {
		log.Printf("[SERVICE ERROR] %v", err)
	}

	log.Println(seatsReserved)
	res = bytes.Join([][]byte{res, marshal.MarshalUint32Array(seatsReserved)}, []byte{})
	return res
}

func SubscribeFlightByIdHandler(res_pointer *[]byte, data []byte, fdb *FlightDatabase, user string) []byte {
	res := *res_pointer

	id := marshal.UnmarshalUint32(data[:4])
	endTime := time.Unix(marshal.UnmarshalInt64(data[4:12]), 0)

	_, err := fdb.SubscribeFlightById(id, endTime, user)
	if err != nil {
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(0))}, []byte{})
		log.Printf("[SERVICE ERROR] %v", err)
	}

	res = bytes.Join([][]byte{res, marshal.MarshalUint32(1)}, []byte{})
	return res
}

func GetSeatsByIdHandler(res_pointer *[]byte, data []byte, fdb *FlightDatabase, user string) []byte {
	res := *res_pointer

	id := marshal.UnmarshalUint32(data[:4])

	seatsReserved, err := fdb.GetSeatsById(id, user)
	if err != nil {
		log.Printf("[SERVICE ERROR] %v", err)
	}

	res = bytes.Join([][]byte{res, marshal.MarshalUint32Array(seatsReserved)}, []byte{})
	return res
}

func RefundSeatBySeatNumHandler(res_pointer *[]byte, data []byte, fdb *FlightDatabase, user string) []byte {
	res := *res_pointer

	id := marshal.UnmarshalUint32(data[:4])
	seatNum := marshal.UnmarshalUint32(data[4:8])

	seatsLeft, err := fdb.RefundSeatBySeatNum(id, seatNum, user)
	if err != nil {
		log.Printf("[SERVICE ERROR] %v", err)
	}

	res = bytes.Join([][]byte{res, marshal.MarshalUint32Array(seatsLeft)}, []byte{})
	return res
}
