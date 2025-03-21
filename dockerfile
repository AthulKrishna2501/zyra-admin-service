FROM golang:1.24

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o admin-service ./cmd

EXPOSE 5001

CMD ["./admin-service"]
