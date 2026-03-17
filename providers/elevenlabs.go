package providers

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	elevenLabsBaseURL   = "https://api.elevenlabs.io/v1/text-to-speech"
	elevenLabsDefaultVoice = "21m00Tcm4TlvDq8ikWAM"
	elevenLabsDefaultModel = "eleven_turbo_v2"
)

type elevenLabsRequest struct {
	Text          string              `json:"text"`
	ModelID       string              `json:"model_id"`
	VoiceSettings elevenLabsVoiceSettings `json:"voice_settings"`
}

type elevenLabsVoiceSettings struct {
	Stability       float64 `json:"stability"`
	SimilarityBoost float64 `json:"similarity_boost"`
}

// ElevenLabs implements TTSProvider for the ElevenLabs API.
type ElevenLabs struct{}

func (ElevenLabs) Call(config APIConfig, text string) (TTSStreamResult, error) {
	voiceID := config.VoiceID
	if voiceID == "" {
		voiceID = elevenLabsDefaultVoice
	}
	model := config.Model
	if model == "" {
		model = elevenLabsDefaultModel
	}

	body := elevenLabsRequest{
		Text:    text,
		ModelID: model,
		VoiceSettings: elevenLabsVoiceSettings{
			Stability:       0.5,
			SimilarityBoost: 0.75,
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return TTSStreamResult{}, err
	}

	url := elevenLabsBaseURL + "/" + voiceID + "/stream"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return TTSStreamResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", config.APIKey)

	return MeasureStream(req)
}
