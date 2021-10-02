FROM golang:1.16-alpine as builder

RUN mkdir app
WORKDIR app/

COPY go.mod .
COPY go.sum .
COPY app/ ./app/

RUN CGO_ENABLED=0 GOOS=linux go build -o ./LogWatcher ./app/

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/app/LogWatcher .
COPY config.yaml .

EXPOSE 27100/udp

ENTRYPOINT ["./LogWatcher"]
