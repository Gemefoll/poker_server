FROM golang:latest

COPY . /app

WORKDIR /app
RUN go fix

ENV START_BALANCE=15000

CMD ["go", "run", "."]