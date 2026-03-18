package providers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestMeasureStream_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fl, ok := w.(http.Flusher)
		if ok {
			w.Write([]byte("chunk1"))
			fl.Flush()
			time.Sleep(5 * time.Millisecond)
			w.Write([]byte("chunk2"))
			return
		}
		w.Write([]byte("all"))
	}))
	defer srv.Close()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	res, err := MeasureStream(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.AudioSizeBytes < 5 {
		t.Fatalf("bytes %d", res.AudioSizeBytes)
	}
	if res.TTFB < 0 || res.TTLB < res.TTFB {
		t.Fatalf("ttfb %v ttlb %v", res.TTFB, res.TTLB)
	}
}

func TestMeasureStream_httpError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusBadRequest)
	}))
	defer srv.Close()

	req, _ := http.NewRequest(http.MethodGet, srv.URL, nil)
	_, err := MeasureStream(req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Fatalf("want 400 in error, got %v", err)
	}
}

func TestHTTPError_Error(t *testing.T) {
	e := &HTTPError{StatusCode: 404}
	if !strings.Contains(e.Error(), "404") {
		t.Fatal(e.Error())
	}
}

func TestRequestURLBlocksSSRF(t *testing.T) {
	tests := []struct {
		raw     string
		wantErr bool
	}{
		{"https://api.elevenlabs.io/v1/foo", false},
		{"http://169.254.169.254/latest/meta-data/", true},
		{"http://127.0.0.1:8080/", true},
		{"https://10.0.0.1/", true},
		{"ftp://example.com/", true},
	}
	for _, tt := range tests {
		u, err := url.Parse(tt.raw)
		if err != nil {
			t.Fatalf("parse %q: %v", tt.raw, err)
		}
		got := requestURLBlocksSSRF(u)
		if (got != nil) != tt.wantErr {
			t.Errorf("%q: err=%v wantErr=%v", tt.raw, got, tt.wantErr)
		}
	}
}
