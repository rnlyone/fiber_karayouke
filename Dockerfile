FROM golang:1.24-alpine AS builder

EXPOSE 3000

RUN apk update \
  && apk add --no-cache \
    postgresql-client \
    build-base
RUN mkdir /app
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .

RUN go build -o main main.go

CMD ["./main"]