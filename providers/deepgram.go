package providers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

const (
	deepgramBaseURL      = "https://api.deepgram.com/v1/speak"
	deepgramDefaultModel = "aura-asteria-en"
)

type deepgramRequest struct {
	Text string `json:"text"`
}

// Deepgram implements TTSProvider for the Deepgram Aura API.
type Deepgram struct{}

func (Deepgram) Call(config APIConfig, text string) (TTSStreamResult, error) {
	model := config.Model
	if model == "" {
		model = deepgramDefaultModel
	}

	body := deepgramRequest{Text: text}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return TTSStreamResult{}, err
	}

	u, err := url.Parse(deepgramBaseURL)
	if err != nil {
		return TTSStreamResult{}, err
	}
	q := u.Query()
	q.Set("model", model)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		return TTSStreamResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+config.APIKey)

	return MeasureStream(req)
}
