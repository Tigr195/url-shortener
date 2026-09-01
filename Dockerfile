FROM golang:1.27-alpine AS builder

WORKDIR /app

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN swag init -g cmd/api/main.go
RUN go build -o bin/api ./cmd/api

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/bin/api .
COPY --from=builder /app/frontend ./frontend

EXPOSE 8080

CMD ["./api"]