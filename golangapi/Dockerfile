FROM golang:latest

WORKDIR /go/src/godocker

COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

CMD ["godocker"]
