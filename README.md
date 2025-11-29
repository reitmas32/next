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
# Cuenta básica (acepta todos los repos del dominio)
next login --provider github --url https://github.com --token <TOKEN> --name personal

# Cuenta con owners específicos (solo para repos de esos usuarios/orgs)
next login --provider github --url https://github.com --token <TOKEN> --name trabajo --owners mi-empresa,empresa-tools

# GitLab
next login --provider gitlab --url https://gitlab.example.com --token <TOKEN> --name company

# GitLab con owners
next login --provider gitlab --url https://gitlab.com --token <TOKEN> --name gitlab-work --owners my-group
```

**Flags:**
- `-p, --provider` - Proveedor: `github` o `gitlab` (requerido)
- `-u, --url` - URL del dominio (requerido)
- `-t, --token` - Token de acceso PAT (requerido)
- `-n, --name` - Alias de la cuenta (opcional)
- `-o, --owners` - Usuarios/organizaciones que maneja esta cuenta (opcional)

---

### Múltiples cuentas del mismo dominio

Puedes tener varias cuentas de GitHub (o GitLab) con diferentes tokens. Usa `--owners` para especificar qué usuarios/organizaciones maneja cada cuenta:

```bash
# Cuenta personal (solo repos de reitmas32)
next login --provider github --url https://github.com \
    --token ghp_personal_xxx \
    --name personal \
    --owners reitmas32

# Cuenta de trabajo (repos de la empresa)
next login --provider github --url https://github.com \
    --token ghp_empresa_xxx \
    --name trabajo \
    --owners mi-empresa,empresa-tools

# Cuenta fallback (sin owners = acepta cualquier otro repo)
next login --provider github --url https://github.com \
    --token ghp_otro_xxx \
    --name github-otros
```

**Lógica de selección:**

| Módulo | Cuenta usada |
|--------|--------------|
| `github.com/reitmas32/mathutils` | `personal` ✓ |
| `github.com/mi-empresa/core-lib` | `trabajo` ✓ |
| `github.com/empresa-tools/utils` | `trabajo` ✓ |
| `github.com/otro-user/lib` | `github-otros` (fallback) |

**Prioridad:** Cuentas con `owners` > Cuentas sin `owners` (wildcard)

---

### `next list`

Lista todas las librerías Go disponibles en una cuenta.

```bash
# Listar todas las librerías
next list --account personal

# Solo públicas
next list --account personal --visibility public

# Solo privadas
next list --account trabajo --visibility private

# De una organización específica
next list --account trabajo --owner mi-empresa
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
next versions reitmas32/mathutils --account personal
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
- ✅ Detecta automáticamente el remote y la cuenta correcta (por owner)
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
  Cuenta: personal

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
- ✅ Selecciona la cuenta correcta para cada dependencia (por owner)
- ✅ Configura automáticamente `GOPRIVATE`
- ✅ Configura credenciales de git para repos privados
- ✅ Después de ejecutar, `go mod tidy` funciona correctamente

**Salida ejemplo:**
```
🔍 Analizando dependencias del proyecto...

Dependencias encontradas: 3

📦 Dependencias privadas detectadas: 2

  • github.com/reitmas32/mathutils
    cuenta: personal (owners: reitmas32)
  • github.com/mi-empresa/core-lib
    cuenta: trabajo (owners: mi-empresa, empresa-tools)

⚙️  Configurando GOPRIVATE...
✔ GOPRIVATE=github.com/*

🔐 Configurando credenciales de git...
✔ Credenciales configuradas para github.com (cuenta: personal)
✔ Credenciales configuradas para github.com (cuenta: trabajo)

✔ Configuración completada

Ahora puedes ejecutar:
  go mod tidy
  go build
```

---

## Flujo de trabajo típico

### 1. Configurar cuentas

```bash
# Cuenta personal
next login --provider github --url https://github.com \
    --token ghp_personal_xxx --name personal --owners mi-usuario

# Cuenta de trabajo (opcional)
next login --provider github --url https://github.com \
    --token ghp_trabajo_xxx --name trabajo --owners mi-empresa
```

### 2. Crear una librería

```bash
mkdir mi-libreria
cd mi-libreria
go mod init github.com/mi-usuario/mi-libreria

# Escribir código...

git init
git add -A
git commit -m "Initial commit"
git remote add origin git@github.com:mi-usuario/mi-libreria.git
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
go get github.com/mi-usuario/mi-libreria@v1.0.0
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
      "name": "personal",
      "provider": "github",
      "api_url": "https://api.github.com",
      "domain": "https://github.com",
      "token": "ghp_xxxxxxxxxxxx",
      "owners": ["reitmas32"]
    },
    {
      "name": "trabajo",
      "provider": "github",
      "api_url": "https://api.github.com",
      "domain": "https://github.com",
      "token": "ghp_yyyyyyyyyyyy",
      "owners": ["mi-empresa", "empresa-tools"]
    },
    {
      "name": "company-gitlab",
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
