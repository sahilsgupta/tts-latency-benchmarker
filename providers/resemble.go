package providers

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	resembleStreamURL = "https://f.cluster.resemble.ai/stream"
)

type resembleRequest struct {
	VoiceUUID string `json:"voice_uuid"`
	Data     string `json:"data"`
	Model    string `json:"model,omitempty"`
}

// Resemble implements TTSProvider for the Resemble AI streaming TTS API.
// VoiceID must be a Resemble voice UUID (required).
type Resemble struct{}

func (Resemble) Call(config APIConfig, text string) (TTSStreamResult, error) {
	voiceUUID := config.VoiceID
	if voiceUUID == "" {
		return TTSStreamResult{}, &ResembleConfigError{Msg: "Resemble requires a voice UUID (voiceId)"}
	}

	body := resembleRequest{
		VoiceUUID: voiceUUID,
		Data:     text,
		Model:    config.Model, // e.g. chatterbox-turbo
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return TTSStreamResult{}, err
	}

	req, err := http.NewRequest(http.MethodPost, resembleStreamURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return TTSStreamResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	return MeasureStream(req)
}

// ResembleConfigError indicates missing Resemble config.
type ResembleConfigError struct {
	Msg string
}

func (e *ResembleConfigError) Error() string {
	return e.Msg
}
