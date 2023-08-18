# syntax=docker/dockerfile:1
FROM golang:1.20.3-alpine as base

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# build image
FROM base as build

EXPOSE 3000

RUN go build -o go-debugger .

CMD ["/app/go-debugger"]

# debug image
FROM base as debug

EXPOSE 3000 4000

RUN CGO_ENABLED=0 go install -ldflags "-s -w -extldflags '-static'" github.com/go-delve/delve/cmd/dlv@latest
RUN CGO_ENABLED=0 go build -gcflags "all=-N -l" -o go-debugger .

CMD [ "/go/bin/dlv", "--listen=:4000", "--headless=true", "--log=true", "--accept-multiclient", "--api-version=2", "exec", "/app/go-debugger" ]
