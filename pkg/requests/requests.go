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
)

const (
	logsTFURL            = "http://logs.tf/upload"
	PickupAPITemplateUrl = "https://api.tf2pickup.%s"
)

const uploaderSignTemplate = "LogWatcher %s"

var Version = "dev"

type LogsClient struct {
	Client HTTPDoer
	ApiKey string
}

type LogProcessor interface {
	MakeMultipartMap(_map, domain string, pickupID int, buf bytes.Buffer) map[string]io.Reader
	UploadLogFile(payload map[string]io.Reader) error
	GetPickupGames(domain string) (GamesResponse, error)
	ResolvePlayers(domain string, players []*stats.PickupPlayer) error
}

// HTTPDoer is interface for doing http requests
type HTTPDoer interface {
	Do(r *http.Request) (*http.Response, error)
}

func NewRequester(apiKey string, client HTTPDoer) *LogsClient {
	return &LogsClient{
		ApiKey: apiKey,
		Client: client,
	}
}

func (r *LogsClient) MakeMultipartMap(Map, domain string, pickupID int, buf bytes.Buffer) map[string]io.Reader {
	m := make(map[string]io.Reader)
	m["title"] = strings.NewReader(fmt.Sprintf("tf2pickup.%s #%d", domain, pickupID))
	m["map"] = strings.NewReader(Map)
	m["key"] = strings.NewReader(r.ApiKey)
	m["logfile"] = &buf
	m["uploader"] = strings.NewReader(fmt.Sprintf(uploaderSignTemplate, Version))
	return m
}

// UploadLogFile is used for uploading multipart payload to logs.tf/upload endpoint
func (r *LogsClient) UploadLogFile(payload map[string]io.Reader) error {
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

	res, err := r.Client.Do(req)
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
func (r *LogsClient) GetPickupGames(domain string) (GamesResponse, error) {
	var gr GamesResponse
	url := fmt.Sprintf(PickupAPITemplateUrl+"/games", domain)
	req, _ := http.NewRequest(http.MethodGet, url, nil) // err is always nil
	resp, err := r.Client.Do(req)
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
func (r *LogsClient) ResolvePlayers(domain string, players []*stats.PickupPlayer) error {
	var responses []PlayersResponse
	url := fmt.Sprintf(PickupAPITemplateUrl+"/players", domain)
	req, _ := http.NewRequest(http.MethodGet, url, nil) // err is always nil
	resp, err := r.Client.Do(req)
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
	fmt.Println(players)
	for _, pickupPlayer := range players {
		for _, pr := range responses {
			if pickupPlayer.PlayerID == pr.Id {
				pickupPlayer.SteamID = pr.SteamId
			}
		}
	}
	return nil
}
