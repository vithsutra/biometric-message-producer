FROM golang:1.23-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o ./bin/main ./cmd/*

FROM alpine:latest

WORKDIR /app

COPY --from=build /app/bin/main .

ENTRYPOINT [ "./main" ]