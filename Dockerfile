# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/shortlink .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
	&& adduser -D -H -u 10001 shortlink

WORKDIR /app

COPY --from=builder /out/shortlink /app/shortlink
COPY web /app/web
COPY etc/shorturl-api.example.yaml /app/etc/shorturl-api.example.yaml

USER shortlink

EXPOSE 8888

CMD ["/app/shortlink", "-f", "/app/etc/shorturl-api.example.yaml"]
