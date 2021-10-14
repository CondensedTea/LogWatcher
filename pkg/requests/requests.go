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

// Version is build version, used in logs.tf uploader field
var Version = "dev"

// Client holding http client and logs.tf API key
type Client struct {
	Client HTTPDoer
	ApiKey string
}

// LogUploader provides methods for processing logs
// and interacting with logs.tf and tf2pickup APIs
type LogUploader interface {
	MakeMultipartMap(_map, domain string, pickupID int, buf bytes.Buffer) map[string]io.Reader
	UploadLogFile(payload map[string]io.Reader) error
	GetPickupGames(domain string) (GamesResponse, error)
	ResolvePlayers(domain string, players []*stats.PickupPlayer) error
}

// HTTPDoer is interface for doing http requests
type HTTPDoer interface {
	Do(r *http.Request) (*http.Response, error)
}

// NewClient is client factory
func NewClient(apiKey string, client HTTPDoer) *Client {
	return &Client{
		ApiKey: apiKey,
		Client: client,
	}
}

// MakeMultipartMap constructs logs.tf/upload multipart payload from provided values
func (c *Client) MakeMultipartMap(Map, domain string, pickupID int, buf bytes.Buffer) map[string]io.Reader {
	m := make(map[string]io.Reader)
	m["title"] = strings.NewReader(fmt.Sprintf("tf2pickup.%s #%d", domain, pickupID))
	m["map"] = strings.NewReader(Map)
	m["key"] = strings.NewReader(c.ApiKey)
	m["logfile"] = &buf
	m["uploader"] = strings.NewReader(fmt.Sprintf(uploaderSignTemplate, Version))
	return m
}

// UploadLogFile is used for uploading multipart payload to logs.tf/upload endpoint
func (c *Client) UploadLogFile(payload map[string]io.Reader) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, reader := range payload {
		var writer io.Writer
		if key == "logfile" {
			writer, _ = w.CreateFormFile(key, "upload.log") // err is almost always nil
		} else {
			writer, _ = w.CreateFormField(key)
		}
		io.Copy(writer, reader)
	}
	w.Close()

	req, _ := http.NewRequest(http.MethodPost, logsTFURL, &b) // err is always nil
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(res.Body) // err is almost always nil
		return fmt.Errorf("logs.tf returned code: %d, body: %s", res.StatusCode, string(bodyBytes))
	}
	return nil
}

// GetPickupGames makes http request to pickup API and returns GamesResponse, containing list of games
func (c *Client) GetPickupGames(domain string) (GamesResponse, error) {
	var gr GamesResponse
	url := fmt.Sprintf(PickupAPITemplateUrl+"/games", domain)
	req, _ := http.NewRequest(http.MethodGet, url, nil) // err is always nil
	resp, err := c.Client.Do(req)
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

// ResolvePlayers populating PickupPlayer entries with correct SteamIDs
func (c *Client) ResolvePlayers(domain string, players []*stats.PickupPlayer) error {
	var responses []PlayersResponse
	url := fmt.Sprintf(PickupAPITemplateUrl+"/players", domain)
	req, _ := http.NewRequest(http.MethodGet, url, nil) // err is always nil
	resp, err := c.Client.Do(req)
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
