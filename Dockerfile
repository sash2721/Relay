# Stage 1: Build
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o relay .

# Stage 2: Runtime
FROM alpine:latest

RUN apk add --no-cache git docker-cli ca-certificates

WORKDIR /app

COPY --from=builder /build/relay .

EXPOSE 3000 8080

ENTRYPOINT ["./relay"]
