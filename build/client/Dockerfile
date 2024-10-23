FROM golang:1.23-alpine

WORKDIR /app

COPY . .

RUN go build -o client ./cmd/client

CMD ["/app/client"]