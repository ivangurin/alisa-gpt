# Сборка: статический бинарник без CGO (зависимости чистого Go).
FROM golang:1.27-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/alisa-gpt ./cmd/app/main.go

# Рантайм: бинарник + CA-сертификаты для TLS к api.openai.com:443.
# Конфигурация — только из окружения (-e OPENAI_API_KEY=... и опциональные
# переопределения), файлов конфига в образе нет.
FROM alpine:3.24
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /bin/alisa-gpt /bin/alisa-gpt
USER nobody
EXPOSE 8080
ENTRYPOINT ["/bin/alisa-gpt"]
