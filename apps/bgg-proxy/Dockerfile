FROM golang:1.25-bookworm AS instrumentation-builder

RUN apt-get update && apt-get install -y git make gcc llvm clang
RUN git clone https://github.com/open-telemetry/opentelemetry-go-instrumentation.git
RUN cd opentelemetry-go-instrumentation/ && \
    make build

FROM golang:1.25-bookworm AS builder
WORKDIR /app

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o bgg-proxy ./cmd/main.go

FROM alpine:latest AS production
WORKDIR /app
COPY --from=instrumentation-builder \
    /opentelemetry-go-instrumentation/otel-go-instrumentation \
    /app/otel-go-instrumentation
COPY --from=builder \
    /app/bgg-proxy \
    /app/bgg-proxy

EXPOSE 8080
ENTRYPOINT ["./app/otel-go-instrumentation", "-target-exe", "/app/bgg-proxy"]