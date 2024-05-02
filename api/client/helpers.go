package client

import (
	"encoding/json"
	"fmt"
	"io"
)

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
