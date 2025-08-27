FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates python3 py3-pip && \
    pip3 install --no-cache-dir numpy pandas pillow openpyxl

WORKDIR /app
    
COPY go.mod go.sum ./
RUN go mod download

COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY cmd/ ./cmd/
COPY config/ ./config/
COPY migrations/ ./migrations/

COPY scripts/ ./scripts/

RUN CGO_ENABLED=0 go build -o main ./cmd

FROM alpine:latest

RUN apk add --no-cache ca-certificates python3 py3-pip && \
    pip3 install --no-cache-dir numpy pandas pillow openpyxl

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/scripts ./scripts
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./main"] 