package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/PeterKWIlliams/http/internal/request"
	"github.com/PeterKWIlliams/http/internal/response"
	"github.com/PeterKWIlliams/http/internal/server"
)

const port = 32020

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
	resHeaders := response.GetDefaultHeaders(0)
	resHeaders.Set("content-type", "text/html")
	switch req.RequestLine.RequestTarget {
	case "/yourproblem":
		body, err := os.ReadFile("message1.html")
		if err != nil {
			err = server.WriteError(w, response.InternalServerError, "could not retrieve file")
			if err != nil {
				log.Printf("Could not write error %v", err)
			}
			return
		}
		resHeaders.Set("content-length", strconv.Itoa(len(body)))
		err = w.Write(response.BadRequest, resHeaders, body)
		if err != nil {
			log.Printf("Could not write error %v", err)
		}

	case "/myproblem":
		body, err := os.ReadFile("message2.html")
		if err != nil {
			err = server.WriteError(w, response.InternalServerError, "could not retrieve file")
			if err != nil {
				log.Printf("Could not write error %v", err)
			}
			return
		}
		resHeaders.Set("content-length", strconv.Itoa(len(body)))
		contentLength, _ := resHeaders.Get("content-length")
		log.Printf("This is where i got to %s", contentLength)
		err = w.Write(response.InternalServerError, resHeaders, body)
		if err != nil {
			log.Printf("Could not write error %v", err)
		}
	default:
		body, err := os.ReadFile("message3.html")
		if err != nil {
			err = server.WriteError(w, response.InternalServerError, "could not retrieve file")
			if err != nil {
				log.Printf("Could not write error %v", err)
			}
			return
		}
		resHeaders.Set("content-length", strconv.Itoa(len(body)))
		err = w.Write(response.OK, resHeaders, body)
		if err != nil {
			log.Printf("Error writing success response body:%v", err)
		}
	}
}
