FROM golang:1.27-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/cars-service .

FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/cars-service .

EXPOSE 8070

CMD ["./cars-service"]