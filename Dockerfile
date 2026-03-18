# Build stage
FROM golang:1.22-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /tts-benchmarker .

# Runtime stage (non-root for container hardening / Trivy DS002)
FROM alpine:3.23
RUN apk --no-cache add ca-certificates \
    && adduser -D -u 1000 appuser
WORKDIR /app
COPY --from=build --chown=appuser:appuser /tts-benchmarker .
USER appuser

EXPOSE 8080
ENV PORT=8080
CMD ["./tts-benchmarker"]
