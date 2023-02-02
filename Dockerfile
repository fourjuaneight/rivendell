FROM golang:alpine

ENV PATH="/usr/local/go/bin:${PATH}"

RUN echo "http://dl-cdn.alpinelinux.org/alpine/edge/main" > /etc/apk/repositories \
  && echo "http://dl-cdn.alpinelinux.org/alpine/edge/community" >> /etc/apk/repositories \
  && echo "http://dl-cdn.alpinelinux.org/alpine/edge/testing" >> /etc/apk/repositories \
  && echo "http://dl-cdn.alpinelinux.org/alpine/v3.11/main" >> /etc/apk/repositories \
  && apk upgrade -U -a \
  && apk add --no-cache \
  autoconf \
  automake \
  bash \
  build-base \
  ca-certificates \
  curl \
  freetype \
  g++ \
  gcc \
  git \
  harfbuzz \
  libstdc++ \
  libtool \
  make \
  nasm \
  nss \
  openssh-client \
  pkgconfig \
  python \
  ttf-freefont \
  && rm -rf /var/cache/* \
  && mkdir /var/cache/apk

WORKDIR /app
COPY go.mod /app/
COPY go.sum /app/
RUN go mod download
RUN go mod tidy

EXPOSE 8090
