package utils

import (
	"context"
	"io"
	"net/http"
)

func Get(ctx context.Context, path string, headers ...map[string]string) ([]byte, error) {
	cli := http.Client{}

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}

	response, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	res, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	return res, nil
}
