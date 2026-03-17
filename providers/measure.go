package providers

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// MeasureStream times a single HTTP request using streaming response body reads.
// It returns TTFB (time to first byte) and TTLB (time to last byte) in milliseconds.
// The client should use a timeout (e.g. 30s); callers are responsible for setting it on the request.
func MeasureStream(req *http.Request) (TTSStreamResult, error) {
	start := time.Now()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return TTSStreamResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		_, _ = io.Copy(io.Discard, resp.Body)
		return TTSStreamResult{}, &HTTPError{StatusCode: resp.StatusCode}
	}

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
		if err == io.EOF {
			break
		}
		if err != nil {
			return TTSStreamResult{}, err
		}
	}

	ttlb := float64(time.Since(start).Milliseconds())
	return TTSStreamResult{TTFB: ttfb, TTLB: ttlb, AudioSizeBytes: totalBytes}, nil
}

// HTTPError represents a non-2xx response for error reporting.
type HTTPError struct {
	StatusCode int
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, http.StatusText(e.StatusCode))
}
