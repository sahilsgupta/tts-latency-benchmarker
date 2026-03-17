package providers

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	googleSynthesizeURL   = "https://texttospeech.googleapis.com/v1/text:synthesize"
	googleDefaultVoice    = "en-US-Neural2-A"
	googleDefaultLang     = "en-US"
	googleDefaultEncoding = "MP3"
	googleDefaultSample   = 24000
)

// Google implements TTSProvider for Google Cloud Text-to-Speech API.
type Google struct{}

type googleRequest struct {
	Input      googleInput      `json:"input"`
	Voice      googleVoice      `json:"voice"`
	AudioConfig googleAudioConfig `json:"audioConfig"`
}

type googleInput struct {
	Text string `json:"text"`
}

type googleVoice struct {
	Name         string `json:"name"`
	LanguageCode string `json:"languageCode"`
}

type googleAudioConfig struct {
	AudioEncoding   string `json:"audioEncoding"`
	SampleRateHertz int    `json:"sampleRateHertz"`
}

// Call synthesizes speech. APIKey should be a Bearer access token (e.g. gcloud auth print-access-token).
func (Google) Call(config APIConfig, text string) (TTSStreamResult, error) {
	voiceName := config.VoiceID
	if voiceName == "" {
		voiceName = googleDefaultVoice
	}
	lang := config.Model
	if lang == "" {
		lang = googleDefaultLang
	}

	body := googleRequest{
		Input: googleInput{Text: text},
		Voice: googleVoice{
			Name:         voiceName,
			LanguageCode: lang,
		},
		AudioConfig: googleAudioConfig{
			AudioEncoding:   googleDefaultEncoding,
			SampleRateHertz: googleDefaultSample,
		},
	}
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return TTSStreamResult{}, err
	}

	req, err := http.NewRequest(http.MethodPost, googleSynthesizeURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return TTSStreamResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+config.APIKey)

	return MeasureStream(req)
}
