FROM golang:1.27-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    && update-ca-certificates

ENV GOPATH=/go
ENV PATH=$GOPATH/bin:/usr/local/go/bin:$PATH

WORKDIR /app

# Copy source code
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o git-workspace ./cmd/web

# Build git-askpass from Go source
COPY cmd/git-askpass/main.go /tmp/git-askpass.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /tmp/git-askpass /tmp/git-askpass.go

FROM alpine:3.20

# Install runtime dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    openssh-client \
    tzdata \
    wget \
    && update-ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S appuser && \
    adduser -S -u 1001 -G appuser appuser

# Create directories
RUN mkdir -p /app /data

# Copy built binaries from builder stage
COPY --from=builder --chown=appuser:appuser /app/git-workspace /app/git-workspace
COPY --from=builder --chown=appuser:appuser /app/static /app/static
COPY --from=builder --chown=appuser:appuser /tmp/git-askpass /usr/local/bin/git-askpass

# Set permissions
RUN chmod +x /usr/local/bin/git-askpass

# Set environment variables
ENV GIT_ASKPASS=/usr/local/bin/git-askpass
ENV PORT=8080
ENV WORKSPACE_BASE_DIR=/data
ENV STATIC_DIR=/app/static

# Switch to non-root user
USER appuser
WORKDIR /app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["/app/git-workspace"]
