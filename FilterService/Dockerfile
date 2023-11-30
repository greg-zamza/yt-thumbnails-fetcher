FROM golang:1.21

WORKDIR /usr/src/app

COPY main.go ./

RUN go build -v -o /usr/local/bin/app ./main.go

CMD ["app"]
EXPOSE 8081
