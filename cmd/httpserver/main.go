package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/WadeGulbrandsen/httpfromtcp/internal/request"
	"github.com/WadeGulbrandsen/httpfromtcp/internal/response"
	"github.com/WadeGulbrandsen/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handleProblems)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handleProblems(w *response.Writer, req *request.Request) {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		w.WriteStatusLine(response.StatusCodeBadRequest)
		body := []byte(`<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>
`)
		headers := response.GetDefualtHeaders(len(body))
		headers.Override("Content-Type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody(body)
	case "/myproblem":
		w.WriteStatusLine(response.StatusCodeInternalServerError)
		body := []byte(`<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>
`)
		headers := response.GetDefualtHeaders(len(body))
		headers.Override("Content-Type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody(body)
	default:
		w.WriteStatusLine(response.StatusCodeSuccess)
		body := []byte(`<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>
`)
		headers := response.GetDefualtHeaders(len(body))
		headers.Override("Content-Type", "text/html")
		w.WriteHeaders(headers)
		w.WriteBody(body)
	}
}
