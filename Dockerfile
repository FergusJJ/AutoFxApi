FROM golang:1.19.2-alpine as builder

WORKDIR /usr/src/fxapi

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /usr/src/fxapi/cmd/monitor.exe /usr/src/fxapi/cmd/monitor/main.go 