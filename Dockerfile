# Build stage
FROM golang:1.22-alpine AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /tts-benchmarker .

# Runtime stage
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=build /tts-benchmarker .

EXPOSE 8080
ENV PORT=8080
CMD ["./tts-benchmarker"]
