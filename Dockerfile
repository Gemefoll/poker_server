FROM golang:latest

COPY ./src/* /app/
COPY ./resources/* /app/
COPY go.mod /app/
COPY go.sum /app/

WORKDIR /app
RUN go fix

CMD ["go", "run", "."]