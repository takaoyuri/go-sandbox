FROM golang:latest AS build
WORKDIR /go/src/github.com/takaoyuri/go-sandbox/golangapi
ENV GO111MODULE=on
# make cache
COPY go.mod .
COPY go.sum .
RUN go mod download
# COPY the source code as the last step
COPY . .
# build app for next stage
RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM alpine:latest
WORKDIR /app
COPY --from=build /go/bin/app .
COPY ken-all.json .
# add user
RUN addgroup go && \
    adduser -D -G go go && \
    chown -R go:go /app/app
CMD [ "./app" ]
