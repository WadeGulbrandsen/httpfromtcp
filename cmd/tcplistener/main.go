package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
		lines := getLinesChannel(connection)
		for line := range lines {
			fmt.Println(line)
		}
		fmt.Printf("Connection closed: %v\n", connection.RemoteAddr())
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	linesChannel := make(chan string)
	go func() {
		defer close(linesChannel)
		line := ""
		for {
			b := make([]byte, 8)
			_, err := f.Read(b)
			if err != nil {
				if line != "" {
					linesChannel <- line
				}
				if !errors.Is(err, io.EOF) {
					log.Printf("Error: %v\n", err)
				}
				break
			}
			parts := strings.Split(string(b), "\n")
			for i := 0; i < len(parts)-1; i++ {
				linesChannel <- line + parts[i]
				line = ""
			}
			line += parts[len(parts)-1]
		}
	}()
	return linesChannel
}
