FROM golang:1.26-alpine AS builder
WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/control-api ./cmd/control-api

FROM alpine:3.21
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /out/control-api /usr/local/bin/control-api

EXPOSE 9091

USER app
ENTRYPOINT ["control-api"]
