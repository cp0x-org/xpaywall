FROM golang:1.26-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/example-server ./cmd

FROM alpine:3.21
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app
COPY --from=builder /out/example-server /usr/local/bin/example-server

EXPOSE 4021

USER app
ENTRYPOINT ["example-server"]
