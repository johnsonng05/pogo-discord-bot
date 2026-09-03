# Stage 1: Compilation Environment
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o pogo-discord-bot .

# Stage 2: Final Minimal Run Iron
FROM alpine:3.22.5
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/pogo-discord-bot .
CMD ["./pogo-discord-bot"]
