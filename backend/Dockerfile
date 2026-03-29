# Default uses DaoCloud mirror of Docker Hub (same as docker-compose defaults). Override: --build-arg GOLANG_IMAGE=golang:1.22-alpine
ARG GOLANG_IMAGE=docker.m.daocloud.io/library/golang:1.22-alpine
FROM ${GOLANG_IMAGE} AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /xledger ./cmd/api

# Literal FROM avoids ARG default parsing issues with ':' in image refs; change here if you use another registry.
FROM docker.m.daocloud.io/library/alpine:3.19

RUN apk add --no-cache tzdata

WORKDIR /app

COPY --from=builder /xledger /app/xledger
COPY migrations /app/migrations

EXPOSE 8080

CMD ["/app/xledger"]
