FROM golang:1.21

WORKDIR /usr/src/app

COPY go.mod go.sum main.go ./
RUN go mod download && go mod verify

RUN go build -v -o /usr/local/bin/app ./...

CMD ["app"]
