package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const serverApi = "https://go-pjatk-server.fly.dev/api"
const tokenKey = "X-Auth-Token"

type Client struct {
	Token string
}

type GameSettings struct {
	Coords      []string `json:"coords"`
	Description string   `json:"description"`
	Nick        string   `json:"nick"`
	TargetNick  string   `json:"target_nick"`
	AgainstBot  bool     `json:"wpbot"`
}

func InitGame(settings GameSettings) (Client, error) {
	body, err := json.Marshal(settings)
	client := Client{}
	if err != nil {
		return client, fmt.Errorf("failed InitGame: %w", err)
	}
	r := bytes.NewReader(body)
	res, err := http.Post(serverApi+"/game", "application/json", r)
	if err != nil {
		return client, fmt.Errorf("failed InitGame: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode == 400 {
		return client, fmt.Errorf("failed InitGame: %s", res.Status)
	}
	token := res.Header.Get(tokenKey)
	c := Client{Token: token}
	return c, nil
}

func Board(c Client) ([]string, error) {
	res, err := c.sendRequest(http.MethodGet, "/game/board", nil)
	if err != nil {
		return nil, fmt.Errorf("failed Board: %w", err)
	}
	jsonBody := make(map[string][]string)
	err = unmarshalFromReadCloser(&res.Body, &jsonBody)

	return jsonBody["board"], err
}

func Status(c Client) error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/game", serverApi), nil)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	req.Header.Add(tokenKey, c.Token)

	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	body := res.Body
	bytes, _ := io.ReadAll(body)
	fmt.Println(string(bytes))
	return nil
}

func (c Client) sendRequest(method string, path string, body io.Reader) (*http.Response, error) {
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

func unmarshalFromReadCloser(rc *io.ReadCloser, value any) error {
	defer (*rc).Close()
	b, err := io.ReadAll(*rc)
	if err != nil {
		return fmt.Errorf("failed mapToJson: %w", err)
	}
	err = json.Unmarshal(b, value)
	if err != nil {
		err = fmt.Errorf("failed mapToJson: %w", err)
	}
	return err
}
