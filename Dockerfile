FROM golang:1.15-alpine AS build

RUN apk add --no-cache git ca-certificates
WORKDIR /nas-bot
ADD . / /nas-bot/
RUN CGO_ENABLED=0 go build -o /bin/nas-bot ./cmd

FROM alpine:latest
MAINTAINER ihciah <ihciah@gmail.com>

RUN apk add --no-cache ca-certificates ipmitool
COPY --from=build /bin/nas-bot /bin/nas-bot
CMD exec nas-bot