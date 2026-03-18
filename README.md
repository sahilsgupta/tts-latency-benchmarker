# TTS Latency Benchmarker

A geo-distributed **text-to-speech (TTS) API latency** benchmarking tool. Deploy a single Go worker to multiple Fly.io regions, then run the frontend to fire benchmark requests at each region and compare TTFB (time to first byte) and TTLB (time to last byte) per provider.

Supports configurable **concurrent load** per API (e.g. 10 simultaneous requests to ElevenLabs, 5 to OpenAI). Results include per-request timings and aggregated stats (min, max, mean, p50, p95).

**Providers:** ElevenLabs, OpenAI, Cartesia, Deepgram, Murf (Falcon), AWS Polly.

---

## Run locally

1. **Install Go** (e.g. [go.dev/dl](https://go.dev/dl/) or `brew install go`).

2. **From the repo root:**
   ```bash
   go build -o tts-benchmarker .
   ./tts-benchmarker
   ```
   Or: `go run .`

3. **Open in a browser:**
   - Status: [http://localhost:8080](http://localhost:8080)
   - Benchmark UI: [http://localhost:8080/app](http://localhost:8080/app)

4. Add API configs (provider, API key, voice/model, concurrency), select regions, enter text, and click **Run benchmark**.

### Where to put your API key (e.g. Murf)

**Do not save the API key in the project or in any file.** The app is designed so keys are never stored on the server.

1. Run the app (e.g. `./tts-benchmarker` or `go run .`).
2. Open the benchmark UI: [http://localhost:8080/app](http://localhost:8080/app).
3. In the **API config** section, use **Add API** (or the existing card).
4. Set **Provider** to **Murf**, give a **Display name** (e.g. "Murf Falcon"), and paste your Murf API key into the **API Key** field.
5. Optionally set **Voice ID** (default is Matthew) and **Concurrency** (e.g. 1 to start).
6. Select at least one **Region** (for local runs the server reports region as "local").
7. Click **Run benchmark**.

The key is sent only in that HTTP request and used once for the benchmark; it is not written to disk or logged.

---

## Tests

No API keys or network calls to real TTS vendors are required for the default test suite.

```bash
go test ./... -count=1
```

- **`benchmark_test.go`** — `computeStats` (empty, single value, percentiles), `RunConcurrent` (success, all errors, concurrency clamp 1–50, panic recovery).
- **`main_test.go`** — HTTP handlers: `/health`, `/`, `/app`, `/benchmark` (OPTIONS, bad JSON, empty APIs, unknown provider, success with injected mock provider, `FLY_REGION` in response).
- **`providers/measure_test.go`** — `MeasureStream` against a local test server (streaming + HTTP error).

## Running E2E Tests

1. Copy `.env.example` to `.env`
2. Fill in your API keys (leave any provider blank to skip that test)
3. Run:

   ```bash
   make test-e2e
   ```

Each `TestE2E_*` calls the real provider API once with the phrase `"This is a latency test."` (equivalent to concurrency 1). Missing or empty env vars cause that test to be **skipped**, not failed. Requires network access.

---

## Deploy to Fly.io

1. **Install [flyctl](https://fly.io/docs/hub/installing/).**

2. **Launch and deploy:**
   ```bash
   fly launch   # first time: name app, pick region, etc.
   fly deploy   # build and deploy
   ```

3. **Scale to multiple regions** (optional):
   ```bash
   fly scale count 1 --region bom   # Mumbai
   fly scale count 1 --region fra   # Frankfurt
   fly scale count 1 --region sin   # Singapore
   fly scale count 1 --region iad   # Virginia
   fly scale count 1 --region lax   # Los Angeles
   fly scale count 1 --region syd   # Sydney
   ```

4. **Use the app:**
   - Status: `https://<your-app>.fly.dev/`
   - Benchmark UI: `https://<your-app>.fly.dev/app`

The frontend sends `POST /benchmark` with a `fly-prefer-region` header so each request hits the chosen region. API keys are sent in the request body only and are not stored on the server.

---

## Security CI (GitHub Actions)

On push/PR to `main` or `master`:

| Workflow | What it does |
|----------|----------------|
| [`.github/workflows/security.yml`](.github/workflows/security.yml) | **govulncheck**, **gosec**, **gitleaks**, **Trivy** (Dockerfile; fails on CRITICAL/HIGH misconfigs). |
| [`.github/workflows/codeql.yml`](.github/workflows/codeql.yml) | **CodeQL** for Go → **Security → Code scanning** (enable in repo settings; see [SECURITY.md](SECURITY.md)). |

**Dependabot** ([`.github/dependabot.yml`](.github/dependabot.yml)) opens weekly PRs for `go.mod`, GitHub Actions, and Docker. Turn on **Dependabot alerts** in repo settings — steps in [SECURITY.md](SECURITY.md).

---

## Project spec

See [SPEC.md](SPEC.md) for architecture, API contract, provider details, and behaviour.
