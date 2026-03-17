package main

import (
	"os"
	"sort"
	"sync"

	"github.com/fly-apps/go-example/providers"
)

func getRegion() string {
	if r := os.Getenv("FLY_REGION"); r != "" {
		return r
	}
	return "local"
}

// RunConcurrent fires n simultaneous provider calls and aggregates results.
// Concurrency is clamped to 1–50. Stats are computed over successful requests only.
func RunConcurrent(provider providers.TTSProvider, config providers.APIConfig, text string) providers.BenchmarkResult {
	n := config.Concurrency
	if n < 1 {
		n = 1
	}
	if n > 50 {
		n = 50
	}

	requestResults := make([]providers.RequestResult, n)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					requestResults[idx] = providers.RequestResult{
						Index: idx,
						Error: "panic during request",
					}
				}
			}()
			result, err := provider.Call(config, text)
			if err != nil {
				requestResults[idx] = providers.RequestResult{Index: idx, Error: err.Error()}
			} else {
				requestResults[idx] = providers.RequestResult{
					Index:          idx,
					TTFB:          result.TTFB,
					TTLB:          result.TTLB,
					AudioSizeBytes: result.AudioSizeBytes,
				}
			}
		}(i)
	}
	wg.Wait()

	var ttfbValues, ttlbValues []float64
	successCount := 0
	for _, r := range requestResults {
		if r.Error != "" {
			continue
		}
		successCount++
		ttfbValues = append(ttfbValues, r.TTFB)
		ttlbValues = append(ttlbValues, r.TTLB)
	}

	return providers.BenchmarkResult{
		APIConfigID:  config.ID,
		APIName:      config.Name,
		Region:       getRegion(),
		Concurrency:  n,
		Requests:     requestResults,
		TTFBStats:    computeStats(ttfbValues),
		TTLBStats:    computeStats(ttlbValues),
		SuccessCount: successCount,
		ErrorCount:   n - successCount,
	}
}

// computeStats returns min, max, mean, p50, p95 using nearest-rank on the sorted slice.
// Returns zeroed Stats if values is empty.
func computeStats(values []float64) providers.Stats {
	if len(values) == 0 {
		return providers.Stats{}
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	min := sorted[0]
	max := sorted[len(sorted)-1]
	var sum float64
	for _, v := range sorted {
		sum += v
	}
	mean := sum / float64(len(sorted))

	p50Idx := int(float64(len(sorted)) * 0.50)
	if p50Idx >= len(sorted) {
		p50Idx = len(sorted) - 1
	}
	p95Idx := int(float64(len(sorted)) * 0.95)
	if p95Idx >= len(sorted) {
		p95Idx = len(sorted) - 1
	}

	return providers.Stats{
		Min:  min,
		Max:  max,
		Mean: mean,
		P50:  sorted[p50Idx],
		P95:  sorted[p95Idx],
	}
}
