FROM alpine:latest

WORKDIR /app

COPY app/app .

# For audio conversion
RUN apk add --no-cache ffmpeg

EXPOSE 8080

CMD ["./app"]
