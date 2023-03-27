package server

import (
	"fmt"
	"log"
	"net"
)

type UDPServer struct {
	ListenAddr string
	quitch     chan struct{}
	MsgChannel chan Message
	Ln         net.UDPConn
}

func NewUDPServer(listenAddr string) *UDPServer {
	return &UDPServer{
		ListenAddr: listenAddr,
		quitch:     make(chan struct{}),
		MsgChannel: make(chan Message, 10),
	}
}

func (s *UDPServer) Start() error {
	laddr, err := net.ResolveUDPAddr("udp", s.ListenAddr)
	if err != nil {
		log.Fatal(err)
	}

	ln, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return err
	}

	defer ln.Close()
	s.Ln = *ln

	go s.readIngress(*ln)

	<-s.quitch
	close(s.MsgChannel)

	return nil
}

func (s *UDPServer) readIngress(conn net.UDPConn) {
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Fatal((err))
			continue
		}

		s.MsgChannel <- Message{
			Sender:  addr.String(),
			Payload: buf[:n],
		}

		fmt.Println("New connection received from:", addr)

		_, err = conn.WriteToUDP([]byte("Processing request\n"), addr)
		if err != nil {
			log.Fatal(err)
			continue
		}
	}
}
