package providers

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	rimeBaseURL        = "https://users.rime.ai/v1/rime-tts"
	rimeDefaultSpeaker = "celeste"
	rimeDefaultModel   = "mistv2"
)

type rimeRequest struct {
	Text    string `json:"text"`
	Speaker string `json:"speaker"`
	ModelID string `json:"modelId"`
}

// Rime implements TTSProvider for the Rime AI TTS API.
type Rime struct{}

func (Rime) Call(config APIConfig, text string) (TTSStreamResult, error) {
	speaker := config.VoiceID
	if speaker == "" {
		speaker = rimeDefaultSpeaker
	}
	modelID := config.Model
	if modelID == "" {
		modelID = rimeDefaultModel
	}

	body := rimeRequest{
		Text:    text,
		Speaker: speaker,
		ModelID: modelID,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return TTSStreamResult{}, err
	}

	req, err := http.NewRequest(http.MethodPost, rimeBaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return TTSStreamResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	return MeasureStream(req)
}
