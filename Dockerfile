# =============================================================================
# Stage 1: Build
# =============================================================================
FROM golang:1.25-alpine AS builder

# Install git (needed by some go modules) and build essentials
RUN apk --no-cache add ca-certificates tzdata git

WORKDIR /app

# Dependency layer — cached separately so rebuilds after code-only changes are fast
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build a statically-linked, stripped binary
# -trimpath: removes local file paths from stack traces (security)
# -ldflags "-s -w": strip debug info for smaller image
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /app/suberes_app \
    ./main.go

# =============================================================================
# Stage 2: Minimal runtime image
# =============================================================================
FROM alpine:3.21

# Security: keep OS packages up to date in the image layer
RUN apk --no-cache upgrade && \
    apk --no-cache add ca-certificates tzdata && \
    # Create a non-root user/group for running the app
    addgroup -S appgroup && \
    adduser  -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/suberes_app .

# Copy static assets the app serves at runtime
COPY --from=builder /app/images ./images

# Fix ownership so appuser can read the binary and assets
RUN chown -R appuser:appgroup /app

# Drop privileges — never run as root in production
USER appuser

EXPOSE 8080

# Docker will restart the container if the health check fails repeatedly
HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["./suberes_app"]