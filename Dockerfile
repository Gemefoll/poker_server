FROM golang AS builder

COPY ./src/* /app/
COPY ./resources/card_img /app/
COPY ./resources/index.html /app/
COPY ./resources/start.html /app/
COPY go.mod /app/
COPY go.sum /app/

WORKDIR /app
RUN go fix
RUN go build -o /app/poker_server .
CMD ["/app/poker_server"]
