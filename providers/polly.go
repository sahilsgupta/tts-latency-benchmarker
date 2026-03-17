package providers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

const (
	pollyDefaultRegion   = "us-east-1"
	pollyDefaultVoice   = "Joanna"
	pollyDefaultEngine  = "neural"
	pollyDefaultFormat  = "mp3"
	pollyDefaultSampleRate = "24000"
)

// pollyEndpoint returns the Polly REST endpoint for the given region.
func pollyEndpoint(region string) string {
	if region == "" {
		region = pollyDefaultRegion
	}
	return "https://polly." + region + ".amazonaws.com/v1/speech"
}

type pollyRequest struct {
	Engine       string `json:"Engine,omitempty"`
	Text         string `json:"Text"`
	VoiceID      string `json:"VoiceId"`
	OutputFormat string `json:"OutputFormat"`
	SampleRate   string `json:"SampleRate,omitempty"`
	TextType     string `json:"TextType,omitempty"`
}

// Polly implements TTSProvider for the Amazon Polly API.
// APIConfig.APIKey = AWS Access Key ID, APIConfig.APISecret = AWS Secret Access Key.
type Polly struct{}

func (Polly) Call(config APIConfig, text string) (TTSStreamResult, error) {
	region := pollyDefaultRegion
	voiceID := config.VoiceID
	if voiceID == "" {
		voiceID = pollyDefaultVoice
	}
	engine := config.Model
	if engine == "" {
		engine = pollyDefaultEngine
	}
	// Optional: use config.Model as "engine:region" for flexibility
	if parts := strings.SplitN(config.Model, ":", 2); len(parts) == 2 {
		engine = parts[0]
		region = parts[1]
	}

	body := pollyRequest{
		Engine:       engine,
		Text:         text,
		VoiceID:      voiceID,
		OutputFormat: pollyDefaultFormat,
		SampleRate:   pollyDefaultSampleRate,
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return TTSStreamResult{}, err
	}

	url := pollyEndpoint(region)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return TTSStreamResult{}, err
	}

	if config.APISecret == "" {
		return TTSStreamResult{}, &PollyConfigError{Msg: "Polly requires apiSecret (AWS Secret Access Key)"}
	}

	err = SignPollyRequest(req, bodyBytes, config.APIKey, config.APISecret, region)
	if err != nil {
		return TTSStreamResult{}, err
	}

	return MeasureStream(req)
}

// PollyConfigError indicates missing or invalid Polly/AWS config.
type PollyConfigError struct {
	Msg string
}

func (e *PollyConfigError) Error() string {
	return e.Msg
}
