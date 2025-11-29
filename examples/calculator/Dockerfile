# Dockerfile para calculator - ejemplo de uso de next CLI
FROM golang:1.22-bookworm AS builder

# Argumentos de build para autenticación con next
ARG NEXT_PROVIDER=github
ARG NEXT_TOKEN
ARG NEXT_ACCOUNT_NAME=gh
ARG NEXT_OWNERS
ARG NEXT_URL

# Instalar dependencias necesarias para el keyring (libsecret)
RUN apt-get update && apt-get install -y \
    libsecret-1-0 \
    libsecret-1-dev \
    gnome-keyring \
    dbus-x11 \
    && rm -rf /var/lib/apt/lists/*

# Crear directorio de trabajo
WORKDIR /app

# Copiar el binario de next
COPY next /usr/local/bin/next
RUN chmod +x /usr/local/bin/next

# Copiar archivos del proyecto
COPY go.mod go.sum* ./
COPY *.go ./

# Login con next
RUN echo "🔐 Configurando cuenta '$NEXT_ACCOUNT_NAME'..." && \
    if [ -n "$NEXT_OWNERS" ]; then \
        next login --provider "$NEXT_PROVIDER" --url "$NEXT_URL" --token "$NEXT_TOKEN" --name "$NEXT_ACCOUNT_NAME" --owners "$NEXT_OWNERS"; \
    else \
        next login --provider "$NEXT_PROVIDER" --url "$NEXT_URL" --token "$NEXT_TOKEN" --name "$NEXT_ACCOUNT_NAME"; \
    fi

# Verificar dependencias privadas
RUN echo "🔍 Verificando dependencias privadas..." && \
    next check || true

# Descargar dependencias
RUN echo "📦 Descargando dependencias..." && \
    go mod tidy

# Compilar
RUN echo "🔨 Compilando calculator..." && \
    go build -o calculator . && \
    echo "✅ Compilación exitosa"

# Etapa final - imagen mínima
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copiar solo el binario compilado
COPY --from=builder /app/calculator /app/calculator

# Ejecutar la calculadora
CMD ["./calculator"]
