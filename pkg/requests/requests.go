package requests

import (
	"LogWatcher/pkg/stats"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const (
	logsTFURL            = "http://logs.tf/upload"
	PickupAPITemplateUrl = "https://api.tf2pickup.%s"
)

const uploaderSignTemplate = "LogWatcher %s"

var Version = "dev"

type Request struct {
	client HTTPDoer
	apiKey string
}

type Requester interface {
	MakeMultipartMap(_map, domain string, pickupID int, buf bytes.Buffer) map[string]io.Reader
	UploadLogFile(payload map[string]io.Reader) error
	GetPickupGames(domain string) (GamesResponse, error)
	ResolvePlayers(domain string, players []*stats.PickupPlayer) error
}

// HTTPDoer is interface for doing http requests
type HTTPDoer interface {
	Do(r *http.Request) (*http.Response, error)
}

// PlayersResponse represents single player entry from api.tf2pickup.*/players
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
	Id             string    `json:"id"`
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
	ID                 string    `json:"id"`
	LogsUrl            string    `json:"logsUrl,omitempty"`
	Score              Score     `json:"score,omitempty"`
	DemoUrl            string    `json:"demoUrl,omitempty"`
}

// GamesResponse represents response from api.tf2pickup.*/games
type GamesResponse struct {
	Results   []Result `json:"results"`
	ItemCount int      `json:"itemCount"`
}

func NewRequest(apiKey string, client HTTPDoer) *Request {
	return &Request{
		apiKey: apiKey,
		client: client,
	}
}

func (r *Request) MakeMultipartMap(_map, domain string, pickupID int, buf bytes.Buffer) map[string]io.Reader {
	m := make(map[string]io.Reader)
	m["title"] = strings.NewReader(fmt.Sprintf("tf2pickup.%s #%d", domain, pickupID))
	m["map"] = strings.NewReader(_map)
	m["key"] = strings.NewReader(r.apiKey)
	m["logfile"] = &buf
	m["uploader"] = strings.NewReader(fmt.Sprintf(uploaderSignTemplate, Version))
	return m
}

// UploadLogFile is used for uploading multipart payload to logs.tf/upload endpoint
func (r *Request) UploadLogFile(payload map[string]io.Reader) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, reader := range payload {
		var writer io.Writer
		if key == "logfile" {
			writer, _ = w.CreateFormFile(key, "upload.log") // err is almost always nil
		} else {
			writer, _ = w.CreateFormField(key) // err is almost always nil
		}
		io.Copy(writer, reader) // err is almost always nil
	}
	w.Close()

	req, err := http.NewRequest(http.MethodPost, logsTFURL, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := r.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(res.Body) // err is almost always nil
		return fmt.Errorf("logs.tf returned code: %d, body: %s", res.StatusCode, string(bodyBytes))
	}
	return nil
}

// GetPickupGames is making http request to pickup API and returns GamesResponse,
// containing list of games
func (r *Request) GetPickupGames(domain string) (GamesResponse, error) {
	var gr GamesResponse
	url := fmt.Sprintf(PickupAPITemplateUrl+"/games", domain)
	req, _ := http.NewRequest(http.MethodGet, url, nil) // err is always nil
	resp, err := r.client.Do(req)
	if err != nil {
		return gr, err
	}
	if resp.StatusCode != http.StatusOK {
		return gr, fmt.Errorf("api.tf2pickup.%s/games returned bad status: %d", domain, resp.StatusCode)
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return gr, err
	}
	return gr, nil
}

// ResolvePlayers is used for populating PickupPlayer entries with correct Steam ids
func (r *Request) ResolvePlayers(domain string, players []*stats.PickupPlayer) error {
	var responses []PlayersResponse
	url := fmt.Sprintf(PickupAPITemplateUrl+"/players", domain)
	req, _ := http.NewRequest(http.MethodGet, url, nil) // err is always nil
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("api.tf2pickup.%s/players returned bad status: %d", domain, resp.StatusCode)
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		return err
	}
	for _, pickupPlayer := range players {
		for _, pr := range responses {
			if pickupPlayer.PlayerID == pr.Id {
				pickupPlayer.SteamID = pr.SteamId
			}
		}
	}
	return nil
}
