package openai

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
)

type WhisperResponse struct {
	Text string `json:"text"`
}

func (c *Client) TranscribeRawWAV(ctx context.Context, wavReader io.Reader) (*WhisperResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "recording.wav")
	if err != nil {
		return nil, err
	}

	if _, err := io.Copy(part, wavReader); err != nil {
		return nil, err
	}

	if err := writer.WriteField("model", "whisper-1"); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/audio/transcriptions",
		body,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	var res WhisperResponse
	if err := c.doRequest(ctx, req, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
