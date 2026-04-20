# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder

WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# Stage 2: Build backend
FROM golang:1.25-alpine AS backend-builder

RUN apk add --no-cache git

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o relay .

# Stage 3: Runtime
FROM alpine:latest

RUN apk add --no-cache git docker-cli ca-certificates

WORKDIR /app

COPY --from=backend-builder /build/relay .
COPY --from=frontend-builder /frontend/dist ./frontend/dist

EXPOSE 3000 8080

ENTRYPOINT ["./relay"]
