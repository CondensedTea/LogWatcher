package requests

import "time"

// PlayersResponse represents single player entry from api.tf2pickup.*/Players
type PlayersResponse struct {
	SteamId string `json:"steamId"`
	Name    string `json:"name"`
	Avatar  struct {
		Small  string `json:"small"`
		Medium string `json:"medium"`
		Large  string `json:"large"`
	} `json:"avatar,omitempty"`
	Roles          []string  `json:"roles"`
	Etf2LProfileId int       `json:"etf2lProfileId"`
	JoinedAt       time.Time `json:"joinedAt"`
	Id             string    `json:"ID"`
	Links          []struct {
		Href  string `json:"href"`
		Title string `json:"title"`
	} `json:"_links"`
}

type Score struct {
	Red int `json:"red"`
	Blu int `json:"blu"`
}

type Slot struct {
	ConnectionStatus string `json:"connectionStatus"`
	Status           string `json:"status"`
	GameClass        string `json:"gameClass"`
	Team             string `json:"team"`
	Player           string `json:"player"`
}

type Result struct {
	ConnectInfoVersion int       `json:"connectInfoVersion"`
	State              string    `json:"state"`
	Number             int       `json:"number"`
	Map                string    `json:"map"`
	Slots              []Slot    `json:"slots"`
	LaunchedAt         time.Time `json:"launchedAt"`
	GameServer         string    `json:"gameServer"`
	StvConnectString   string    `json:"stvConnectString"`
	ID                 string    `json:"ID"`
	LogsUrl            string    `json:"logsUrl,omitempty"`
	Score              Score     `json:"score,omitempty"`
	DemoUrl            string    `json:"demoUrl,omitempty"`
}

// GamesResponse represents response from api.tf2pickup.*/games
type GamesResponse struct {
	Results   []Result `json:"results"`
	ItemCount int      `json:"itemCount"`
}
