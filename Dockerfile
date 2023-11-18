FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN  go build ./cmd/main.go

CMD ["./main"]

EXPOSE 3001 9090
