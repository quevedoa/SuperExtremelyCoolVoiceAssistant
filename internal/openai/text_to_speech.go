package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func (c *Client) ConvertTextToSpeech(ctx context.Context, text string, w io.Writer) error {
	reqBody := map[string]interface{}{
		"model":        "gpt-4o-mini-tts",
		"input":        text,
		"voice":        "ballad",
		"instructions": "Voice Affect: Calm, composed, and reassuring; project quiet authority and confidence. Tone: Sincere, empathetic, and gently authoritative—express genuine apology while conveying competence.",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		baseURL+"/audio/speech",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	return c.streamResponseToWriter(ctx, req, w)
}
