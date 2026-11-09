FROM golang:1.26-alpine3.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o build .

FROM alpine:3.24

RUN apk add --no-cache wget

WORKDIR /app

COPY --from=builder /app/build .
COPY --from=builder /app/migrations ./migrations

EXPOSE 3100

CMD ["./build"]
