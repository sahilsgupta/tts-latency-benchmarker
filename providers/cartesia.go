package providers

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	cartesiaBaseURL      = "https://api.cartesia.ai/tts/bytes"
	cartesiaDefaultModel = "sonic-3"
	cartesiaVersion      = "2025-04-16"
	// Default voice (Katie) when none provided; Cartesia requires a valid UUID.
	cartesiaDefaultVoiceID = "f786b574-daa5-4673-aa0c-cbe3e8534c02"
)

type cartesiaRequest struct {
	Transcript   string              `json:"transcript"`
	ModelID      string              `json:"model_id"`
	Voice        cartesiaVoice       `json:"voice"`
	OutputFormat cartesiaOutputFormat `json:"output_format"`
}

type cartesiaVoice struct {
	Mode string `json:"mode"`
	ID   string `json:"id"`
}

type cartesiaOutputFormat struct {
	Container  string `json:"container"`
	Encoding   string `json:"encoding"`
	SampleRate int    `json:"sample_rate"`
}

// Cartesia implements TTSProvider for the Cartesia TTS API.
type Cartesia struct{}

func (Cartesia) Call(config APIConfig, text string) (TTSStreamResult, error) {
	voiceID := config.VoiceID
	if voiceID == "" {
		voiceID = cartesiaDefaultVoiceID
	}
	model := config.Model
	if model == "" {
		model = cartesiaDefaultModel
	}

	body := cartesiaRequest{
		Transcript: text,
		ModelID:    model,
		Voice: cartesiaVoice{
			Mode: "id",
			ID:   voiceID,
		},
		OutputFormat: cartesiaOutputFormat{
			Container:  "raw",
			Encoding:   "pcm_f32le",
			SampleRate: 44100,
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return TTSStreamResult{}, err
	}

	req, err := http.NewRequest(http.MethodPost, cartesiaBaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return TTSStreamResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	req.Header.Set("Cartesia-Version", cartesiaVersion)

	return MeasureStream(req)
}
