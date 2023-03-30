package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"goflysys/pkg/marshal"

	"github.com/hashicorp/go-memdb"
)

type Flight struct {
	id            uint32
	source        string
	destination   string
	departureTime time.Time
	price         float64
	seatsLeft     uint32
	seats         map[uint32]Seat
	subs          []Subscriber
}

type Subscriber struct {
	listenAddr string
	endTime    time.Time
}

type Seat struct {
	reserved bool
	buyer    string
}

type FlightDatabase struct {
	db *memdb.MemDB
}

func NewDatabase(timeout time.Duration) (*FlightDatabase, error) {
	_, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	schema := &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"flights": &memdb.TableSchema{
				Name: "flights",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.UintFieldIndex{Field: "id"},
					},
					"source": &memdb.IndexSchema{
						Name:    "source",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "source"},
					},
					"destination": &memdb.IndexSchema{
						Name:    "destination",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "destination"},
					},
				},
			},
		},
	}

	db, err := memdb.NewMemDB(schema)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	boostrapDatabase(db)

	txn := db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("flights", "id")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Ping db:")
	for obj := it.Next(); obj != nil; obj = it.Next() {
		flight := obj.(*Flight)
		fmt.Printf("[%d]%s->%s,", flight.id, flight.source, flight.destination)
	}
	fmt.Println()

	return &FlightDatabase{db: db}, nil
}

func (fdb *FlightDatabase) GetFlights(source string, destination string) ([]*Flight, error) {
	txn := fdb.db.Txn(false)
	defer txn.Abort()

	it, err := txn.Get("flights", "id")
	if err != nil {
		return nil, err
	}

	flights := []*Flight{}

	for obj := it.Next(); obj != nil; obj = it.Next() {
		flight := obj.(*Flight)
		if flight.source == source && flight.destination == destination {
			flights = append(flights, flight)
		}
	}
	if len(flights) == 0 {
		return nil, errors.New("NotFoundException")
	}

	return flights, nil
}

func (fdb *FlightDatabase) GetFlightById(id uint32) (*Flight, error) {
	txn := fdb.db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("flights", "id", id)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, errors.New("NotFoundException")
	}

	return raw.(*Flight), nil
}

func (fdb *FlightDatabase) ReserveFlight(id uint32, numSeats uint32, buyer string) ([]uint32, error) {
	txn := fdb.db.Txn(true)

	flight, err := txn.First("flights", "id", id)
	if err != nil {
		return nil, err
	}
	if flight == nil {
		return nil, errors.New("NotFoundException")
	}

	seatsReserved := make([]uint32, int(numSeats))
	count := 0
	seats := flight.(*Flight).seats
	for seatNum, seat := range seats {
		if count >= int(numSeats) {
			break
		}
		if !seat.reserved {
			seatsReserved[count] = seatNum
			flight.(*Flight).seats[seatNum] = Seat{reserved: true, buyer: buyer}
			flight.(*Flight).seatsLeft--
			count++
		}
	}

	if count == 0 {
		return nil, errors.New("Conflict")
	}

	if txn.Insert("flights", flight); err != nil {
		return nil, err
	}

	go publishToSubscribers(flight.(*Flight).subs, flight.(*Flight).id, flight.(*Flight).seatsLeft)

	txn.Commit()

	return seatsReserved, nil
}

func (fdb *FlightDatabase) SubscribeFlightById(id uint32, endTime time.Time, subscriber string) (*Flight, error) {
	newSub := Subscriber{listenAddr: subscriber, endTime: endTime}
	txn := fdb.db.Txn(true)

	flight, err := txn.First("flights", "id", id)
	if err != nil {
		return nil, err
	}
	if flight == nil {
		return nil, errors.New("NotFoundException")
	}

	flight.(*Flight).subs = append(flight.(*Flight).subs, newSub)

	if txn.Insert("flights", flight); err != nil {
		return nil, errors.New("BadRequestException")
	}

	txn.Commit()

	txn = fdb.db.Txn(false)
	defer txn.Abort()

	updatedFlight, err := txn.First("flights", "id", id)
	if err != nil {
		return nil, err
	}
	if updatedFlight == nil {
		return nil, errors.New("NotFoundException")
	}

	return updatedFlight.(*Flight), nil
}
func (fdb *FlightDatabase) GetSeatsById(id uint32, flyer string) ([]uint32, error) {
	txn := fdb.db.Txn(false)
	defer txn.Abort()

	raw, err := txn.First("flights", "id", id)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, errors.New("NotFoundException")
	}

	flight := raw.(*Flight)
	seatsReserved := make([]uint32, 0)

	for seatNum, seat := range flight.seats {
		if seat.buyer == flyer {
			seatsReserved = append(seatsReserved, seatNum)
		}
	}

	return seatsReserved, nil
}

func (fdb *FlightDatabase) RefundSeatBySeatNum(id uint32, seatNum uint32, flyer string) ([]uint32, error) {
	txn := fdb.db.Txn(true)

	raw, err := txn.First("flights", "id", id)
	if err != nil {
		txn.Abort()
		return nil, err
	}
	if raw == nil {
		txn.Abort()
		return nil, errors.New("NotFoundException")
	}

	flightseat, ok := raw.(*Flight).seats[seatNum]
	if !ok || !flightseat.reserved {
		txn.Abort()
		return nil, errors.New("BadRequestException")
	}

	if flightseat.buyer != flyer {
		txn.Abort()
		return nil, errors.New("UnauthorizedException")
	}

	raw.(*Flight).seats[seatNum] = Seat{reserved: false, buyer: ""}

	if txn.Insert("flights", raw); err != nil {
		return nil, err
	}

	txn.Commit()

	txn = fdb.db.Txn(false)
	defer txn.Abort()

	raw, err = txn.First("flights", "id", id)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, errors.New("NotFoundException")
	}

	flight := raw.(*Flight)
	seatsReserved := make([]uint32, 0)
	for seatNum, seat := range flight.seats {
		if seat.buyer == flyer {
			seatsReserved = append(seatsReserved, seatNum)
		}
	}

	return seatsReserved, nil
}

func boostrapDatabase(db *memdb.MemDB) {
	txn := db.Txn(true)

	flights := []*Flight{
		&Flight{id: uint32(1), source: "CDG", destination: "HND", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(2), source: "BKK", destination: "CUN", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(3), source: "FCO", destination: "BCN", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(4), source: "LHR", destination: "SYD", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(5), source: "DXB", destination: "JFK", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(6), source: "HND", destination: "CDG", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(7), source: "CUN", destination: "DXB", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(8), source: "JFK", destination: "LHR", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(9), source: "BCN", destination: "FCO", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(10), source: "SYD", destination: "BKK", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(11), source: "CDG", destination: "JFK", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(12), source: "BKK", destination: "SYD", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(13), source: "FCO", destination: "DXB", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(14), source: "LHR", destination: "BCN", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(15), source: "DXB", destination: "HND", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(16), source: "SYD", destination: "CUN", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(17), source: "JFK", destination: "CDG", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(18), source: "BCN", destination: "LHR", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(19), source: "HND", destination: "BKK", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(20), source: "CUN", destination: "FCO", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
		&Flight{id: uint32(21), source: "SYD", destination: "BKK", departureTime: time.Now().Add(time.Duration(rand.Intn(24000)) * time.Hour), price: rand.Float64() * 2000, seatsLeft: uint32(100), subs: []Subscriber{}},
	}

	for _, f := range flights {
		seatMap := make(map[uint32]Seat, 100)
		go func() {
			for i := uint32(0); i <= 100; i++ {
				seatMap[i] = Seat{reserved: false, buyer: ""}
			}
		}()
		f.seats = seatMap
		if err := txn.Insert("flights", f); err != nil {
			log.Fatal(err)
		}
	}

	txn.Commit()
}

func publishToSubscribers(subs []Subscriber, id uint32, numSeats uint32) {
	now := time.Now()
	for _, sub := range subs {
		log.Printf("EndTime for user %s is: %s", sub.listenAddr, sub.endTime.String())
		if now.Before(sub.endTime) {
			sendAddr, _ := net.ResolveUDPAddr("udp", sub.listenAddr)
			ln, err := net.DialUDP("udp", nil, sendAddr) //don't care if the user is listening, just send
			if err != nil {
				log.Printf("Failed to send notification to user %s: %v\n", sub.listenAddr, err)
			} else {
				resp := bytes.Join([][]byte{marshal.MarshalUint32(uint32(8888)), marshal.MarshalUint32(id), marshal.MarshalUint32(numSeats)}, []byte{})
				log.Printf("Sending notification to user %s\n", sub.listenAddr)
				ln.Write(resp)
			}
			defer ln.Close()
		}
	}
}
