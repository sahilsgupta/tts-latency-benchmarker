package e2e_test

import (
	"os"
	"strings"
	"testing"

	"tts-benchmarker/providers"
)

const e2ePhrase = "This is a latency test."

func envKey(t *testing.T, name string) string {
	t.Helper()
	v := strings.TrimSpace(os.Getenv(name))
	if v == "" {
		t.Skipf("skip: %s not set or empty", name)
	}
	return v
}

func assertLatency(t *testing.T, name string, res providers.TTSStreamResult, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	t.Logf("%s: TTFB=%.2f ms TTLB=%.2f ms AudioSizeBytes=%d", name, res.TTFB, res.TTLB, res.AudioSizeBytes)
	if res.TTFB <= 0 {
		t.Errorf("%s: want TTFB > 0, got %.2f", name, res.TTFB)
	}
	if res.TTLB <= res.TTFB {
		t.Errorf("%s: want TTLB > TTFB, got TTFB=%.2f TTLB=%.2f", name, res.TTFB, res.TTLB)
	}
	if res.AudioSizeBytes <= 0 {
		t.Errorf("%s: want AudioSizeBytes > 0, got %d", name, res.AudioSizeBytes)
	}
}

func TestE2E_ElevenLabs(t *testing.T) {
	key := envKey(t, "ELEVENLABS_API_KEY")
	cfg := providers.APIConfig{
		APIKey:  key,
		VoiceID: "21m00Tcm4TlvDq8ikWAM",
		Model:   "eleven_flash_v2_5",
	}
	res, err := providers.ElevenLabs{}.Call(cfg, e2ePhrase)
	assertLatency(t, "ElevenLabs", res, err)
}

func TestE2E_OpenAI(t *testing.T) {
	key := envKey(t, "OPENAI_API_KEY")
	cfg := providers.APIConfig{APIKey: key, VoiceID: "alloy", Model: "tts-1"}
	res, err := providers.OpenAI{}.Call(cfg, e2ePhrase)
	assertLatency(t, "OpenAI", res, err)
}

func TestE2E_Cartesia(t *testing.T) {
	key := envKey(t, "CARTESIA_API_KEY")
	cfg := providers.APIConfig{
		APIKey:  key,
		VoiceID: "f786b574-daa5-4673-aa0c-cbe3e8534c02",
		Model:   "sonic-3",
	}
	res, err := providers.Cartesia{}.Call(cfg, e2ePhrase)
	assertLatency(t, "Cartesia", res, err)
}

func TestE2E_Deepgram(t *testing.T) {
	key := envKey(t, "DEEPGRAM_API_KEY")
	cfg := providers.APIConfig{APIKey: key, Model: "aura-asteria-en"}
	res, err := providers.Deepgram{}.Call(cfg, e2ePhrase)
	assertLatency(t, "Deepgram", res, err)
}

func TestE2E_Murf(t *testing.T) {
	key := envKey(t, "MURF_API_KEY")
	cfg := providers.APIConfig{APIKey: key, VoiceID: "Matthew", Model: "FALCON"}
	res, err := providers.Murf{}.Call(cfg, e2ePhrase)
	assertLatency(t, "Murf", res, err)
}

func TestE2E_Rime(t *testing.T) {
	key := envKey(t, "RIME_API_KEY")
	cfg := providers.APIConfig{APIKey: key, VoiceID: "celeste", Model: "mistv2"}
	res, err := providers.Rime{}.Call(cfg, e2ePhrase)
	assertLatency(t, "Rime", res, err)
}

func TestE2E_Sarvam(t *testing.T) {
	key := envKey(t, "SARVAM_API_KEY")
	cfg := providers.APIConfig{APIKey: key, VoiceID: "shubh", Model: "bulbul:v3"}
	res, err := providers.Sarvam{}.Call(cfg, e2ePhrase)
	assertLatency(t, "Sarvam", res, err)
}

func TestE2E_Google(t *testing.T) {
	token := envKey(t, "GOOGLE_TTS_ACCESS_TOKEN")
	cfg := providers.APIConfig{
		APIKey:  token,
		VoiceID: "en-US-Neural2-A",
		Model:   "en-US",
	}
	res, err := providers.Google{}.Call(cfg, e2ePhrase)
	assertLatency(t, "Google", res, err)
}

func TestE2E_Resemble(t *testing.T) {
	key := envKey(t, "RESEMBLE_API_KEY")
	voice := envKey(t, "RESEMBLE_VOICE_UUID")
	cfg := providers.APIConfig{APIKey: key, VoiceID: voice, Model: "chatterbox-turbo"}
	res, err := providers.Resemble{}.Call(cfg, e2ePhrase)
	assertLatency(t, "Resemble", res, err)
}

func TestE2E_Polly(t *testing.T) {
	ak := envKey(t, "POLLY_AWS_ACCESS_KEY_ID")
	sk := envKey(t, "POLLY_AWS_SECRET_ACCESS_KEY")
	cfg := providers.APIConfig{
		APIKey:    ak,
		APISecret: sk,
		VoiceID:   "Joanna",
		Model:     "neural",
	}
	res, err := providers.Polly{}.Call(cfg, e2ePhrase)
	assertLatency(t, "Polly", res, err)
}
