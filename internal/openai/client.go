package openai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const baseURL = "https://api.openai.com/v1"

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) doRequest(ctx context.Context, req *http.Request, out interface{}) error {
	req = req.WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		data, _ := io.ReadAll(res.Body)
		return &HTTPError{
			StatusCode: res.StatusCode,
			Body:       string(data),
		}
	}
	if out != nil {
		return json.NewDecoder(res.Body).Decode(out)
	}

	return nil
}

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return "openai error: status=" + http.StatusText(e.StatusCode) + " body=" + e.Body
}
