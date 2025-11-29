#!/bin/bash
set -e

echo "🚀 Calculator - Ejemplo de next CLI"
echo "===================================="
echo ""

# Función para mostrar ayuda
show_help() {
    echo "Uso: docker run calculator [comando]"
    echo ""
    echo "Comandos:"
    echo "  build     - Configura dependencias y compila (default)"
    echo "  run       - Ejecuta la calculadora"
    echo "  shell     - Abre un shell interactivo"
    echo "  help      - Muestra esta ayuda"
    echo ""
    echo "Variables de entorno requeridas:"
    echo "  NEXT_PROVIDER     - Proveedor: github o gitlab"
    echo "  NEXT_TOKEN        - Token de acceso"
    echo "  NEXT_ACCOUNT_NAME - Nombre de la cuenta (opcional)"
    echo "  NEXT_OWNERS       - Owners separados por coma (opcional)"
    echo ""
    echo "Ejemplo:"
    echo "  docker run -e NEXT_PROVIDER=github -e NEXT_TOKEN=ghp_xxx calculator"
}

# Función para hacer login con next
do_login() {
    if [ -z "$NEXT_PROVIDER" ] || [ -z "$NEXT_TOKEN" ]; then
        echo "⚠️  Variables NEXT_PROVIDER y NEXT_TOKEN son requeridas"
        echo ""
        show_help
        exit 1
    fi

    # Determinar URL según proveedor
    case "$NEXT_PROVIDER" in
        github)
            NEXT_URL="https://github.com"
            ;;
        gitlab)
            NEXT_URL="${NEXT_URL:-https://gitlab.com}"
            ;;
        *)
            echo "❌ Proveedor no soportado: $NEXT_PROVIDER"
            exit 1
            ;;
    esac

    # Nombre de cuenta por defecto
    ACCOUNT_NAME="${NEXT_ACCOUNT_NAME:-$NEXT_PROVIDER}"

    echo "🔐 Configurando cuenta '$ACCOUNT_NAME'..."

    # Construir comando de login
    LOGIN_CMD="next login --provider $NEXT_PROVIDER --url $NEXT_URL --token $NEXT_TOKEN --name $ACCOUNT_NAME"

    if [ -n "$NEXT_OWNERS" ]; then
        LOGIN_CMD="$LOGIN_CMD --owners $NEXT_OWNERS"
    fi

    # Ejecutar login
    $LOGIN_CMD

    echo ""
}

# Función para configurar dependencias
do_check() {
    echo "🔍 Verificando dependencias privadas..."
    next check || true
    echo ""
}

# Función para descargar dependencias
do_tidy() {
    echo "📦 Descargando dependencias..."
    go mod tidy
    echo ""
}

# Función para compilar
do_build() {
    echo "🔨 Compilando calculator..."
    go build -o calculator .
    echo "✔ Compilación exitosa"
    echo ""
}

# Procesar comando
case "${1:-build}" in
    build)
        do_login
        do_check
        do_tidy
        do_build
        echo "✅ Listo! Ejecuta: ./calculator"
        ;;
    run)
        do_login
        do_check
        do_tidy
        do_build
        echo "🧮 Iniciando calculator..."
        echo ""
        ./calculator
        ;;
    shell)
        echo "🐚 Abriendo shell..."
        exec /bin/bash
        ;;
    help|--help|-h)
        show_help
        ;;
    *)
        echo "❌ Comando desconocido: $1"
        echo ""
        show_help
        exit 1
        ;;
esac

