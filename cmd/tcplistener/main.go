package main

import (
	"fmt"
	"log"
	"net"

	"github.com/WadeGulbrandsen/httpfromtcp/internal/request"
)

const port = 42069

func main() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Error opening port: %v\n", err)
	}
	defer listener.Close()

	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Fatalf("Error getting connection: %v\n", err)
		}
		fmt.Printf("Connection accepted from: %v\n", connection.RemoteAddr())

		req, err := request.RequestFromReader(connection)
		if err != nil {
			log.Printf("Error reading request: %v\n", err)
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %v\n", req.RequestLine.Method)
		fmt.Printf("- Target: %v\n", req.RequestLine.RequestTarget)
		fmt.Printf("- Version: %v\n", req.RequestLine.HttpVersion)

		fmt.Printf("Connection closed: %v\n", connection.RemoteAddr())
	}
}
