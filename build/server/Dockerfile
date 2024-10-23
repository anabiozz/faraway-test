FROM golang:1.23-alpine

# Install netcat (nc) for the health check
RUN apk add --no-cache netcat-openbsd

WORKDIR /app

COPY . .

RUN go build -o server ./cmd/server

EXPOSE 8080

CMD ["/app/server"]
