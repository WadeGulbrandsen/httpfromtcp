package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	messagesFile, err := os.Open("messages.txt")
	if err != nil {
		log.Fatalf("Error reading messages.txt: %v\n", err)
	}
	linesChannel := getLinesChannel(messagesFile)
	for line := range linesChannel {
		fmt.Printf("read: %s\n", line)
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
