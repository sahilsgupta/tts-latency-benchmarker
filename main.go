package main

import (
	"embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"

	"tts-benchmarker/providers"
)

//go:embed templates/*
var resources embed.FS

//go:embed frontend/index.html
var frontendFS embed.FS

var t = template.Must(template.ParseFS(resources, "templates/*"))

func main() {
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/app", handleApp)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/benchmark", handleBenchmark)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("listening on", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	data := map[string]string{
		"Region": os.Getenv("FLY_REGION"),
	}
	if data["Region"] == "" {
		data["Region"] = "local"
	}
	_ = t.ExecuteTemplate(w, "index.html.tmpl", data)
}

func handleApp(w http.ResponseWriter, r *http.Request) {
	data, err := frontendFS.ReadFile("frontend/index.html")
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

type benchmarkRequest struct {
	Text string          `json:"text"`
	APIs []providers.APIConfig `json:"apis"`
}

type benchmarkResponse struct {
	Region  string                  `json:"region"`
	Results []providers.BenchmarkResult `json:"results"`
}

func handleBenchmark(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		setCORS(w)
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		setCORS(w)
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	setCORS(w)

	var req benchmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid JSON"})
		return
	}
	if len(req.APIs) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "apis required"})
		return
	}

	region := os.Getenv("FLY_REGION")
	if region == "" {
		region = "local"
	}

	results := make([]providers.BenchmarkResult, len(req.APIs))
	var wg sync.WaitGroup
	for i := range req.APIs {
		cfg := req.APIs[i]
		provider, ok := getProvider(cfg.Provider)
		if !ok {
			results[i] = providers.BenchmarkResult{
				APIConfigID: cfg.ID,
				APIName:     cfg.Name,
				Region:      region,
				Concurrency: 0,
				ErrorCount:  1,
			}
			results[i].Requests = []providers.RequestResult{{Index: 0, Error: "unknown provider: " + string(cfg.Provider)}}
			continue
		}
		wg.Add(1)
		go func(idx int, config providers.APIConfig) {
			defer wg.Done()
			results[idx] = RunConcurrent(provider, config, req.Text)
			results[idx].Region = region
		}(i, cfg)
	}
	wg.Wait()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(benchmarkResponse{Region: region, Results: results})
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, fly-prefer-region")
}

func getProvider(p providers.SupportedProvider) (providers.TTSProvider, bool) {
	switch p {
	case providers.ProviderElevenLabs:
		return providers.ElevenLabs{}, true
	case providers.ProviderOpenAI:
		return providers.OpenAI{}, true
	case providers.ProviderCartesia:
		return providers.Cartesia{}, true
	case providers.ProviderDeepgram:
		return providers.Deepgram{}, true
	case providers.ProviderMurf:
		return providers.Murf{}, true
	case providers.ProviderPolly:
		return providers.Polly{}, true
	case providers.ProviderRime:
		return providers.Rime{}, true
	case providers.ProviderSarvam:
		return providers.Sarvam{}, true
	case providers.ProviderGoogle:
		return providers.Google{}, true
	case providers.ProviderResemble:
		return providers.Resemble{}, true
	default:
		return nil, false
	}
}
