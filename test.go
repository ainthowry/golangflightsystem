go func() {
	for {
		select {
		case req := <-reqChan:
			// handle the request
			// send the response back to the client
			conn.WriteToUDP([]byte("response"), req.addr)
			// remove the request from the channel
			<-reqChan

		}
	}
}()

// receive requests and add them to the request channel
for {
	data := make([]byte, 1024)
	n, addr, err := conn.ReadFromUDP(data)
	if err != nil {
		log.Println(err)
		continue
	}
	req := Request{data[:n], addr}
	reqChan <- req
}