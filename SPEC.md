# TTS Latency Benchmarker — Project Spec

## Overview

A geo-distributed TTS (Text-to-Speech) API latency benchmarking tool. A single Go worker binary is deployed to multiple Fly.io regions. A lightweight frontend sends benchmark requests to each regional worker simultaneously and displays TTFB (Time to First Byte) and TTLB (Time to Last Byte) per provider per region.

Each API can be tested under **configurable concurrent load** — e.g. fire 10 simultaneous requests at ElevenLabs and 5 at OpenAI at the same time. Results include per-request timings plus aggregated statistics (min, max, mean, p50, p95) across all concurrent calls.

**Forked from:** `https://github.com/fly-apps/go-example`

---

## Architecture

```
[ Browser / Frontend ]
        |
        | POST /benchmark (with fly-prefer-region header)
        |---> bom.tts-benchmarker.fly.dev  (Mumbai)
        |---> fra.tts-benchmarker.fly.dev  (Frankfurt)
        |---> sin.tts-benchmarker.fly.dev  (Singapore)
        |---> iad.tts-benchmarker.fly.dev  (Virginia)
        |---> lax.tts-benchmarker.fly.dev  (Los Angeles)
        |
        | Each worker:
        |   1. Receives { text, apis[] }
        |   2. For each API, fires `concurrency` requests simultaneously
        |   3. Measures TTFB + TTLB per request using streaming reads
        |   4. Aggregates stats (min, max, mean, p50, p95) across concurrent calls
        |   5. Returns { region, results[] } — one result entry per API
```

---

## Repository Structure

```
tts-benchmarker/
├── main.go                        # HTTP server, routes, request handling
├── benchmark.go                   # Core timing logic, streaming reader
├── providers/
│   ├── types.go                   # TTSProvider interface + shared types
│   ├── elevenlabs.go              # ElevenLabs streaming adapter
│   ├── openai.go                  # OpenAI TTS adapter
│   ├── cartesia.go                # Cartesia adapter
│   └── deepgram.go                # Deepgram Aura adapter
├── templates/
│   └── index.html.tmpl            # Simple status page (shows FLY_REGION)
├── frontend/
│   └── index.html                 # Single-file frontend (vanilla JS)
├── fly.toml                       # Fly.io config
├── Dockerfile                     # Multi-stage Go build (keep from base repo)
├── go.mod
├── go.sum
└── SPEC.md                        # This file
```

---

## Core Types (`providers/types.go`)

```go
package providers

// SupportedProvider is the set of known TTS provider identifiers
type SupportedProvider string

const (
    ProviderElevenLabs SupportedProvider = "elevenlabs"
    ProviderOpenAI     SupportedProvider = "openai"
    ProviderCartesia   SupportedProvider = "cartesia"
    ProviderDeepgram   SupportedProvider = "deepgram"
)

// APIConfig holds user-supplied configuration for a single TTS API
type APIConfig struct {
    ID          string            `json:"id"`
    Name        string            `json:"name"`
    Provider    SupportedProvider `json:"provider"`
    APIKey      string            `json:"apiKey"`
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
```

---

## API Endpoints (`main.go`)

### `POST /benchmark`

**Request body:**
```json
{
  "text": "Hello, this is a test of text to speech latency.",
  "apis": [
    {
      "id": "el-1",
      "name": "ElevenLabs Turbo",
      "provider": "elevenlabs",
      "apiKey": "sk-...",
      "voiceId": "21m00Tcm4TlvDq8ikWAM",
      "model": "eleven_turbo_v2",
      "concurrency": 10
    },
    {
      "id": "oai-1",
      "name": "OpenAI TTS-1",
      "provider": "openai",
      "apiKey": "sk-...",
      "voiceId": "alloy",
      "model": "tts-1",
      "concurrency": 5
    }
  ]
}
```

**Response body:**
```json
{
  "region": "bom",
  "results": [
    {
      "apiConfigId": "el-1",
      "apiName": "ElevenLabs Turbo",
      "region": "bom",
      "concurrency": 10,
      "requests": [
        { "index": 0, "ttfb": 298.1, "ttlb": 1801.3, "audioSizeBytes": 48320 },
        { "index": 1, "ttfb": 334.5, "ttlb": 1923.7, "audioSizeBytes": 48320 },
        { "index": 2, "ttfb": 412.0, "ttlb": 2100.1, "audioSizeBytes": 48320 },
        "..."
      ],
      "ttfbStats": { "min": 298.1, "max": 601.4, "mean": 389.2, "p50": 371.0, "p95": 558.3 },
      "ttlbStats": { "min": 1801.3, "max": 2644.0, "mean": 2100.5, "p50": 2050.0, "p95": 2580.1 },
      "successCount": 10,
      "errorCount": 0
    }
  ]
}
```

**Behaviour:**
- Different APIs in the request are called **concurrently with each other** (goroutines per API)
- Within each API, `concurrency` requests are fired **simultaneously** — all start at the same instant using a `sync.WaitGroup`
- `concurrency` defaults to `1` if omitted or set to `0`
- `concurrency` is capped at `50` server-side to prevent abuse
- All concurrent calls for an API use the **same text and config** — the goal is load simulation, not variation
- Stats (min, max, mean, p50, p95) are computed over successful requests only (those without errors)
- If all requests for an API fail, `ttfbStats` and `ttlbStats` are zeroed and `errorCount` equals `concurrency`
- Region is read from `os.Getenv("FLY_REGION")`, defaults to `"local"` if unset
- Timeout per individual request: **30 seconds**
- CORS headers must be set to allow requests from the frontend origin

### `GET /health`
Returns `200 OK` with body `ok`. Used by Fly.io health checks.

### `GET /`
Returns the simple status page from `templates/index.html.tmpl` showing the current `FLY_REGION`. Keep this from the base repo.

---

## Timing Logic (`benchmark.go`)

### `MeasureStream` — times a single TTS API call

TTFB and TTLB are measured using Go's `http.Client` with streaming response body reads:

```go
// Pseudocode — implement in benchmark.go
func MeasureStream(req *http.Request) (TTSStreamResult, error) {
    start := time.Now()

    resp, err := http.DefaultClient.Do(req)
    // handle error

    var ttfb float64
    var totalBytes int64
    reader := resp.Body

    buf := make([]byte, 4096)
    for {
        n, err := reader.Read(buf)
        if n > 0 {
            if ttfb == 0 {
                ttfb = float64(time.Since(start).Milliseconds())
            }
            totalBytes += int64(n)
        }
        if err == io.EOF { break }
        // handle other errors
    }

    ttlb := float64(time.Since(start).Milliseconds())
    return TTSStreamResult{TTFB: ttfb, TTLB: ttlb, AudioSizeBytes: totalBytes}, nil
}
```

### `RunConcurrent` — fires N simultaneous calls and aggregates results

```go
// Pseudocode — implement in benchmark.go
func RunConcurrent(provider TTSProvider, config APIConfig, text string) BenchmarkResult {
    n := config.Concurrency
    if n < 1 { n = 1 }
    if n > 50 { n = 50 }

    requestResults := make([]RequestResult, n)
    var wg sync.WaitGroup

    // Launch all N goroutines at the same time
    for i := 0; i < n; i++ {
        wg.Add(1)
        go func(idx int) {
            defer wg.Done()
            result, err := provider.Call(config, text)
            if err != nil {
                requestResults[idx] = RequestResult{Index: idx, Error: err.Error()}
            } else {
                requestResults[idx] = RequestResult{
                    Index: idx,
                    TTFB:  result.TTFB,
                    TTLB:  result.TTLB,
                    AudioSizeBytes: result.AudioSizeBytes,
                }
            }
        }(i)
    }
    wg.Wait()

    return BenchmarkResult{
        // ... populate from requestResults
        TTFBStats: computeStats(successfulTTFBs),
        TTLBStats: computeStats(successfulTTLBs),
    }
}
```

### `computeStats` — calculates min, max, mean, p50, p95

```go
// Pseudocode
func computeStats(values []float64) Stats {
    // sort values
    // compute min, max, mean
    // p50 = values[len*0.50]
    // p95 = values[len*0.95]
    // return Stats{...}
}
```

**Important:** p50 and p95 are computed by sorting the slice and taking the value at the relevant index (nearest rank method). With small N (e.g. concurrency=5), p95 will equal the max — that is expected and correct.

---

## Provider Implementations

### ElevenLabs (`providers/elevenlabs.go`)
- **Endpoint:** `POST https://api.elevenlabs.io/v1/text-to-speech/{voiceId}/stream`
- **Auth header:** `xi-api-key: {apiKey}`
- **Default voice:** `21m00Tcm4TlvDq8ikWAM`
- **Default model:** `eleven_turbo_v2`
- **Request body:**
  ```json
  {
    "text": "...",
    "model_id": "eleven_turbo_v2",
    "voice_settings": { "stability": 0.5, "similarity_boost": 0.75 }
  }
  ```
- **Streaming:** Yes — use `MeasureStream`

### OpenAI (`providers/openai.go`)
- **Endpoint:** `POST https://api.openai.com/v1/audio/speech`
- **Auth header:** `Authorization: Bearer {apiKey}`
- **Default voice:** `alloy`
- **Default model:** `tts-1`
- **Request body:**
  ```json
  { "model": "tts-1", "input": "...", "voice": "alloy" }
  ```
- **Streaming:** Yes — use `MeasureStream`

### Cartesia (`providers/cartesia.go`)
- **Endpoint:** `POST https://api.cartesia.ai/tts/bytes`
- **Auth header:** `X-API-Key: {apiKey}`
- **Extra header:** `Cartesia-Version: 2024-06-10`
- **Default model:** `sonic-english`
- **Request body:**
  ```json
  {
    "transcript": "...",
    "model_id": "sonic-english",
    "voice": { "mode": "id", "id": "{voiceId}" },
    "output_format": { "container": "raw", "encoding": "pcm_f32le", "sample_rate": 44100 }
  }
  ```
- **Streaming:** Yes — use `MeasureStream`

### Deepgram (`providers/deepgram.go`)
- **Endpoint:** `POST https://api.deepgram.com/v1/speak?model={model}`
- **Auth header:** `Authorization: Token {apiKey}`
- **Default model:** `aura-asteria-en`
- **Request body:** `{ "text": "..." }`
- **Streaming:** Yes — use `MeasureStream`

---

## Fly.io Configuration (`fly.toml`)

```toml
app = 'tts-benchmarker'
primary_region = 'bom'

[build]

[env]
  PORT = '8080'

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 0
  processes = ['app']

[[vm]]
  cpu_kind = 'shared'
  cpus = 1
  memory_mb = 256
```

**Deploy to multiple regions after initial deploy:**
```bash
fly scale count 1 --region bom   # Mumbai
fly scale count 1 --region fra   # Frankfurt
fly scale count 1 --region sin   # Singapore
fly scale count 1 --region iad   # Virginia
fly scale count 1 --region lax   # Los Angeles
fly scale count 1 --region syd   # Sydney
```

**To force requests to a specific region from the frontend**, set the header:
```
fly-prefer-region: fra
```

---

## Frontend (`frontend/index.html`)

A **single self-contained HTML file** with vanilla JS. No build step, no framework.

### UI Sections

1. **API Config Panel**
   - Add/remove API cards
   - Each card: Provider dropdown, Display name, API Key (password input), Voice ID, Model, **Concurrency** (number input, 1–50, default 1)
   - The concurrency field is per-API — different APIs can have different values
   - "Add API" button

2. **Test Config**
   - Text input (textarea, pre-filled with a default sentence)
   - Region checkboxes: Mumbai, Frankfurt, Singapore, Virginia, Los Angeles, Sydney
   - "Run Benchmark" button

3. **Results**

   **Per-request table** (expandable / collapsible per API):
   - Columns: # | TTFB (ms) | TTLB (ms) | Audio Size | Status
   - Shows all individual concurrent request results for each API

   **Summary table** (always visible):
   - Columns: API Name | Region | Concurrency | TTFB min | TTFB p50 | TTFB p95 | TTFB max | TTLB mean | Success | Errors
   - Color coding on TTFB p50: green < 300ms, amber 300–700ms, red > 700ms
   - Each row shows a loading spinner while in-flight

### Key JS Logic

```javascript
// Pseudocode
async function runBenchmark() {
  const requests = selectedRegions.map(region =>
    fetch('https://tts-benchmarker.fly.dev/benchmark', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'fly-prefer-region': region.code
      },
      body: JSON.stringify({ text, apis })  // apis includes concurrency per entry
    }).then(r => r.json())
  );

  // Fire all regions in parallel, update table as each resolves
  results = await Promise.allSettled(requests);
  renderResults(results);
}
```

### Region Map

```javascript
const REGIONS = [
  { code: 'bom', label: 'Mumbai 🇮🇳' },
  { code: 'fra', label: 'Frankfurt 🇩🇪' },
  { code: 'sin', label: 'Singapore 🇸🇬' },
  { code: 'iad', label: 'Virginia 🇺🇸' },
  { code: 'lax', label: 'Los Angeles 🇺🇸' },
  { code: 'syd', label: 'Sydney 🇦🇺' },
];
```

---

## Environment & Secrets

No secrets are stored on the server. API keys are passed in the request body from the frontend and used only for the duration of the benchmark call. They are never logged or persisted.

The only environment variable the worker reads is `FLY_REGION` (set automatically by Fly.io).

---

## CORS

The `/benchmark` endpoint must return:
```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: POST, OPTIONS
Access-Control-Allow-Headers: Content-Type, fly-prefer-region
```

Handle `OPTIONS` preflight requests with a `200` response.

---

## Error Handling

- If an individual request within a concurrent batch fails (non-2xx, timeout, network error), record it as a `RequestResult` with `error` set — do not abort the other concurrent requests
- `errorCount` reflects how many of the `concurrency` requests failed; `successCount` reflects how many succeeded
- Stats are computed only over successful requests. If `successCount` is 0, stats fields are all `0`
- If the request body is malformed, return `400` with a JSON error message
- Never return `500` from `/benchmark` — always return `200` with per-API errors embedded in results
- Each goroutine must use `recover()` to prevent a panic in one request from crashing the worker

---

## What To Keep From `fly-apps/go-example`

- `Dockerfile` — keep as-is, it's a clean multi-stage Go build
- `go.mod` / `go.sum` — update module name to `tts-benchmarker`
- `templates/index.html.tmpl` — keep, update to show region + "worker is running" status
- `fly.toml` — replace contents with config above

**Replace entirely:**
- `main.go` — replace with new server logic

**Add new:**
- `benchmark.go`
- `providers/` directory
- `frontend/index.html`
