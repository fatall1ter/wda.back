############################
# STEP 0 get dependencies
############################
FROM golang:1.16.5 AS dependencies
ENV GOPRIVATE=*.countmax.ru
RUN apt-get update && apt-get install openssl -y
ENV DOMAIN_NAME=git.countmax.ru \
  TCP_PORT=443
RUN openssl s_client -connect $DOMAIN_NAME:$TCP_PORT -showcerts </dev/null 2>/dev/null | openssl x509 -outform PEM | tee /usr/local/share/ca-certificates/$DOMAIN_NAME.crt
RUN update-ca-certificates
WORKDIR /go/src
COPY go.mod .
COPY go.sum .
RUN go mod download
############################
# STEP 1 build executable binary
############################
FROM dependencies AS builder
ARG BUILD_NUMBER
ARG GIT_HASH
ENV BUILD_NUMBER ${BUILD_NUMBER}
ENV GIT_HASH ${GIT_HASH}
ENV TZ=Europe/Moscow
ENV GOPRIVATE=*.countmax.ru
ENV GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /go/src
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
COPY . .
RUN update-ca-certificates
RUN DOCKER_TLS_VERIFY=0

RUN go build -ldflags="-X 'main.build=$BUILD_NUMBER' -X 'main.githash=$GIT_HASH'" -o /go/bin/wda.back

############################
# STEP 2 build frontend dist folder
############################
FROM node:14.2.0-alpine3.11 as front
ARG A_SERVER
ENV AUTH_SERVER=${A_SERVER}
RUN env
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
# ADD ./ssl/git.countmax.ru.crt /usr/local/share/ca-certificates/
# RUN update-ca-certificates
RUN apk add --no-cache git
RUN git config --global http.sslVerify false
RUN git clone https://git.countmax.ru/countmax/wda.front.git /web
WORKDIR /web
RUN yarn global add @quasar/cli && \
  yarn install && \
  quasar build
############################
# STEP 3 build a small image
############################
FROM alpine
LABEL maintainer="it@watcom.ru" version="0.0.3"
RUN apk add --no-cache tzdata wget
ENV HTTP_PORT=8000
ENV TZ=Europe/Moscow
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
COPY --from=builder /go/bin/wda.back /go/bin/wda.back
COPY --from=builder /go/src/config.yaml /go/bin/config.yaml
COPY --from=front /web/dist /go/bin/web

WORKDIR /go/bin/
EXPOSE 8000 8001
CMD ["./wda.back"]
