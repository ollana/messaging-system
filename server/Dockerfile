FROM golang:1.22-alpine

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN env GOOS=linux GOARCH=amd64 go build -o ./messaging-service .

EXPOSE 80

ENTRYPOINT ["./messaging-service"]
