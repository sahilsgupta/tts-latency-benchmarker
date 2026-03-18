package main

import (
	"errors"
	"testing"

	"tts-benchmarker/providers"
)

func TestComputeStats_empty(t *testing.T) {
	s := computeStats(nil)
	if s.Min != 0 || s.Max != 0 || s.Mean != 0 || s.P50 != 0 || s.P95 != 0 {
		t.Fatalf("expected zero stats, got %+v", s)
	}
}

func TestComputeStats_single(t *testing.T) {
	s := computeStats([]float64{42})
	if s.Min != 42 || s.Max != 42 || s.Mean != 42 || s.P50 != 42 || s.P95 != 42 {
		t.Fatalf("single value: %+v", s)
	}
}

func TestComputeStats_sorted(t *testing.T) {
	s := computeStats([]float64{10, 20, 30, 40, 50})
	if s.Min != 10 || s.Max != 50 {
		t.Fatalf("min/max: %+v", s)
	}
	if s.Mean != 30 {
		t.Fatalf("mean want 30, got %v", s.Mean)
	}
	// len 5: p50 idx 2 -> 30, p95 idx 4 -> 50
	if s.P50 != 30 || s.P95 != 50 {
		t.Fatalf("p50/p95 want 30/50, got %v/%v", s.P50, s.P95)
	}
}

func TestComputeStats_unsortedInput(t *testing.T) {
	s := computeStats([]float64{50, 10, 30})
	if s.Min != 10 || s.Max != 50 || s.Mean != 30 {
		t.Fatalf("unsorted: %+v", s)
	}
}

type stubProvider struct {
	res providers.TTSStreamResult
	err error
}

func (s stubProvider) Call(_ providers.APIConfig, _ string) (providers.TTSStreamResult, error) {
	return s.res, s.err
}

func TestRunConcurrent_success(t *testing.T) {
	p := stubProvider{res: providers.TTSStreamResult{TTFB: 1, TTLB: 2, AudioSizeBytes: 100}}
	cfg := providers.APIConfig{ID: "a", Name: "Test", Concurrency: 3}
	out := RunConcurrent(p, cfg, "hello")
	if out.Concurrency != 3 || out.SuccessCount != 3 || out.ErrorCount != 0 {
		t.Fatalf("concurrency/success: %+v", out)
	}
	if len(out.Requests) != 3 {
		t.Fatalf("requests len %d", len(out.Requests))
	}
	for i, r := range out.Requests {
		if r.Index != i || r.Error != "" || r.TTFB != 1 || r.TTLB != 2 {
			t.Fatalf("request %d: %+v", i, r)
		}
	}
	if out.TTFBStats.Min != 1 || out.TTFBStats.Max != 1 {
		t.Fatalf("ttfb stats: %+v", out.TTFBStats)
	}
}

func TestRunConcurrent_errors(t *testing.T) {
	p := stubProvider{err: errors.New("boom")}
	cfg := providers.APIConfig{Concurrency: 2}
	out := RunConcurrent(p, cfg, "x")
	if out.SuccessCount != 0 || out.ErrorCount != 2 {
		t.Fatalf("all errors: success=%d err=%d", out.SuccessCount, out.ErrorCount)
	}
	if out.TTFBStats.Min != 0 && out.TTFBStats.Max != 0 {
		t.Fatalf("stats should be zero on all failures: %+v", out.TTFBStats)
	}
}

func TestRunConcurrent_concurrencyClamp(t *testing.T) {
	p := stubProvider{res: providers.TTSStreamResult{TTFB: 1, TTLB: 1, AudioSizeBytes: 1}}
	out0 := RunConcurrent(p, providers.APIConfig{Concurrency: 0}, "x")
	if out0.Concurrency != 1 || len(out0.Requests) != 1 {
		t.Fatalf("clamp 0->1: %v", out0.Concurrency)
	}
	outHigh := RunConcurrent(p, providers.APIConfig{Concurrency: 100}, "x")
	if outHigh.Concurrency != 50 || len(outHigh.Requests) != 50 {
		t.Fatalf("clamp 100->50: %v", outHigh.Concurrency)
	}
}

type panicProvider struct{}

func (panicProvider) Call(_ providers.APIConfig, _ string) (providers.TTSStreamResult, error) {
	panic("intentional")
}

func TestRunConcurrent_panicRecovered(t *testing.T) {
	out := RunConcurrent(panicProvider{}, providers.APIConfig{Concurrency: 1}, "x")
	if out.SuccessCount != 0 || out.ErrorCount != 1 {
		t.Fatalf("panic should count as error: %+v", out)
	}
	if out.Requests[0].Error != "panic during request" {
		t.Fatalf("want panic message, got %q", out.Requests[0].Error)
	}
}
