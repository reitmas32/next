# Calculator - Ejemplo de Docker con next CLI

Este ejemplo muestra cómo usar `next` en un contenedor Docker para manejar dependencias privadas.

## Estructura

```
calculator/
├── Dockerfile          # Imagen con golang y next
├── docker-compose.yml  # Configuración de servicios
├── entrypoint.sh       # Script de entrada
├── env.example         # Ejemplo de variables de entorno
├── go.mod              # Módulo Go
├── main.go             # Código de la calculadora
└── next                # Binario de next CLI (Linux)
```

## Uso rápido

### 0. Compilar el binario de next (una sola vez)

```bash
# Desde la raíz del proyecto next
cd /path/to/next
GOOS=linux GOARCH=amd64 go build -o examples/calculator/next .
```

### 1. Configurar variables de entorno

```bash
cp env.example .env
# Editar .env con tu token
```

### 2. Construir y ejecutar

```bash
# Compilar
docker-compose run calculator build

# Ejecutar la calculadora
docker-compose run calculator run

# Shell interactivo para debug
docker-compose run dev
```

### 3. Sin docker-compose

```bash
# Construir imagen
docker build -t calculator .

# Ejecutar con variables
docker run -it \
  -e NEXT_PROVIDER=github \
  -e NEXT_TOKEN=ghp_xxxxxxxxxx \
  -e NEXT_ACCOUNT_NAME=personal \
  -e NEXT_OWNERS=reitmas32 \
  calculator run
```

## Variables de entorno

| Variable | Descripción | Requerido |
|----------|-------------|-----------|
| `NEXT_PROVIDER` | `github` o `gitlab` | ✅ |
| `NEXT_TOKEN` | Token de acceso | ✅ |
| `NEXT_ACCOUNT_NAME` | Nombre de la cuenta | ❌ |
| `NEXT_OWNERS` | Owners separados por coma | ❌ |

## Comandos disponibles

```bash
# Configurar y compilar
docker-compose run calculator build

# Compilar y ejecutar
docker-compose run calculator run

# Shell para debug
docker-compose run calculator shell

# Ayuda
docker-compose run calculator help
```

## Flujo del contenedor

```
┌─────────────────────────────────────────┐
│  1. next login                          │
│     - Configura la cuenta con el token  │
├─────────────────────────────────────────┤
│  2. next check                          │
│     - Detecta dependencias privadas     │
│     - Configura GOPRIVATE               │
│     - Configura credenciales git        │
├─────────────────────────────────────────┤
│  3. go mod tidy                         │
│     - Descarga dependencias             │
├─────────────────────────────────────────┤
│  4. go build                            │
│     - Compila el proyecto               │
└─────────────────────────────────────────┘
```

## Dependencias privadas

Este ejemplo usa `github.com/reitmas32/mathutils` que puede ser privado.
El flujo de `next` se encarga de configurar todo automáticamente.

## Notas

- El binario `next` se copia directamente al contenedor
- La imagen usa `golang:1.22-bookworm` (Debian)
- Se instala `libsecret` para el keyring (encriptación de tokens)

