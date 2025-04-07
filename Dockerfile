FROM golang:1.24-alpine

WORKDIR /src

COPY . .

RUN go build -ldflags "-s -w" -o /librus

ENV TELEGRAM_TOKEN=token
ENV MONGO_HOST=mongodb

CMD "/librus"
