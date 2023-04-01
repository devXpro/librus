FROM golang:latest as build

WORKDIR /src
COPY . .
RUN go build -ldflags "-s -w" -o /librus

FROM chromedp/headless-shell:latest
RUN apt-get update; apt install tini -y
ENTRYPOINT ["tini", "--"]
COPY --from=build /librus /librus
ENV TELEGRAM_TOKEN=token
ENV MONGO_HOST=mongodb
CMD ["/librus"]
