package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/PeterKWIlliams/http/internal/headers"
	"github.com/PeterKWIlliams/http/internal/request"
	"github.com/PeterKWIlliams/http/internal/response"
	"github.com/PeterKWIlliams/http/internal/server"
)

const port = 32020

func main() {
	server, err := server.Serve(port, routingHandler)
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

func routingHandler(w *response.Writer, req *request.Request) {
	resHeaders := response.GetDefaultHeaders(0)
	resHeaders.Set("content-type", "text/html")
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		targetPath := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")
		url := "https://httpbin.org/" + targetPath

		log.Printf("Proxying request for %s to %s", req.RequestLine.RequestTarget, url)

		resp, err := http.Get(url)
		if err != nil {
			log.Printf("httpbin GET request failed for %s: %v", url, err)
			err = server.WriteError(w, response.InternalServerError, "Proxy request failed")
			if err != nil {
				log.Printf("could not write proxy error response: %v", err)
			}
			return
		}
		defer resp.Body.Close()

		resHeaders := headers.NewHeaders()
		resHeaders.Set("transfer-encoding", "chunked")
		if contentType := resp.Header.Get("Content-Type"); contentType != "" {
			resHeaders.Set("content-type", contentType)
		}
		resHeaders.Set("connection", "close")

		err = w.WriteStatusLine(response.StatusCode(resp.StatusCode))
		if err != nil {
			log.Printf("could not write statusLine %v", err)
			return
		}
		err = w.WriteHeaders(resHeaders)
		if err != nil {
			log.Printf("could not write headers %v", err)
			return
		}

		buffer := make([]byte, 32)
		for {
			n, err := resp.Body.Read(buffer)
			if n > 0 {
				log.Printf("Read %d bytes from httpbin", n)
				_, writeErr := w.WriteChunkedBody(buffer[:n])
				if writeErr != nil {
					log.Printf("error writing chunked body %v", writeErr)
					return
				}
			}

			if err != nil {
				if err == io.EOF {
					_, doneErr := w.WriteChunkedBodyDone()
					if doneErr != nil {
						log.Printf("error writing chunked body done %v", doneErr)
					}
					log.Printf("Finished proxying %s", url)
					return
				}
				log.Printf("Error reading from httpbin body: %v", err)
				return
			}
		}

	} else {
		switch req.RequestLine.RequestTarget {
		case "/yourproblem":
			body, err := os.ReadFile("message1.html")
			if err != nil {
				err = server.WriteError(w, response.InternalServerError, "could not retrieve file")
				if err != nil {
					log.Printf("could not write error %v", err)
				}
				return
			}
			resHeaders.Set("content-length", strconv.Itoa(len(body)))
			resHeaders.Set("content-type", "text/html")
			err = w.Write(response.BadRequest, resHeaders, body)
			if err != nil {
				log.Printf("could not write error %v", err)
			}
			return

		case "/myproblem":
			body, err := os.ReadFile("message2.html")
			if err != nil {
				err = server.WriteError(w, response.InternalServerError, "could not retrieve file")
				if err != nil {
					log.Printf("could not write error %v", err)
				}
				return
			}
			resHeaders.Set("content-length", strconv.Itoa(len(body)))
			resHeaders.Set("content-type", "text/html")
			err = w.Write(response.InternalServerError, resHeaders, body)
			if err != nil {
				log.Printf("could not write error %v", err)
			}
			return

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
			resHeaders.Set("content-type", "text/html")
			err = w.Write(response.OK, resHeaders, body)
			if err != nil {
				log.Printf("Error writing default response body:%v", err)
			}
		}
	}
}
