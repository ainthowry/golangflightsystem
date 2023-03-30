package api

import (
	"bytes"
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
	if err != nil && err.Error() == "NotFoundException" {
		log.Println("[SERVICE] No flights found")
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(404))}, []byte{})
		return res
	}
	if err != nil {
		log.Printf("[SERVICE ERROR] %v", err)
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(400))}, []byte{})
		return res
	}

	var flightids []byte
	count := 0
	for _, flight := range flights {
		count++
		flightids = append(flightids, marshal.MarshalUint32(flight.id)...)
	}

	res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(200)), marshal.MarshalUint32(uint32(count)), flightids}, []byte{})
	return res
}

func GetFlightByIdHandler(res_pointer *[]byte, data []byte, fdb *FlightDatabase, user string) []byte {
	res := *res_pointer

	id := marshal.UnmarshalUint32(data[:4])
	flight, err := fdb.GetFlightById(id)
	if err != nil && err.Error() == "NotFoundException" {
		log.Println("[SERVICE] Flight given does not exist", err)
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(404))}, []byte{})
		return res
	}
	if err != nil {
		log.Printf("[SERVICE ERROR] %v", err)
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(400))}, []byte{})
		return res
	}

	res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(200)), marshal.MarshalInt64(flight.departureTime.Unix()), marshal.MarshalFloat64(flight.price), marshal.MarshalUint32(flight.seatsLeft)}, []byte{})
	return res
}

func ReserveFlightHandler(res_pointer *[]byte, data []byte, fdb *FlightDatabase, user string) []byte {
	res := *res_pointer

	id := marshal.UnmarshalUint32(data[:4])
	numSeats := marshal.UnmarshalUint32(data[4:8])

	seatsReserved, err := fdb.ReserveFlight(id, numSeats, user)
	if err != nil && err.Error() == "NotFoundException" {
		log.Println("[SERVICE] Flight given does not exist", err)
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(404))}, []byte{})
		return res
	}
	if err != nil && err.Error() == "Conflict" {
		log.Println("[SERVICE] No flights seats available", err)
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(409))}, []byte{})
		return res
	}
	if err != nil {
		log.Printf("[SERVICE ERROR] %v", err)
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(400))}, []byte{})
		return res
	}

	res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(201)), marshal.MarshalUint32Array(seatsReserved)}, []byte{})
	return res
}

func SubscribeFlightByIdHandler(res_pointer *[]byte, data []byte, fdb *FlightDatabase, user string) []byte {
	res := *res_pointer

	id := marshal.UnmarshalUint32(data[:4])
	endTime := time.Unix(marshal.UnmarshalInt64(data[4:12]), 0)

	_, err := fdb.SubscribeFlightById(id, endTime, user)
	if err != nil && err.Error() == "NotFoundException" {
		log.Println("[SERVICE] Flight given does not exist", err)
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(404))}, []byte{})
		return res
	}
	if err != nil {
		log.Printf("[SERVICE ERROR] %v", err)
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(400))}, []byte{})
		return res
	}

	res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(201)), marshal.MarshalUint32(1)}, []byte{})
	return res
}

func GetSeatsByIdHandler(res_pointer *[]byte, data []byte, fdb *FlightDatabase, user string) []byte {
	res := *res_pointer

	id := marshal.UnmarshalUint32(data[:4])

	seatsReserved, err := fdb.GetSeatsById(id, user)
	if err != nil && err.Error() == "NotFoundException" {
		log.Println("[SERVICE] Flight given does not exist", err)
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(404))}, []byte{})
		return res
	}
	if err != nil {
		log.Printf("[SERVICE ERROR] %v", err)
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(400))}, []byte{})
		return res
	}

	res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(200)), marshal.MarshalUint32Array(seatsReserved)}, []byte{})
	return res
}

func RefundSeatBySeatNumHandler(res_pointer *[]byte, data []byte, fdb *FlightDatabase, user string) []byte {
	res := *res_pointer

	id := marshal.UnmarshalUint32(data[:4])
	seatNum := marshal.UnmarshalUint32(data[4:8])

	isRefunded, err := fdb.RefundSeatBySeatNum(id, seatNum, user)
	if err != nil && err.Error() == "NotFoundException" {
		log.Println("[SERVICE] Flight given does not exist", err)
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(404))}, []byte{})
		return res
	}
	if err != nil && err.Error() == "UnauthorizedException" {
		log.Println("[SERVICE] User is not the buyer", err)
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(401))}, []byte{})
		return res
	}
	if err != nil {
		log.Printf("[SERVICE ERROR] %v", err)
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(400))}, []byte{})
		return res
	}

	if isRefunded {
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(201)), marshal.MarshalUint32(uint32(1))}, []byte{})
	} else {
		res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(400)), marshal.MarshalUint32(uint32(0))}, []byte{})
	}

	return res
}
