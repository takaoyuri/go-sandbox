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
RUN CGO_ENABLED=1 go build -o /go/bin/app

# FROM alpine:latest
# RUN apk add --no-cache libc6-compat sqlite  # 互換ライブラリを追加
# WORKDIR /app
# COPY --from=build /go/bin/app .
# COPY KEN_ALL.CSV .
# # add user
# RUN addgroup go && \
#     adduser -D -G go go && \
#     chown -R go:go /app
# USER go
# CMD [ "./app" ]

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /go/bin/app .
COPY KEN_ALL.CSV .
RUN groupadd go && useradd -g go go && chown -R go:go /app
USER go
CMD [ "./app" ]

