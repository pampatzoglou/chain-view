# Specify the Go version
ARG GO_VERSION=1.23.2
FROM golang:${GO_VERSION}-alpine AS development

WORKDIR /app
COPY . .

# Download dependencies
RUN go mod download

# Command to run the application in development mode
CMD ["go", "run", "app/main.go"]

FROM development AS build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o chain-view app/main.go

# FROM gcr.io/distroless/cc-debian12@sha256:899570acf85a1f1362862a9ea4d9e7b1827cb5c62043ba5b170b21de89618608 AS production
FROM golang:${GO_VERSION}-alpine AS production

WORKDIR /app
COPY --from=build --chown=nobody:nogroup /app/chain-view /bin/chain-view
COPY --from=development --chown=nobody:nogroup /app/config/config.yaml config/

# Use a non-root user for security in production
USER nobody

CMD ["/bin/chain-view"]