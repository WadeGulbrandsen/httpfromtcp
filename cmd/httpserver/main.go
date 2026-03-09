package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/WadeGulbrandsen/httpfromtcp/internal/headers"
	"github.com/WadeGulbrandsen/httpfromtcp/internal/request"
	"github.com/WadeGulbrandsen/httpfromtcp/internal/response"
	"github.com/WadeGulbrandsen/httpfromtcp/internal/server"
)

const port = 42069
const bufferSize = 1024

func main() {
	server, err := server.Serve(port, handler)
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

func handler(w *response.Writer, req *request.Request) {
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		handle400(w, req)
	case "/myproblem":
		handle500(w, req)
	case "/video":
		videoHandler(w, req)
	}
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		proxyHandler(w, req)
		return
	}
	handle200(w, req)
}

func handle200(w *response.Writer, _ *request.Request) {
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

func handle400(w *response.Writer, _ *request.Request) {
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
}

func handle500(w *response.Writer, _ *request.Request) {
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
}

func videoHandler(w *response.Writer, req *request.Request) {
	videoPath := "assets/vim.mp4"
	fmt.Println("Sending", videoPath)

	data, err := os.ReadFile(videoPath)
	if err != nil {
		handle500(w, req)
		return
	}

	w.WriteStatusLine(response.StatusCodeSuccess)

	h := response.GetDefualtHeaders(len(data))
	h.Override("Content-Type", "video/mp4")
	w.WriteHeaders(h)

	w.WriteBody(data)
}

func proxyHandler(w *response.Writer, req *request.Request) {
	url := fmt.Sprintf("https://httpbin.org/%s", strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/"))
	fmt.Println("Proxying to", url)

	res, err := http.Get(url)
	if err != nil {
		handle500(w, req)
		return
	}
	defer res.Body.Close()

	w.WriteStatusLine(response.StatusCodeSuccess)

	h := response.GetDefualtHeaders(0)
	h.Del("Content-Length")
	h.Override("Transfer-Encoding", "chunked")
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")
	w.WriteHeaders(h)

	buf := make([]byte, bufferSize)
	body := []byte{}
	for {
		n, err := res.Body.Read(buf)
		fmt.Println("Read", n, "bytes")
		if n > 0 {
			body = append(body, buf[:n]...)
			_, err = w.WriteChunkedBody(buf[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("Error reading response body:", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}
	sum := sha256.Sum256(body)

	trailers := headers.NewHeaders()
	trailers.Override("X-Content-SHA256", fmt.Sprintf("%x", sum))
	trailers.Override("X-Content-Length", fmt.Sprintf("%d", len(body)))
	w.WriteTrailers(trailers)
}
