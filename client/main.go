package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"os"
	"time"
)

const serverHost = "localhost:27100"

func main() {
	logPath := flag.String("log", "", "Path to log file")
	flag.Parse()

	file, err := os.Open(*logPath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		conn, err := net.Dial("udp", serverHost)
		if err != nil {
			log.Fatal(err)
		}
		_, err = conn.Write([]byte(scanner.Text()))
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
