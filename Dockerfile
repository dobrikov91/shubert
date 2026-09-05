# syntax=docker/dockerfile:1

# build stage
FROM golang:bookworm AS builder
RUN apt-get update && apt-get install -y \
    alsa-utils \
    libasound2 \
    libasound2-dev \
    libasound2-plugins \
    libportmidi-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build .

# final stage
FROM debian:bookworm
RUN apt-get update && apt-get install -y \
    alsa-utils libportmidi0 curl

WORKDIR /app
COPY --from=builder /app/shubert .

CMD ["./shubert"]
