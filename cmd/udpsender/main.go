package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	UDPaddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32020")
	if err != nil {
		log.Fatalf("could not resolve udp address: %s", err)
	}
	conn, err := net.DialUDP("udp", nil, UDPaddr)
	if err != nil {
		log.Println("Error dial UDP", err)
		os.Exit(1)
	}
	defer conn.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println(">")
		string, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from stdin %s", err)
			os.Exit(1)
		}
		_, err = conn.Write([]byte(string))
		if err != nil {
			log.Printf("Error writing to UDP %s", err)
			os.Exit(1)
		}
	}
}
