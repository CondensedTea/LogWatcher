package main

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strings"
	"time"
)

type ParsedLogLine struct {
	LogTimestamp time.Time
	SourceServer *net.UDPAddr
	LogLine      string
}

func handleConnection(data string, remote *net.UDPAddr) {
	parsed := parseLogLine(data, remote)
	fmt.Printf("[%s] [%s] : %s\n", parsed.LogTimestamp, parsed.SourceServer, parsed.LogLine)
}

func parseLogLine(line string, remote *net.UDPAddr) ParsedLogLine {
	rLogTimestamp, _ := regexp.Compile(`\d{2}/\d{2}/\d{4} - \d{2}:\d{2}:\d{2}`)
	logTimestamp := rLogTimestamp.FindString(line)
	t, err := time.Parse("01/02/2006 - 15:04:05", logTimestamp)
	if err != nil {
		log.Fatal(err)
	}

	return ParsedLogLine{LogTimestamp: t, SourceServer: remote, LogLine: line}
}

func main() {
	laddr := net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 27100}

	lner, err := net.ListenUDP("udp", &laddr)
	if err != nil {
		log.Fatal(err)
	}
	defer lner.Close()

	fmt.Printf("Log Server listening on %s\n", lner.LocalAddr().String())

	for {
		message := make([]byte, 1024)
		rlen, remote, err := lner.ReadFromUDP(message[:])
		if err != nil {
			log.Panic(err)
		}
		data := strings.TrimSpace(string(message[:rlen]))
		handleConnection(data, remote)
	}
}
