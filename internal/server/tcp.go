package server

import (
	"fmt"
	"log"
	"net"
)

type Message struct {
	Sender  string
	Payload []byte
}

type TCPServer struct {
	ListenAddr string
	quitch     chan struct{}
	MsgChannel chan Message
	ln         net.Listener
}

func NewTCPServer(listenAddr string) *TCPServer {
	return &TCPServer{
		ListenAddr: listenAddr,
		quitch:     make(chan struct{}),
		MsgChannel: make(chan Message, 10),
	}
}

func (s *TCPServer) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln

	go s.acceptConnections()

	<-s.quitch
	close(s.MsgChannel)

	return nil
}

func (s *TCPServer) acceptConnections() {
	for {
		conn, err := s.ln.Accept()
		if err != nil {
			log.Fatal(err)
			continue
		}

		fmt.Println("New connection received from:", conn.RemoteAddr())

		go s.readIngress(conn)
	}
}

func (s *TCPServer) readIngress(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 2048)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatal(err)
			continue
		}

		s.MsgChannel <- Message{
			Sender:  conn.RemoteAddr().String(),
			Payload: buf[:n],
		}

		conn.Write([]byte("Processing request\n"))
	}

}
