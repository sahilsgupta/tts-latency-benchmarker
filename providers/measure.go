package providers

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

// requestURLBlocksSSRF rejects loopback, link-local, and private targets to limit SSRF
// when benchmark configs could otherwise point at cloud metadata or internal services.
func requestURLBlocksSSRF(u *url.URL) error {
	if u == nil || u.Host == "" {
		return fmt.Errorf("missing URL host")
	}
	switch strings.ToLower(u.Scheme) {
	case "https", "http":
	default:
		return fmt.Errorf("unsupported URL scheme %q", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("missing hostname")
	}
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return fmt.Errorf("blocked address")
		}
		return nil
	}
	h := strings.ToLower(host)
	if h == "localhost" || strings.HasSuffix(h, ".localhost") {
		return fmt.Errorf("blocked host")
	}
	// Common cloud metadata hostnames
	if h == "metadata.google.internal" || strings.HasSuffix(h, ".internal") {
		return fmt.Errorf("blocked host")
	}
	return nil
}

// MeasureStream times a single HTTP request using streaming response body reads.
// It returns TTFB (time to first byte) and TTLB (time to last byte) in milliseconds.
// The client should use a timeout (e.g. 30s); callers are responsible for setting it on the request.
func MeasureStream(req *http.Request) (TTSStreamResult, error) {
	// Unit tests use httptest (127.0.0.1); production enforces SSRF limits below.
	if !testing.Testing() {
		if err := requestURLBlocksSSRF(req.URL); err != nil {
			return TTSStreamResult{}, fmt.Errorf("measure: %w", err)
		}
	}
	start := time.Now()

	client := &http.Client{Timeout: 30 * time.Second}
	// #nosec G704 -- non-test callers validate req.URL in requestURLBlocksSSRF (SSRF guard).
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
