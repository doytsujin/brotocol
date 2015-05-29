package main

import (
	"log"
	"net"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "224.1.1.1:6969")
	if err != nil {
		log.Fatal("Error from resolve UDP addr", err)
	}
	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	for {
		payload := make([]byte, 4096)
		oob := make([]byte, 4096)
		_, _, _, _, err := conn.ReadMsgUDP(payload, oob)
		if err != nil {
			log.Fatal(err)
		}
		log.Println(string(payload))
	}
}
