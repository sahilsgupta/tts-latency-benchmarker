package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tts-benchmarker/providers"
)

func TestHandleHealth(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	handleHealth(w, r)
	if w.Code != http.StatusOK || w.Body.String() != "ok" {
		t.Fatalf("health: %d %q", w.Code, w.Body.String())
	}
}

func TestHandleRoot(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handleRoot(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("root status %d", w.Code)
	}
	body := w.Body.String()
	if !bytes.Contains([]byte(body), []byte("Hello from Fly")) {
		t.Fatalf("root body missing title: %s", body)
	}
}

func TestHandleRoot_notFound(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/other", nil)
	handleRoot(w, r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func TestHandleApp(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/app", nil)
	handleApp(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("app status %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Fatalf("content-type: %s", w.Header().Get("Content-Type"))
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("TTS Latency Benchmarker")) {
		t.Fatal("app HTML missing title")
	}
}

func TestHandleBenchmark_options(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodOptions, "/benchmark", nil)
	handleBenchmark(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("options %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("missing CORS")
	}
}

func TestHandleBenchmark_methodNotAllowed(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/benchmark", nil)
	handleBenchmark(w, r)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("GET benchmark want 405, got %d", w.Code)
	}
}

func TestHandleBenchmark_invalidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/benchmark", bytes.NewReader([]byte("not json")))
	handleBenchmark(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestHandleBenchmark_emptyApis(t *testing.T) {
	w := httptest.NewRecorder()
	body := `{"text":"hi","apis":[]}`
	r := httptest.NewRequest(http.MethodPost, "/benchmark", bytes.NewReader([]byte(body)))
	handleBenchmark(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestHandleBenchmark_unknownProvider(t *testing.T) {
	w := httptest.NewRecorder()
	reqBody := map[string]any{
		"text": "hello",
		"apis": []map[string]any{{
			"id": "1", "name": "X", "provider": "not-a-real-provider", "apiKey": "k",
		}},
	}
	b, _ := json.Marshal(reqBody)
	r := httptest.NewRequest(http.MethodPost, "/benchmark", bytes.NewReader(b))
	handleBenchmark(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("benchmark should return 200 with embedded errors, got %d", w.Code)
	}
	var resp benchmarkResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 1 || resp.Results[0].Requests[0].Error == "" {
		t.Fatalf("expected error in result: %+v", resp.Results)
	}
}

type benchMockProvider struct{}

func (benchMockProvider) Call(_ providers.APIConfig, _ string) (providers.TTSStreamResult, error) {
	return providers.TTSStreamResult{TTFB: 10, TTLB: 20, AudioSizeBytes: 500}, nil
}

func TestHandleBenchmark_successWithMockProvider(t *testing.T) {
	old := resolveProvider
	resolveProvider = func(providers.SupportedProvider) (providers.TTSProvider, bool) {
		return benchMockProvider{}, true
	}
	defer func() { resolveProvider = old }()

	w := httptest.NewRecorder()
	reqBody := map[string]any{
		"text": "hello world",
		"apis": []map[string]any{{
			"id": "1", "name": "Mock", "provider": "openai", "apiKey": "x",
			"concurrency": 2,
		}},
	}
	b, _ := json.Marshal(reqBody)
	r := httptest.NewRequest(http.MethodPost, "/benchmark", bytes.NewReader(b))
	handleBenchmark(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d body %s", w.Code, w.Body.String())
	}
	var resp benchmarkResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 1 {
		t.Fatalf("results len %d", len(resp.Results))
	}
	res := resp.Results[0]
	if res.SuccessCount != 2 || res.ErrorCount != 0 {
		t.Fatalf("success/err: %d/%d", res.SuccessCount, res.ErrorCount)
	}
	if len(res.Requests) != 2 {
		t.Fatalf("requests %d", len(res.Requests))
	}
}

func TestHandleBenchmark_regionFromEnv(t *testing.T) {
	old := resolveProvider
	resolveProvider = func(providers.SupportedProvider) (providers.TTSProvider, bool) {
		return benchMockProvider{}, true
	}
	defer func() { resolveProvider = old }()
	t.Setenv("FLY_REGION", "test-region")
	// getRegion in RunConcurrent reads env at runtime
	w := httptest.NewRecorder()
	reqBody := map[string]any{
		"text": "x",
		"apis": []map[string]any{{"id": "1", "name": "M", "provider": "openai", "apiKey": "k"}},
	}
	b, _ := json.Marshal(reqBody)
	r := httptest.NewRequest(http.MethodPost, "/benchmark", bytes.NewReader(b))
	handleBenchmark(w, r)
	var resp benchmarkResponse
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp.Region != "test-region" {
		t.Fatalf("region want test-region, got %q", resp.Region)
	}
}
