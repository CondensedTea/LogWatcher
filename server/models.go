package main

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
	Name    string `yaml:"Name"`
	Address string `yaml:"Address"`
}

//func postLogFile(fields map[string]string, buffer bytes.Buffer) {
//	writer := multipart.NewWriter(&buffer)
//	defer writer.Close()
//	writer.CreateFormFile()
//}

type Config struct {
	Server struct {
		Host string `yaml:"Host"`
	} `yaml:"Server"`
	Clients []Client `yaml:"Clients"`
}
