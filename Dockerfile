FROM golang AS builder

COPY ./src/* /app/
COPY go.mod /app/
COPY go.sum /app/

WORKDIR /app
RUN go mod download
RUN go build -o /app/poker_server .

FROM busybox

COPY --from=builder /app/poker_server .
CMD ["./poker_server"]