package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"errors"
	"github.com/kisom/aescrypt/secretbox"
)

const UpAndMulticast = (net.FlagUp | net.FlagMulticast)
var BrotocolGroup *net.UDPAddr = &net.UDPAddr{IP: []byte{225,0,0,1}, Port: 12345}

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
	conn, err := net.ListenUDP("udp4", BrotocolGroup)
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
		box, ok := secretbox.Seal(b, []byte("$1HLK#5qAFlEH$4TM5W*TSM18&dW5zFh8f7TujH7#RobIgIn"))
		if !ok {
			log.Println("Couldn't encrypt message. It wasn't sent.")
			continue
		}
		_, err = conn.WriteToUDP(box, BrotocolGroup)
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
	conn, err := net.ListenMulticastUDP("udp4", ifi, BrotocolGroup)
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
		jsonMsg, ok := secretbox.Open(payload[:n], []byte("$1HLK#5qAFlEH$4TM5W*TSM18&dW5zFh8f7TujH7#RobIgIn"))
		if !ok {
			log.Println("Couldn't decrypt message")
			continue
		}
		json.Unmarshal(jsonMsg, &msg)
		r.incoming <- msg
	}
}
