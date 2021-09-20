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
	serverHost := flag.String("host", "localhost:27100", "Address of LogWatcher server")
	flag.Parse()

	file, err := os.Open(*logPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	conn, err := net.Dial("udp4", *serverHost)
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
