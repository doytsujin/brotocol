package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"errors"
)

const UpAndMulticast = (net.FlagUp | net.FlagMulticast)
const BrotocolGroup = "225.0.0.1:12345"

type Message struct {
	Sender, Body string
}

type Listener interface {
	Incoming() <-chan Message
}

type Sender interface {
	Send(msg Message)
}

type room struct {
	incoming chan Message
	outgoing chan Message
}

func (r *room) Incoming() <-chan Message {
	return r.incoming
}

func (r *room) Send(msg Message) {
	r.outgoing <- msg
}

func NewRoom() (r *room) {
	return &room{
		incoming: make(chan Message),
		outgoing: make(chan Message),
	}
}

func getMcastInterface() (ifi *net.Interface, err error) {
	ifis, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, ifi := range ifis {
		if (ifi.Flags & UpAndMulticast) == UpAndMulticast && (ifi.Flags & net.FlagLoopback) != net.FlagLoopback{
			return &ifi, nil
		}
	}
	return &net.Interface{}, errors.New("Could not find interface that was up and supports multicast")
}

func main() {
	fmt.Print("Who are you? ")
	var user string
	fmt.Scanln(&user)
	r := NewRoom()
	userInput := make(chan string)
	go r.listenLoop()
	go inputListener(userInput)
	go r.sendLoop(user)
	for {
		select {
		case msg := <-r.Incoming():
			if msg.Sender == user {
				continue
			}
			log.Printf("%s: %s", msg.Sender, msg.Body)
		case body := <-userInput:
			r.Send(Message{Sender: user, Body: body})
		}
	}
}

func inputListener(userInput chan string) {
	rdr := bufio.NewReader(os.Stdin)
	for {
		line, err := rdr.ReadString('\n')
		if err != nil {
			log.Fatal("Error reading your input", err)
		}
		userInput <- line
	}

}

func (r *room) sendLoop(user string) {
	addr, err := net.ResolveUDPAddr("udp4", BrotocolGroup)
	if err != nil {
		log.Fatal("Error resolving UDP addr: ", err)
	}
	conn, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Fatal("Error listening UDP: ", err)
	}
	defer conn.Close()
	for {
		msg := <-r.outgoing
		b, err := json.Marshal(msg)
		if err != nil {
			log.Fatal("Error encoding: ", err)
		}
		_, err = conn.WriteToUDP(b, addr)
		if err != nil {
			log.Printf("Error sending data: %s", err)
		}
	}
}

func (r *room) listenLoop() {
	ifi, err := getMcastInterface()
	log.Printf("Using interface: %v\n", ifi)
	if err != nil {
		log.Fatal("Error obtaining multicast interface: ", err)
	}
	addr, err := net.ResolveUDPAddr("udp4", BrotocolGroup)
	if err != nil {
		log.Fatal("Error from resolve UDP addr: ", err)
	}
	conn, err := net.ListenMulticastUDP("udp4", ifi, addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	for {
		payload := make([]byte, 4096)
		n, _, err := conn.ReadFromUDP(payload)
		if err != nil {
			log.Println("Error reading message: ", err)
		}
		var msg Message
		json.Unmarshal(payload[:n], &msg)
		r.incoming <- msg
	}
}
