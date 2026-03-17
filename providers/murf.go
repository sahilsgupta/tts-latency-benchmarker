package providers

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	murfStreamURL       = "https://global.api.murf.ai/v1/speech/stream"
	murfDefaultModel    = "FALCON"
	murfDefaultLocale   = "en-US"
	murfDefaultFormat   = "WAV"
	murfDefaultSampleRate = 24000
)

type murfStreamRequest struct {
	Text      string  `json:"text"`
	VoiceID   string  `json:"voiceId"`
	Model     string  `json:"model,omitempty"`
	Locale    string  `json:"locale,omitempty"`
	Format    string  `json:"format,omitempty"`
	SampleRate float64 `json:"sampleRate,omitempty"`
}

// Murf implements TTSProvider for the Murf Falcon streaming API.
type Murf struct{}

func (Murf) Call(config APIConfig, text string) (TTSStreamResult, error) {
	voiceID := config.VoiceID
	if voiceID == "" {
		voiceID = "Matthew"
	}
	model := config.Model
	if model == "" {
		model = murfDefaultModel
	}

	body := murfStreamRequest{
		Text:       text,
		VoiceID:    voiceID,
		Model:      model,
		Locale:     murfDefaultLocale,
		Format:     murfDefaultFormat,
		SampleRate: murfDefaultSampleRate,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return TTSStreamResult{}, err
	}

	req, err := http.NewRequest(http.MethodPost, murfStreamURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return TTSStreamResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", config.APIKey)

	return MeasureStream(req)
}
