FROM golang AS builder

COPY ./src/* /app/
COPY go.mod /app/
COPY go.sum /app/

WORKDIR /app
RUN go fix
RUN go build -o /app/poker_server .
CMD ["/app/poker_server"]
