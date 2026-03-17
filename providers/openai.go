package providers

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	openAIBaseURL        = "https://api.openai.com/v1/audio/speech"
	openAIDefaultVoice   = "alloy"
	openAIDefaultModel   = "tts-1"
)

type openAIRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
	Voice string `json:"voice"`
}

// OpenAI implements TTSProvider for the OpenAI TTS API.
type OpenAI struct{}

func (OpenAI) Call(config APIConfig, text string) (TTSStreamResult, error) {
	voice := config.VoiceID
	if voice == "" {
		voice = openAIDefaultVoice
	}
	model := config.Model
	if model == "" {
		model = openAIDefaultModel
	}

	body := openAIRequest{
		Model: model,
		Input: text,
		Voice: voice,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return TTSStreamResult{}, err
	}

	req, err := http.NewRequest(http.MethodPost, openAIBaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return TTSStreamResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	return MeasureStream(req)
}
