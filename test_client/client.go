package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"time"
)

func main() {
	logPath := flag.String("log", "", "Path to log file")
	clientHost := flag.String("from", "localhost:27150", "Address of udp client")
	serverHost := flag.String("to", "localhost:27100", "Address of LogWatcher server")
	flag.Parse()

	file, err := os.Open(*logPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	laddress, err := net.ResolveUDPAddr("udp4", *clientHost)
	if err != nil {
		log.Fatalf("Failed to resolve client host: %s", err)
	}
	raddress, err := net.ResolveUDPAddr("udp4", *serverHost)
	if err != nil {
		log.Fatalf("Failed to resolve server host: %s", err)
	}

	conn, err := net.DialUDP("udp4", laddress, raddress)
	if err != nil {
		log.Fatalf("Failed to dial to UDP server: %s", err)
	}

	for scanner.Scan() {
		_, err = conn.Write([]byte(scanner.Text()))
		if err != nil {
			log.Fatalf("Failed to write to UDP socket: %s", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
