package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type ContentItem struct {
	Type        string        `json:"type"`
	Text        string        `json:"text,omitempty"`
	Annotations []interface{} `json:"annotations"`
}

type OutputItem struct {
	ID      string        `json:"id"`
	Type    string        `json:"type"`
	Role    string        `json:"role"`
	Content []ContentItem `json:"content"`
}

type ChatResponse struct {
	Output []OutputItem `json:"output"`
}

func (c *Client) GetLLMResponse(ctx context.Context, prompt string) (*string, error) {
	reqBody := map[string]interface{}{
		"model": "gpt-4.1",
		"tools": []map[string]string{
			{"type": "web_search_preview"},
		},
		"instructions": "You are a helpful assistant who will respond using speech, due to that make your responses short and conversation-like, don't include links in your text response and responses should be less than 20 words.",
		"input":        prompt,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/responses",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	var res ChatResponse
	if err := c.doRequest(ctx, req, &res); err != nil {
		return nil, err
	}

	var textOutput string
	for _, o := range res.Output {
		if len(o.Content) != 0 {
			textOutput = o.Content[0].Text
		}
	}

	return &textOutput, nil
}
