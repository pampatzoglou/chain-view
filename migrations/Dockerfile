# Specify the Go version
ARG GO_VERSION=1.23.2

FROM golang:${GO_VERSION}-alpine AS builder

# Install git and build tools in a single step to reduce layers
RUN apk add --no-cache git

# Install the golang-migrate tool
RUN go install github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

# If you need a smaller final image, for instance using it in runtime,
# you can use a second stage to copy only the necessary binaries:
FROM alpine AS development

# Copy the migrate binary from the builder stage
COPY --from=builder /go/bin/migrate /usr/local/bin/

# Ensure it's executable
RUN chmod +x /usr/local/bin/migrate

# Copy migration files
COPY --chown=nobody:nogroup migrations /migrations

# Set default entrypoint for running the migrate command
USER nobody
ENTRYPOINT ["/usr/local/bin/migrate"]

# Optional: if you want to default to the version output or help
CMD ["--help"]

# Production stage using Distroless for a minimal and secure image
FROM gcr.io/distroless/cc-debian12@sha256:899570acf85a1f1362862a9ea4d9e7b1827cb5c62043ba5b170b21de89618608 AS production

# Copy the migrate binary from the development stage
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate

# Copy the migration files
COPY --from=development /migrations /migrations

# Set the working directory for the production stage (optional)
WORKDIR /migrations

# Use a non-root user for security in production
USER nobody

# Set default entrypoint for running the migrate command
ENTRYPOINT ["/usr/local/bin/migrate"]

# Optional: default to the --help command if no arguments are provided
CMD ["--help"]