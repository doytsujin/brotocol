package main

import (
	"log"
	"net"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "224.1.1.1:6969")
	if err != nil {
		log.Fatal("Error resolving UDP addr", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal("Error listening UDP", err)
	}
	defer conn.Close()
	payload := []byte("'Sup Bro!")
	n, oobn, err := conn.WriteMsgUDP(payload, nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Sent %d bytes and %d oob", n, oobn)

}
