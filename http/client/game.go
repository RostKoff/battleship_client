package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	serverApi    = "https://go-pjatk-server.fly.dev/api"
	tokenKey     = "X-Auth-Token"
	unmarshalErr = "failed to unmarshal from ReadCloser"
)

type Game struct {
	Token string
}

type GameSettings struct {
	Coords      []string `json:"coords"`
	Description string   `json:"description"`
	Nick        string   `json:"nick"`
	TargetNick  string   `json:"target_nick"`
	AgainstBot  bool     `json:"wpbot"`
}

type StatusResponse struct {
	Status         string   `json:"game_status"`
	LastGameStatus string   `json:"last_game_status"`
	Nick           string   `json:"nick"`
	OpponentShots  []string `json:"opp_shots"`
	Opponent       string   `json:"opponent"`
	ShouldFire     bool     `json:"should_fire"`
	Timer          int      `json:"timer"`
	Message        string
}

type DescriptionResponse struct {
	PlayerDescription   string `json:"desc"`
	OpponentDescription string `json:"opp_desc"`
	Message             string
}

type boardResponse struct {
	Board   []string
	Message string
}

func InitGame(settings GameSettings) (Game, error) {
	requestBody, err := json.Marshal(settings)
	game := Game{}
	if err != nil {
		return game, fmt.Errorf("failed to marshal settings to json: %w", err)
	}
	r := bytes.NewReader(requestBody)
	res, err := http.Post(serverApi+"/game", "application/json", r)
	if err != nil {
		return game, fmt.Errorf("failed to send POST request: %w", err)
	}
	if res.StatusCode == 400 {
		responseBody, err := unmarshalFromReadCloser[map[string]any](&res.Body)
		if err != nil {
			return game, fmt.Errorf("%s: %w", unmarshalErr, err)
		}
		mes, ok := responseBody["message"]
		if !ok {
			mes = res.Status
		}
		return game, fmt.Errorf("response error: %s", mes)
	}
	game.Token = res.Header.Get(tokenKey)
	return game, nil
}

func (g Game) Board() ([]string, error) {
	res, err := g.sendRequest(http.MethodGet, "/game/board", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send GET request: %w", err)
	}
	boardRes, err := unmarshalFromReadCloser[boardResponse](&res.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", unmarshalErr, err)
	}
	if res.StatusCode == 401 || res.StatusCode == 403 {
		return nil, fmt.Errorf("response error: %s", boardRes.Message)
	}
	return boardRes.Board, err
}

func (g Game) Status() (StatusResponse, error) {
	statusRes := StatusResponse{}
	res, err := g.sendRequest(http.MethodGet, "/game", nil)
	if err != nil {
		return statusRes, fmt.Errorf("failed to send GET request: %w", err)
	}
	statusRes, err = unmarshalFromReadCloser[StatusResponse](&res.Body)
	if err != nil {
		return statusRes, fmt.Errorf("%s: %w", unmarshalErr, err)
	}
	if res.StatusCode == 403 || res.StatusCode == 401 || res.StatusCode == 429 {
		return statusRes, fmt.Errorf("response error: %s", statusRes.Message)
	}
	return statusRes, nil
}

func (g Game) Fire(coord string) (string, error) {
	coords := make(map[string]string)
	coords["coord"] = coord
	reqBody, err := json.Marshal(coords)
	if err != nil {
		return "", fmt.Errorf("failed to marshal coord to json: %w", err)
	}
	r := bytes.NewReader(reqBody)
	res, err := g.sendRequest(http.MethodPost, "/game/fire", r)
	if err != nil {
		return "", fmt.Errorf("failed to send GET request: %w", err)
	}
	jsonBody, err := unmarshalFromReadCloser[map[string]string](&res.Body)
	if err != nil {
		return "", fmt.Errorf("%s: %w", unmarshalErr, err)
	}

	if sc := res.StatusCode; sc == 400 || sc == 401 || sc == 403 || sc == 429 {
		mes, ok := jsonBody["message"]
		if !ok {
			mes = res.Status
		}
		return "", fmt.Errorf("response error: %s", mes)
	}
	result, ok := jsonBody["result"]
	fmt.Println(result)
	if !ok {
		return "", fmt.Errorf("result not found")
	}
	return result, nil
}

func (g Game) PlayerDescriptions() (DescriptionResponse, error) {
	descriptionRes := DescriptionResponse{}
	res, err := g.sendRequest(http.MethodGet, "/game/desc", nil)
	if err != nil {
		return descriptionRes, fmt.Errorf("failed to send GET request: %w", err)
	}
	descriptionRes, err = unmarshalFromReadCloser[DescriptionResponse](&res.Body)
	if err != nil {
		return descriptionRes, fmt.Errorf("%s: %w", unmarshalErr, err)
	}
	if sc := res.StatusCode; sc == 401 || sc == 404 || sc == 429 {
		return descriptionRes, fmt.Errorf("response error: %s", descriptionRes.Message)
	}
	return descriptionRes, nil
}

func (g Game) sendRequest(method string, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", serverApi, path), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create new http request: %w", err)
	}
	req.Header.Add(tokenKey, g.Token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send http request: %w", err)
	}
	return res, nil
}
