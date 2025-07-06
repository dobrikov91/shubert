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
RUN go mod download
COPY . .

RUN go build .

# final stage
FROM debian:bookworm
RUN apt-get update && apt-get install -y \
    alsa-utils libportmidi0 curl

WORKDIR /app
COPY --from=builder /app/shubert .
COPY ./templates ./templates

CMD ["./shubert"]
