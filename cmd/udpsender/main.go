package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

const address = "localhost:42069"

func main() {
	udpaddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Fatalf("Error resolving UDP address %v: %v\n", address, err)
	}
	udpconn, err := net.DialUDP("udp", nil, udpaddr)
	if err != nil {
		log.Fatalf("Error creating UPD connection: %v\n", err)
	}
	defer udpconn.Close()
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("> ")
		line, err := r.ReadString('\n')
		if err != nil {
			log.Printf("Error reading line: %v\n", err)
		}
		_, err = udpconn.Write([]byte(line))
		if err != nil {
			log.Printf("Error writing line: %v\n", err)
		}
	}
}
