# Build stage
FROM golang:1.24.4-alpine AS builder

WORKDIR /app
COPY . .

# Install dependencies
RUN apk add --no-cache ffmpeg make git
RUN go install github.com/a-h/templ/cmd/templ@v0.2.543

# Build
RUN make build

# Final image
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app .
RUN apk add --no-cache ffmpeg
EXPOSE 8080
CMD ["./app"]
