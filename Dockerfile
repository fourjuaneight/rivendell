FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o rivendell .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates ffmpeg tzdata wget \
    && wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp \
         -O /usr/local/bin/yt-dlp \
    && chmod a+rx /usr/local/bin/yt-dlp

WORKDIR /app
COPY --from=builder /app/rivendell .

EXPOSE 8090
CMD ["./rivendell", "serve", "--http=0.0.0.0:8090"]
