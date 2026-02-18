FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /llmsa ./cmd/llmsa

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /llmsa /usr/local/bin/llmsa

ENTRYPOINT ["llmsa"]
