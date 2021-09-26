package main

import (
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
	uploaderSign         = "LogWatcher"
	logsTFURL            = "https://logs.tf/upload"
	pickupAPITemplateUrl = "https://api.tf2pickup.%s"
)

type ClientInterface interface {
	Do(r *http.Request) (*http.Response, error)
}

type PickupPlayer struct {
	PlayerID  string
	Class     string
	SteamID64 string
}

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

type GamesResponse struct {
	Results []struct {
		ConnectInfoVersion int    `json:"connectInfoVersion"`
		State              string `json:"state"`
		Number             int    `json:"number"`
		Map                string `json:"map"`
		Slots              []struct {
			ConnectionStatus string `json:"connectionStatus"`
			Status           string `json:"status"`
			GameClass        string `json:"gameClass"`
			Team             string `json:"team"`
			Player           string `json:"player"`
		} `json:"slots"`
		LaunchedAt       time.Time `json:"launchedAt"`
		GameServer       string    `json:"gameServer"`
		StvConnectString string    `json:"stvConnectString"`
		ID               string    `json:"id"`
		LogsUrl          string    `json:"logsUrl,omitempty"`
		Score            struct {
			Red int `json:"red"`
			Blu int `json:"blu"`
		} `json:"score,omitempty"`
		DemoUrl string `json:"demoUrl,omitempty"`
	} `json:"results"`
	ItemCount int `json:"itemCount"`
}

func (lf *LogFile) makeMultipartMap() map[string]io.Reader {
	m := make(map[string]io.Reader)
	m["title"] = strings.NewReader(fmt.Sprintf("tf2pickup.%s #%d", lf.Server.Domain, lf.Game.PickupID))
	m["map"] = strings.NewReader(lf.Game.Map)
	m["key"] = strings.NewReader(lf.apiKey)
	m["logfile"] = &lf.buffer
	m["uploader"] = strings.NewReader(uploaderSign)
	return m
}

func (lf *LogFile) uploadLogFile(client ClientInterface) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, reader := range lf.makeMultipartMap() {
		var writer io.Writer
		var err error
		if key == "logfile" {
			if writer, err = w.CreateFormFile(key, "upload.log"); err != nil {
				return err
			}
		} else {
			if writer, err = w.CreateFormField(key); err != nil {
				return err
			}
		}
		if _, err = io.Copy(writer, reader); err != nil {
			return err
		}
	}
	w.Close()

	req, err := http.NewRequest(http.MethodPost, logsTFURL, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("failed to read buffer: %s", err)
		}
		return fmt.Errorf("logs.tf returned code: %d, body: %s", res.StatusCode, string(bodyBytes))
	}
	return nil
}

func (lf *LogFile) updatePickupInfo(client ClientInterface) error {
	var gr GamesResponse
	url := fmt.Sprintf(pickupAPITemplateUrl+"/games", lf.Server.Domain)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("api.tf2pickup.%s/games returned bad status: %d", lf.Server.Domain, resp.StatusCode)
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return err
	}

	for _, result := range gr.Results {
		if result.State == StartedState && result.Map == lf.Game.Map {
			players := make([]PickupPlayer, len(result.Slots))
			for _, player := range result.Slots {
				p := PickupPlayer{PlayerID: player.Player, Class: player.GameClass}
				players = append(players, p)
			}
			lf.Game.PickupID = result.Number
			break
		}
	}
	return nil
}

func (lf *LogFile) resolvePlayers(client ClientInterface) error {
	var responses []PlayersResponse
	url := fmt.Sprintf(pickupAPITemplateUrl+"/players", lf.Server.Domain)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("api.tf2pickup.%s/players returned bad status: %d", lf.Server.Domain, resp.StatusCode)
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&responses); err != nil {
		return err
	}

	for _, pickupPlayer := range lf.Game.Players {
		for _, pr := range responses {
			if pickupPlayer.PlayerID == pr.Id {
				pickupPlayer.SteamID64 = pr.SteamId
			}
		}
	}
	return nil
}
