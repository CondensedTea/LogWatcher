FROM golang:1.16-alpine as builder


RUN mkdir app
WORKDIR app/

COPY go.mod .
COPY go.sum .
COPY server/ ./server/

RUN CGO_ENABLED=0 GOOS=linux go build -o ./LogWatcher ./server/

FROM scratch

COPY --from=builder /go/app/LogWatcher .
COPY config.yaml .

EXPOSE 27100/udp

ENTRYPOINT ["./LogWatcher"]
CMD [""]