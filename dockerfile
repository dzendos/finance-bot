FROM golang:latest AS builder
WORKDIR /build
COPY . .
RUN go build -o /build/app gitlab.ozon.dev/e.gerasimov/telegram-bot/cmd/bot
