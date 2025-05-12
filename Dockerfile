# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:alpine AS builder
WORKDIR $GOPATH/src/app
ADD . ./
ENV GO111MODULE=on
WORKDIR $GOPATH/src/app/cmd/gmqttd
ARG TARGETPLATFORM
ARG BUILDPLATFORM
RUN go build

FROM alpine:latest
WORKDIR /gmqttd
# RUN apk update && apk add --no-cache tzdata
COPY --from=builder /go/src/app/cmd/gmqttd .
EXPOSE 1883 8883 8082 8083 8084
RUN chmod +x gmqttd
RUN pwd
RUN ls -lrt
ENTRYPOINT ["./gmqttd", "start", "-c", "/gmqttd/default_config.yml"]
