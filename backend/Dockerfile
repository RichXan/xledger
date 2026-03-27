FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /xledger ./cmd/api

FROM alpine:3.19

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /xledger /app/xledger
COPY migrations /app/migrations

EXPOSE 8080

CMD ["/app/xledger"]
