package api

import (
	"bytes"
	"fmt"
	"goflysys/pkg/marshal"
	"log"
)

type FlightsRouter struct {
	Routes map[uint32]func(Response []byte, Request []byte, fdb *FlightDatabase, user string)
}

func NewFlightsRouter() *FlightsRouter {
	return &FlightsRouter{
		Routes: make(map[uint32]func(Response []byte, Request []byte, fdb *FlightDatabase, user string)),
	}
}

func (r *FlightsRouter) HandleFunc(path uint32, handler func(Response []byte, Request []byte, fdb *FlightDatabase, user string)) {
	r.Routes[path] = handler
}

func GetFlightsHandler(res []byte, data []byte, fdb *FlightDatabase, user string) {
	source := marshal.UnmarshalString(data)
	destination := marshal.UnmarshalString(data[len(source)-1:])
	flights, err := fdb.GetFlights(source, destination)
	if err != nil {
		log.Fatal(fmt.Errorf("[SERVICE ERROR] %s", err))
	}
	var flightids []byte
	count := 0
	for _, flight := range flights {
		count++
		flightids = append(res, marshal.MarshalUint32(flight.id)...)
	}

	res = bytes.Join([][]byte{res, marshal.MarshalUint32(uint32(count)), flightids}, []byte{})
}

func GetFlightByIdHandler(res []byte, data []byte, fdb *FlightDatabase, user string) {
	id := marshal.UnmarshalUint32(data[:4])
	flight, err := fdb.GetFlightById(id)
	if err != nil {
		log.Fatal(fmt.Errorf("[SERVICE ERROR] %s", err))
	}

	res = bytes.Join([][]byte{res, marshal.MarshalInt64(flight.departureTime.Unix()), marshal.MarshalFloat64(flight.price), marshal.MarshalUint32(flight.seatsLeft)}, []byte{})
}

func ReserveFlightHandler(res []byte, data []byte, fdb *FlightDatabase, user string) {
	id := marshal.UnmarshalUint32(data[:4])
	numSeats := marshal.UnmarshalUint32(data[4:8])

	seatsReserved, err := fdb.ReserveFlight(id, numSeats, user)
	if err != nil {
		log.Fatal(fmt.Errorf("[SERVICE ERROR] %s", err))
	}

	res = bytes.Join([][]byte{res, marshal.MarshalUint32Array(seatsReserved)}, []byte{})
}

func SubscribeFlightByIdHandler(res []byte, data []byte, fdb *FlightDatabase, user string) {
	id := marshal.UnmarshalUint32((data[:4]))

	flight, err := fdb.SubscribeFlightById(id, user)
	if err != nil {
		log.Fatal(fmt.Errorf("[SERVICE ERROR] %s", err))
	}

	res = bytes.Join([][]byte{res, marshal.MarshalUint32(flight.id)}, []byte{})
}
