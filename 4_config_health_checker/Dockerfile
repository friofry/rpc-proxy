# Build stage
FROM golang:1.21-alpine as builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /health-checker

# Runtime stage  
FROM alpine:3.18

WORKDIR /app
COPY --from=builder /health-checker /health-checker
COPY checker_config.json /app/checker_config.json
COPY test_methods.json /app/test_methods.json

EXPOSE 8080

CMD ["/health-checker", "--checker-config", "/app/checker_config.json", "--default-providers", "$DEFAULT_PROVIDERS_PATH", "--reference-providers", "$REFERENCE_PROVIDERS_PATH"]
