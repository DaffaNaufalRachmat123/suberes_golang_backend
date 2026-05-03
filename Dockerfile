# STAGE 1: Build tahap kompilasi
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod dan go.sum terlebih dahulu untuk caching module
COPY go.mod go.sum ./
RUN go mod download

# Copy seluruh source code
COPY . .

# Build aplikasi menjadi executable bernama "suberes_app"
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o suberes_app main.go

# STAGE 2: Minimal runtime image
FROM alpine:latest

# Wajib install tzdata (untuk time.LoadLocation) dan ca-certificates (untuk request HTTPS / FCM)
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

COPY --from=builder /app/suberes_app .

EXPOSE 8080

CMD ["./suberes_app"]