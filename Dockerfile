# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Instalar dependências para compilação
RUN apk add --no-cache git

# Copiar go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copiar código fonte
COPY . .

# Build da aplicação
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copiar binário da aplicação
COPY --from=builder /app/main .

# Expor porta
EXPOSE 8080

# Comando para executar
CMD ["./main"]
