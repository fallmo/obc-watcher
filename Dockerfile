FROM golang:1.24 AS build-stage

WORKDIR /src

COPY go.mod go.sum ./

RUN go mod download

COPY ./cmd ./cmd

RUN CGO_ENABLED=0 GOOS=linux go build -o /dist/ -a -installsuffix cgo cmd/obc-watcher/main.go

FROM redhat/ubi9-minimal:9.6 AS serve-stage

WORKDIR /obc-watcher

COPY  --from=build-stage /dist/main /obc-watcher

CMD ["/obc-watcher/main"]