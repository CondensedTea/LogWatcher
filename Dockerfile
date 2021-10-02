FROM golang:1.16-alpine as builder

RUN mkdir Logwatcher
WORKDIR Logwatcher/

COPY go.mod .
COPY go.sum .
COPY app/ ./app
COPY pkg/ ./pkg

RUN CGO_ENABLED=0 GOOS=linux go build -o ./service ./app

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/Logwatcher/service .
COPY config.yaml .

EXPOSE 27100/udp

ENTRYPOINT ["./LogWatcher"]
