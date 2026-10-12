FROM golang:1.26.2-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o rivendell .

FROM alpine:latest

RUN apk add --no-cache ca-certificates chromium ffmpeg nodejs npm tzdata wget \
    && wget https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp \
         -O /usr/local/bin/yt-dlp \
    && chmod a+rx /usr/local/bin/yt-dlp \
    && npm install -g single-file-cli \
    && npm cache clean --force

WORKDIR /app
COPY --from=builder /app/rivendell .

VOLUME /app/pb_data

EXPOSE 8090
# --encryptionEnv names the env var holding the AES-256 key (32 chars) used to
# encrypt the settings row in _params at rest, so B2/S3 credentials and OAuth2
# secrets aren't stored in the clear inside data.db. Losing PB_ENCRYPTION_KEY
# makes the stored settings unreadable — back it up with the .env file.
CMD ["./rivendell", "serve", "--http=0.0.0.0:8090", "--dir=/app/pb_data", "--encryptionEnv=PB_ENCRYPTION_KEY"]
