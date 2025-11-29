# Dockerfile para calculator - ejemplo de uso de next CLI
FROM golang:1.22-bookworm

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

# Variables de entorno para configuración
# Estas se pasan al ejecutar el container
ENV NEXT_PROVIDER=""
ENV NEXT_TOKEN=""
ENV NEXT_ACCOUNT_NAME=""
ENV NEXT_OWNERS=""

# Copiar script de entrada
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Script para configurar next y compilar
ENTRYPOINT ["/entrypoint.sh"]

# Comando por defecto: compilar y ejecutar
CMD ["build"]

