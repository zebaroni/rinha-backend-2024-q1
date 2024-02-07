FROM golang:latest

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY src/ src/

RUN go build -o rinha ./src/*.go

CMD ["./rinha"]