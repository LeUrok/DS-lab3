FROM golang:1.27-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/rental-service .
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/rental-service .

EXPOSE 8060

CMD ["./rental-service"]