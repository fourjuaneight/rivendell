FROM golang:alpine as builder

# Config
ENV PATH="/usr/local/go/bin:${PATH}"

# Install dependencies
RUN apk update && apk upgrade && \
  apk add --no-cache bash git openssh \
  && rm -rf /var/cache/* \
  && mkdir /var/cache/apk

# Download App Dependencies
WORKDIR /app
COPY . /app
RUN go mod download
RUN go mod tidy

EXPOSE 8090
