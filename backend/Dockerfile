FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates python3 py3-pip py3-numpy py3-pandas py3-pillow py3-openpyxl

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY cmd/ ./cmd/
COPY config/ ./config/
COPY migrations/ ./migrations/
COPY docs/ ./docs/
COPY scripts/ ./scripts/

RUN CGO_ENABLED=0 go build -o main ./cmd

FROM alpine:latest

RUN apk add --no-cache ca-certificates python3 py3-pip py3-numpy py3-pandas py3-pillow py3-openpyxl

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/pkg/email/templates ./pkg/email/templates
COPY scripts/ ./scripts/

EXPOSE 8080

CMD ["./main"] 