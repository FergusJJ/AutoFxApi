FROM golang:1.19.2-alpine as builder

WORKDIR /usr/src/fxapi

COPY . .
RUN go mod tidy