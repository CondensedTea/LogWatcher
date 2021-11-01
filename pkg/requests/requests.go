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

	"github.com/sirupsen/logrus"
)

const (
	logsTFURL            = "http://logs.tf/upload"
	PickupAPITemplateUrl = "https://api.tf2pickup.%s"
)

const uploaderSignTemplate = "LogWatcher %s"

const StartedState = "started"

// Version is build version, used in logs.tf uploader field
var Version = "dev"

// Client holding http client and logs.tf API key
type Client struct {
	Client HTTPDoer
	ApiKey string
	Log    *logrus.Logger
}

type Pickup struct {
	Players []*stats.PickupPlayer
	ID      int
}

// LogUploader provides methods for processing logs
// and interacting with logs.tf and tf2pickup APIs
type LogUploader interface {
	MakeMultipartMap(_map, domain string, pickupID int, buf bytes.Buffer) map[string]io.Reader
	UploadLogFile(payload map[string]io.Reader) error
	ResolvePlayersSteamIDs(domain string, players []*stats.PickupPlayer) error
	FindMatchingPickup(domain, Map string) (*Pickup, error)
}

// HTTPDoer is interface for doing http requests
type HTTPDoer interface {
	Do(r *http.Request) (*http.Response, error)
}

// NewClient is client factory
func NewClient(apiKey string, client HTTPDoer, log *logrus.Logger) *Client {
	return &Client{
		ApiKey: apiKey,
		Client: client,
		Log:    log,
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

// ResolvePlayersSteamIDs populates PickupPlayer entries with correct SteamIDs
func (c *Client) ResolvePlayersSteamIDs(domain string, players []*stats.PickupPlayer) error {
	var responses []PlayersResponse
	url := fmt.Sprintf(PickupAPITemplateUrl+"/Players", domain)
	req, _ := http.NewRequest(http.MethodGet, url, nil) // err is always nil
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("api.tf2pickup.%s/Players returned bad status: %d", domain, resp.StatusCode)
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

// FindMatchingPickup is used for finding current game on tf2pickup API
// and loading to LogFile list of its Players and pickup ID
func (c *Client) FindMatchingPickup(domain, gameMap string) (*Pickup, error) {
	var pickup = &Pickup{}
	players := make([]*stats.PickupPlayer, 0)

	gamesResponse, err := GetPickupGames(domain, c.Client)
	if err != nil {
		return pickup, err
	}
	for _, game := range gamesResponse.Results {
		c.Log.WithFields(logrus.Fields{
			"state": game.State,
			"map":   game.Map,
		}).Infof("looking for pickup...")
		if game.State == StartedState && game.Map == gameMap {
			for _, player := range game.Slots {
				p := &stats.PickupPlayer{
					PlayerID: player.Player, Class: player.GameClass, Team: player.Team,
				}
				players = append(players, p)
			}
			pickup.Players = players
			pickup.ID = game.Number
		}
	}
	return pickup, nil
}

// GetPickupGames makes http request to pickup API and returns GamesResponse, containing list of games
func GetPickupGames(domain string, client HTTPDoer) (GamesResponse, error) {
	var gr GamesResponse
	url := fmt.Sprintf(PickupAPITemplateUrl+"/games", domain)
	req, _ := http.NewRequest(http.MethodGet, url, nil) // err is always nil
	resp, err := client.Do(req)
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
