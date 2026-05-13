FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o rivendell .

FROM alpine:latest

RUN apk add --no-cache ca-certificates ffmpeg tzdata wget \
    && wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp \
         -O /usr/local/bin/yt-dlp \
    && chmod a+rx /usr/local/bin/yt-dlp

WORKDIR /app
COPY --from=builder /app/rivendell .

VOLUME /app/pb_data

EXPOSE 8090
CMD ["./rivendell", "serve", "--http=0.0.0.0:8090", "--dir=/app/pb_data"]
