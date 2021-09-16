package main

import (
	"bytes"
)

type State int

const (
	Pregame State = iota
	Game    State = iota
)

type Upload struct {
	Title string
	Map   string
	Key   string
}

type Client struct {
	name    string
	address string
}

type LogFile struct {
	address string
	state   State
	buffer  bytes.Buffer
}

type Config struct {
	Server struct {
		Host string
		Port string
	}
	Clients []Client
}
