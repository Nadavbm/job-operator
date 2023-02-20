# Build the manager binary
FROM golang:1.19 as builder

# Copy the go source
COPY . /jobop
WORKDIR /jobop

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o jobop main.go

FROM alpine:latest
WORKDIR /
COPY --from=builder /jobop/jobop /jobop

CMD /jobop
