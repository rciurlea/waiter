FROM golang:1.12 AS builder
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_ENABLED=0
WORKDIR /waiter
COPY . .
RUN go build

FROM alpine
COPY --from=builder /waiter/waiter /bin
ENTRYPOINT ["/bin/waiter"]

