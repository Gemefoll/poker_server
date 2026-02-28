FROM golang:latest AS builder

COPY go.mod /app/
COPY go.sum /app/
WORKDIR /app
RUN go mod download

COPY ./src/* /app/
RUN go build -o /app/poker_server .

FROM busybox:latest

COPY --from=builder /app/poker_server .
CMD ["./poker_server"]