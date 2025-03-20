package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
)

func main() {
	Listener, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatalf("could not listen on port 42069: %s", err)
	}
	defer Listener.Close()
	for {
		conn, err := Listener.Accept()
		if err != nil {
			log.Printf("could not accept incoming connection: %s", err)
			continue
		}
		fmt.Printf("connection from %s\n has been accepted", conn.RemoteAddr().String())
		lineChan := getLinesChannel(conn)

		for line := range lineChan {
			fmt.Printf("line: %s\n", line)
		}
		fmt.Printf("connection from %s\n has been closed", conn.RemoteAddr().String())
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)
	go func() {
		defer f.Close()
		defer close(lines)

		currentLineContents := ""
		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)
			if err != nil {
				if currentLineContents != "" {
					lines <- fmt.Sprintf("%s\n", currentLineContents)
				}
				if errors.Is(err, io.EOF) {
					break
				}
				fmt.Printf("error: %s\n", err.Error())
				return
			}
			str := string(buffer[:n])
			parts := strings.Split(str, "\n")
			for i := 0; i < len(parts)-1; i++ {
				lines <- fmt.Sprintf("%s%s", currentLineContents, parts[i])
				currentLineContents = ""
			}
			currentLineContents += parts[len(parts)-1]
		}
	}()
	return lines
}
