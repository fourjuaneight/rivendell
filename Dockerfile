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
    # single-file hardcodes "--single-process", which crashes chromium >= ~131 so the
    # remote-debugging port never opens and captures fail silently. Strip that flag.
    # grep -q first so the build fails loudly if upstream renames/removes the line.
    && BROWSER_JS="$(npm root -g)/single-file-cli/lib/browser.js" \
    && grep -q 'args.push("--single-process");' "$BROWSER_JS" \
    && sed -i '/args.push("--single-process");/d' "$BROWSER_JS" \
    && npm cache clean --force

WORKDIR /app
COPY --from=builder /app/rivendell .

VOLUME /app/pb_data

EXPOSE 8090
CMD ["./rivendell", "serve", "--http=0.0.0.0:8090", "--dir=/app/pb_data"]
