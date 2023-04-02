FROM golang:latest

WORKDIR /src

COPY . .

RUN go build -ldflags "-s -w" -o /librus

ENV TELEGRAM_TOKEN=token
ENV MONGO_HOST=mongodb

CMD "/librus"
