FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates python3 py3-numpy py3-pandas py3-pillow py3-openpyxl

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Copy scripts
RUN mkdir -p /app/scripts && \
    cp -r scripts/* /app/scripts/ || true

# Build the application
RUN CGO_ENABLED=0 go build -o main ./cmd

# Final stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates python3 py3-numpy py3-pandas py3-pillow py3-openpyxl

WORKDIR /app

# Copy binary, scripts and docs from builder
COPY --from=builder /app/main .
COPY --from=builder /app/scripts ./scripts
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./main"] 