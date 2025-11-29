# next

CLI de gestión de librerías Go privadas y versionado.

## Instalación

```bash
go install github.com/rafa/next@latest
```

O compilar desde el código fuente:

```bash
git clone https://github.com/rafa/next.git
cd next
go build -o next .
```

## Comandos

### `next login`

Autenticar con un dominio GitLab o GitHub.

```bash
# GitHub
next login --provider github --url https://github.com --token <TOKEN> --name gh

# GitLab
next login --provider gitlab --url https://gitlab.example.com --token <TOKEN> --name company

# GitLab público
next login --provider gitlab --url https://gitlab.com --token <TOKEN> --name gitlab
```

**Flags:**
- `-p, --provider` - Proveedor: `github` o `gitlab` (requerido)
- `-u, --url` - URL del dominio (requerido)
- `-t, --token` - Token de acceso PAT (requerido)
- `-n, --name` - Alias de la cuenta (opcional)

---

### `next list`

Lista todas las librerías Go disponibles en una cuenta.

```bash
# Listar todas las librerías
next list --account gh

# Solo públicas
next list --account gh --visibility public

# Solo privadas
next list --account gh --visibility private

# De una organización específica
next list --account gh --owner myorg
```

**Flags:**
- `-a, --account` - Nombre de la cuenta a usar
- `-v, --visibility` - Filtrar: `all`, `public`, `private` (default: `all`)
- `-o, --owner` - Filtrar por usuario/organización

**Salida ejemplo:**
```
mathutils                      [público]
  Utilidades matemáticas para Go

core-lib                       [privado]
  Librería core del sistema

proveedor: github
dominio: https://github.com
```

---

### `next versions`

Lista todas las versiones (tags) de una librería.

```bash
next versions reitmas32/mathutils --account gh
```

**Flags:**
- `-a, --account` - Nombre de la cuenta a usar

**Salida ejemplo:**
```
v1.2.0       2025-11-29
v1.1.0       2025-11-28
v1.0.0       2025-11-27
```

---

### `next create-version`

Crea un tag Git semántico en el repositorio actual.

```bash
cd mi-libreria
next create-version v1.2.0
```

**Características:**
- ✅ Valida formato semántico (`vX.Y.Z`)
- ✅ Verifica que no haya cambios sin commit
- ✅ Detecta automáticamente el remote y la cuenta
- ✅ **Auto-push**: Si hay commits pendientes, los sube automáticamente
- ✅ Crea el tag vía API (GitHub/GitLab)

**Flags:**
- `-f, --force` - Forzar aunque haya cambios sin commit
- `--skip-push` - No hacer push automático

**Flujo típico:**
```bash
# Hacer cambios
git add -A
git commit -m "feat: nueva función"

# Crear versión (auto-push incluido)
next create-version v1.2.0
```

**Salida ejemplo:**
```
🔍 Verificando sincronización con origin...
📤 Subiendo 2 commit(s) pendiente(s) a origin/main...
✔ Código subido exitosamente
🏷️  Creando tag v1.2.0...

✔ Versión v1.2.0 creada exitosamente
  Repositorio: reitmas32/mathutils
  Rama: main

Para instalar esta versión:
  go get github.com/reitmas32/mathutils@v1.2.0
```

---

### `next check`

Verifica y configura dependencias privadas del proyecto.

```bash
cd mi-proyecto
next check
```

**Características:**
- ✅ Analiza `go.mod` y detecta dependencias privadas
- ✅ Configura automáticamente `GOPRIVATE`
- ✅ Configura credenciales de git para repos privados
- ✅ Después de ejecutar, `go mod tidy` funciona correctamente

**Salida ejemplo:**
```
🔍 Analizando dependencias del proyecto...

Dependencias encontradas: 3

📦 Dependencias privadas detectadas: 1

  • github.com/reitmas32/mathutils
    cuenta: gh (github)

⚙️  Configurando GOPRIVATE...
✔ GOPRIVATE=github.com/*

🔐 Configurando credenciales de git...
✔ Credenciales configuradas para github.com

✔ Configuración completada

Ahora puedes ejecutar:
  go mod tidy
  go build
```

---

## Flujo de trabajo típico

### 1. Configurar cuenta

```bash
next login --provider github --url https://github.com --token ghp_xxx --name gh
```

### 2. Crear una librería

```bash
mkdir mi-libreria
cd mi-libreria
go mod init github.com/usuario/mi-libreria

# Escribir código...

git init
git add -A
git commit -m "Initial commit"
git remote add origin git@github.com:usuario/mi-libreria.git
```

### 3. Publicar versión

```bash
next create-version v1.0.0
```

### 4. Usar la librería en otro proyecto

```bash
cd otro-proyecto

# Si es privada, primero configurar
next check

# Instalar
go get github.com/usuario/mi-libreria@v1.0.0
```

### 5. Actualizar versión

```bash
cd mi-libreria

# Hacer cambios
git add -A
git commit -m "feat: nueva funcionalidad"

# Publicar nueva versión
next create-version v1.1.0
```

---

## Configuración

La configuración se guarda en `~/.next/config.json`:

```json
{
  "accounts": [
    {
      "name": "gh",
      "provider": "github",
      "api_url": "https://api.github.com",
      "domain": "https://github.com",
      "token": "ghp_xxxxxxxxxxxx"
    },
    {
      "name": "company",
      "provider": "gitlab",
      "api_url": "https://gitlab.company.com/api/v4",
      "domain": "https://gitlab.company.com",
      "token": "glpat-xxxxxxxxxxxx"
    }
  ]
}
```

---

## Versionado semántico

| Tipo de cambio | Versión | Cuándo usar |
|----------------|---------|-------------|
| **Patch** | v1.0.**1** | Bug fixes, correcciones menores |
| **Minor** | v1.**1**.0 | Nueva funcionalidad, compatible hacia atrás |
| **Major** | v**2**.0.0 | Cambios que rompen compatibilidad |

---

## Requisitos

- Go 1.22+
- Git instalado
- Token de acceso (GitHub PAT o GitLab PAT)

---

## Licencia

MIT

