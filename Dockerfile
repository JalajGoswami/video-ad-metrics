FROM golang:1.23.5-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download
COPY . .

RUN go build -o /app/server ./cmd/server

# Use a small alpine image for the final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from the first (builder) stage
COPY --from=builder /app/server .

ENV PORT=5000
EXPOSE 5000

ENTRYPOINT ["/app/server"]