# --- Стадия сборки ---
FROM golang:1.23 as builder

# Рабочая директория внутри контейнера
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o avito-app ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/avito-app .
COPY --from=builder /app/migrations ./migrations

CMD ["./avito-app"]
