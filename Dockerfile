FROM golang:1.23.2

COPY . /app
WORKDIR /app/api/cmd/

RUN go mod download

RUN go build -o ./server

EXPOSE 5000

CMD ["./server"]
