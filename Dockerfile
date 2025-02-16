#Docker Pipeline
FROM golang:1.23.2-alpine3.20 as builder
WORKDIR /app
COPY . /app
# Создаем директорию bin, если она не существует
RUN go mod download && \
    go build -o ./bin/avito_coin_service ./cmd

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /app/bin/avito_coin_service ./bin/avito_coin_service
EXPOSE 9000
ENTRYPOINT ["./bin/avito_coin_service"]
