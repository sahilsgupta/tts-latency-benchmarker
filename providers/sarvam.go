package providers

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	sarvamBaseURL           = "https://api.sarvam.ai/text-to-speech"
	sarvamDefaultSpeaker    = "shubh"
	sarvamDefaultModel      = "bulbul:v3"
	sarvamDefaultLang       = "en-IN"
)

type sarvamRequest struct {
	Text               string `json:"text"`
	TargetLanguageCode string `json:"target_language_code"`
	Speaker            string `json:"speaker"`
	Model              string `json:"model"`
}

// Sarvam implements TTSProvider for the Sarvam AI TTS REST API.
// Response is JSON with base64 audio; we measure TTFB/TTLB on the HTTP response.
type Sarvam struct{}

func (Sarvam) Call(config APIConfig, text string) (TTSStreamResult, error) {
	speaker := config.VoiceID
	if speaker == "" {
		speaker = sarvamDefaultSpeaker
	}
	model := config.Model
	if model == "" {
		model = sarvamDefaultModel
	}

	body := sarvamRequest{
		Text:               text,
		TargetLanguageCode: sarvamDefaultLang,
		Speaker:            speaker,
		Model:              model,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return TTSStreamResult{}, err
	}

	req, err := http.NewRequest(http.MethodPost, sarvamBaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return TTSStreamResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-subscription-key", config.APIKey)

	return MeasureStream(req)
}
