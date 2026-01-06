# Multi-stage Docker build for minimal final image size
# Stage 1: Build the Go binary
# Stage 2: Create minimal runtime image with just the binary

# ==============================================================================
# Stage 1: Builder
# ==============================================================================
# Use official Go image based on Alpine Linux
# Alpine is chosen for small size (~300 MB with Go vs ~800 MB with Debian)
# golang:1.24-alpine includes Go compiler and build tools
FROM golang:1.24-alpine AS builder

# Install git (needed for go mod download with private repos, though we don't have any)
# ca-certificates needed for downloading dependencies over HTTPS
RUN apk add --no-cache git ca-certificates

# Set working directory inside the container
# All subsequent commands will run from this directory
WORKDIR /app

# Copy go.mod and go.sum first (before copying source code)
# This leverages Docker layer caching:
# - If go.mod/go.sum haven't changed, Docker reuses cached layer with downloaded deps
# - Only re-downloads dependencies when go.mod/go.sum change
# - This speeds up builds significantly
COPY go.mod go.sum ./

# Download dependencies
# go mod download fetches all modules listed in go.mod
# Dependencies are cached in Docker layer
RUN go mod download

# Verify dependencies (security check)
# go mod verify checks that dependencies match go.sum checksums
# Ensures no tampering with downloaded modules
RUN go mod verify

# Copy the entire source code
# This happens AFTER dependency download to maximize cache reuse
# If source changes but dependencies don't, we reuse the dependency layer
COPY . .

# Build the Go binary
# CGO_ENABLED=0: Build a static binary without C dependencies
#   - Why? Alpine uses musl libc, not glibc. Static binary avoids compatibility issues
#   - Static binary can run on any Linux, including minimal base images
# GOOS=linux: Target operating system (Linux)
# GOARCH=amd64: Target architecture (64-bit x86, Cloud Run uses this)
# -ldflags="-w -s": Linker flags to reduce binary size
#   - -w: Omit DWARF debugging information
#   - -s: Omit symbol table and debug info
#   - Reduces binary size by ~30% with no runtime impact
# -a: Force rebuild of all packages (ensures clean build)
# -installsuffix cgo: Use different install suffix for cgo vs non-cgo builds
# -o /app/run-tbot: Output binary to /app/run-tbot (not /app/bot to avoid conflict with bot/ directory)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -a \
    -installsuffix cgo \
    -o /app/run-tbot \
    .

# ==============================================================================
# Stage 2: Runtime
# ==============================================================================
# Use minimal Alpine Linux image for runtime
# alpine:latest is ~7 MB (vs ubuntu:latest ~77 MB)
# We don't need Go compiler or build tools in production, just the binary
FROM alpine:latest

# Install ca-certificates for HTTPS requests to Telegram API
# Without this, bot can't verify Telegram's SSL certificate
# As we discussed, this is needed for OUTGOING HTTPS requests
# Size: ~150 KB
RUN apk --no-cache add ca-certificates

# Create non-root user for security
# Running as root is a security risk - if bot is compromised, attacker has root access
# Creating dedicated user limits damage from potential exploits
# -D: Don't assign a password (can't login as this user)
# -g: GECOS field (user description)
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /home/appuser

# Copy the binary from builder stage
# --from=builder: Copy from the builder stage, not the current stage
# --chown=appuser:appuser: Set ownership to appuser (not root)
# This ensures the binary is owned by non-root user
# Using explicit paths: source /app/run-tbot → destination /home/appuser/run-tbot
COPY --from=builder --chown=appuser:appuser /app/run-tbot /home/appuser/run-tbot

# Switch to non-root user
# All subsequent commands and the container itself will run as this user
USER appuser

# Expose port 8080
# This is documentary - doesn't actually open the port
# Cloud Run will route traffic to this port (set via PORT env var)
# Most bots use 8080, but Cloud Run sets PORT automatically
EXPOSE 8080

# Health check (optional, but good practice)
# Docker/Cloud Run can use this to check if container is healthy
# Every 30 seconds, curl the health endpoint
# If it fails 3 times in a row, container is marked unhealthy
# Cloud Run doesn't use this (uses its own probes), but good for local testing
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Command to run when container starts
# No need for shell (we're running a single binary)
# Using JSON array format (exec form) is more efficient than shell form
# This starts the bot directly without wrapping in a shell
# Using absolute path to the binary
CMD ["/home/appuser/run-tbot"]
