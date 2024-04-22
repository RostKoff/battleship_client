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
	Message        string   `json:"message"`
}

type BoardResponse struct {
	Board   []string `json:"board"`
	Message string   `json:"message"`
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
		return game, fmt.Errorf("failed to initialize game: %s", mes)
	}
	game.Token = res.Header.Get(tokenKey)
	return game, nil
}

func (g Game) Board() ([]string, error) {
	res, err := g.sendRequest(http.MethodGet, "/game/board", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to send GET request: %w", err)
	}
	boardRes, err := unmarshalFromReadCloser[BoardResponse](&res.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", unmarshalErr, err)
	}
	if res.StatusCode == 401 || res.StatusCode == 403 {
		return nil, fmt.Errorf("failed get board information: %s", boardRes.Message)
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
		return statusRes, fmt.Errorf("failed to retrieve game status: %s", statusRes.Message)
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
	sc := res.StatusCode
	if sc == 400 || sc == 401 || sc == 403 || sc == 429 {
		mes, ok := jsonBody["message"]
		if !ok {
			mes = res.Status
		}
		return "", fmt.Errorf("failed to fire: %s", mes)
	}
	result, ok := jsonBody["result"]
	fmt.Println(result)
	if !ok {
		return "", fmt.Errorf("failed to fire: Result not found")
	}
	return result, nil
}

func (c Game) sendRequest(method string, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", serverApi, path), body)
	if err != nil {
		return nil, fmt.Errorf("failed sendRequest: %w", err)
	}
	req.Header.Add(tokenKey, c.Token)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed sendRequest: %w", err)
	}
	return res, nil
}

func unmarshalFromReadCloser[T any](rc *io.ReadCloser) (T, error) {
	defer (*rc).Close()
	t := new(T)
	bytes, err := io.ReadAll(*rc)
	if err != nil {
		return *t, fmt.Errorf("failed to read from Reader: %w", err)
	}
	err = json.Unmarshal(bytes, t)
	if err != nil {
		err = fmt.Errorf("failed to map value from JSON: %w", err)
	}
	return *t, err
}
