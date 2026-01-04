FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files from root
COPY go.mod go.sum ./
RUN go mod download

# Copy entire source code
COPY . .

# Build the service from cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o /payment-service ./services/payment-service/cmd

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /payment-service .

EXPOSE 8084

CMD ["./payment-service"]
