package providers

// SupportedProvider is the set of known TTS provider identifiers
type SupportedProvider string

const (
	ProviderElevenLabs SupportedProvider = "elevenlabs"
	ProviderOpenAI     SupportedProvider = "openai"
	ProviderCartesia   SupportedProvider = "cartesia"
	ProviderDeepgram   SupportedProvider = "deepgram"
	ProviderMurf       SupportedProvider = "murf"
	ProviderPolly      SupportedProvider = "polly"
)

// APIConfig holds user-supplied configuration for a single TTS API
type APIConfig struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Provider    SupportedProvider `json:"provider"`
	APIKey      string            `json:"apiKey"`
	APISecret   string            `json:"apiSecret,omitempty"` // used by AWS Polly (secret access key)
	VoiceID     string            `json:"voiceId,omitempty"`
	Model       string            `json:"model,omitempty"`
	Concurrency int               `json:"concurrency"` // number of simultaneous requests; defaults to 1
}

// RequestResult holds timing for a single individual request within a concurrent run
type RequestResult struct {
	Index          int     `json:"index"`           // 0-based index of this request in the batch
	TTFB           float64 `json:"ttfb"`            // milliseconds
	TTLB           float64 `json:"ttlb"`            // milliseconds
	AudioSizeBytes int64   `json:"audioSizeBytes"`
	Error          string  `json:"error,omitempty"`
}

// Stats holds aggregate statistics computed across all concurrent RequestResults
type Stats struct {
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
	Mean float64 `json:"mean"`
	P50  float64 `json:"p50"`
	P95  float64 `json:"p95"`
}

// BenchmarkResult holds timing results for a single API across all concurrent calls
type BenchmarkResult struct {
	APIConfigID    string          `json:"apiConfigId"`
	APIName        string          `json:"apiName"`
	Region         string          `json:"region"`
	Concurrency    int             `json:"concurrency"`
	Requests       []RequestResult `json:"requests"`   // one entry per concurrent call
	TTFBStats      Stats           `json:"ttfbStats"`  // aggregated across successful requests
	TTLBStats      Stats           `json:"ttlbStats"`  // aggregated across successful requests
	SuccessCount   int             `json:"successCount"`
	ErrorCount     int             `json:"errorCount"`
}

// TTSProvider is the interface every provider adapter must implement
type TTSProvider interface {
	Call(config APIConfig, text string) (TTSStreamResult, error)
}

// TTSStreamResult is the raw timing data returned by a single provider call
type TTSStreamResult struct {
	TTFB           float64
	TTLB           float64
	AudioSizeBytes int64
}
