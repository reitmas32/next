# 📄 **DOCUMENTO DE REQUISITOS DE DISEÑO (DRD)**

## **Proyecto: `next` – CLI de gestión de librerías Go privadas y versionado**

---

# 1. **Propósito y visión general**

El sistema `next` es una herramienta CLI escrita en **Golang** cuyo objetivo es:

* administrar repositorios de librerías privadas o públicas;
* permitir autenticación con múltiples dominios GitLab y GitHub;
* listar librerías disponibles;
* listar versiones de una librería;
* crear nuevas versiones (tags semánticos);
* facilitar la interoperabilidad con múltiples organizaciones, proyectos y dominios.

El CLI se usará desde entornos locales y pipelines.
Debe estar diseñado para ser **extensible**, **seguro**, **rápido**, **fácil de mantener** y **agradable visualmente** utilizando colores ANSI.

---

# 2. **Objetivos del proyecto**

1. Crear un CLI robusto para la gestión de librerías Go.
2. Proveer una interfaz consistente para conectarse a GitLab/GitHub auto-hosteados o públicos.
3. Manejar múltiples cuentas o dominios simultáneamente.
4. Simplificar el versionado a través de tags (con validación semántica).
5. Dar soporte para listar repos y versiones.
6. Mantener compatibilidad con `go get` para instalar librerías.

---

# 3. **Requerimientos funcionales (RF)**

## RF-1 — Comando `next login`

* Permite autenticar un dominio y guardarlo como una cuenta.
* Soporta múltiples dominios registrados simultáneamente.
* Se debe indicar:

  * proveedor (`--provider=gitlab` o `--provider=github`)
  * URL del dominio o API endpoint
  * token de acceso (Personal Access Token o Deploy Token)
  * nombre opcional de la cuenta (alias)
* Valida el token mediante llamada a la API:

  * GitLab: `GET /user`
  * GitHub: `GET /user`
* Guarda la información en un archivo de configuración local:

  * `~/.next/config.json`
* Formato esperado del guardado:

```json
{
  "accounts": [
    {
      "name": "gitlab-main",
      "provider": "gitlab",
      "api_url": "https://gitlab.example.com/api/v4",
      "domain": "https://gitlab.example.com",
      "token": "xxxx"
    },
    {
      "name": "github-personal",
      "provider": "github",
      "api_url": "https://api.github.com",
      "domain": "https://github.com",
      "token": "xxxx"
    }
  ]
}
```

---

## RF-2 — Comando `next list`

* Lista todas las librerías Go disponibles en una cuenta.
* Permite indicar la cuenta a usar:
  `next list --account gitlab-main`
* Debe detectar repos con archivo `go.mod`.
* Debe mostrar:

  * nombre del repo
  * descripción
  * proveedor
  * URL
* El output debe ser coloreado:

  * nombre en **cyan**
  * proveedor en **magenta**
  * dominio en **gris**

---

## RF-3 — Comando `next create-version <tag>`

* Crea un tag Git semántico en el repo actual.
* Validaciones:

  * El repo debe ser un repositorio Git válido.
  * Debe determinar cuál dominio se está usando según `origin`.
  * Debe validar que no existan cambios sin commit (a menos que se use `-f`).
  * Debe validar que `<tag>` siga formato `vX.Y.Z`.
* Acciones:

  * Crear tag en el remoto usando API:

    * GitLab: `POST /projects/:id/repository/tags`
    * GitHub: `POST /repos/:repo/git/tags` + `POST refs/tags/<tag>`
* Debe imprimir mensajes coloreados:

  * Éxito → verde
  * Advertencia → amarillo
  * Error → rojo

---

## RF-4 — Comando `next <library> versions`

* Lista todas las versiones (tags) de la librería.
* Input:

  * nombre del repo o path completo
  * `--account` para dominios
* Debe mostrar tags ordenados por fecha, descendente.
* Salida coloreada:

  * versión → azul
  * fecha → gris

---

## RF-5 — Manejo de múltiples cuentas

* El usuario puede iniciar sesión en muchos dominios.
* `next` debe recordar las cuentas configuradas.
* Ejemplo:

```
next login --provider=gitlab --url=https://gitlab.company.com --token=xxx --name=company
next login --provider=github --token=xxx --name=gh
```

El comando `next list` puede elegir:

```
next list --account company
```

Si solo hay 1 cuenta configurada, se usa automáticamente.

---

## RF-6 — Colores y estética del CLI

* Se deben usar colores ANSI (paquete recomendado: `github.com/fatih/color`).

* Colores principales:

  * Verde (#00FF00) → Éxito
  * Rojo (#FF5F5F) → Errores
  * Azul → Información general
  * Cian → Nombres de librerías
  * Magenta → Proveedores
  * Amarillo → Advertencias
  * Gris → Metadatos (fechas, URLs)

* El CLI debe evitar saturar la pantalla; se busca estilo minimalista.

---

## RF-7 — Configuración local

* Archivo: `~/.next/config.json`
* Si no existe, se crea automáticamente.
* Debe controlarse la concurrencia (lock simple).
* Debe manejar errores de lectura o corrupción.

---

## RF-8 — Compatibilidad con Git local

* Implementar funciones internas:

  * obtener directorio raíz de repo
  * verificar cambios sin commit
  * obtener remote principal
  * obtener rama actual
* Debe usar comandos:

  * `git rev-parse --show-toplevel`
  * `git status --porcelain`
  * `git remote get-url origin`

---

# 4. **Requerimientos no funcionales (RNF)**

### RNF-1 — Desarrollado en Go

Version recomendada: Go 1.22+

### RNF-2 — Estructura limpia y modular

Usar patrón: **cmd + internal/**

### RNF-3 — CLI con Cobra

Debe usar:

```
github.com/spf13/cobra
```

### RNF-4 — Logs mínimos

Debe imprimir solo lo necesario.
Errores deben ir en **rojo**.

### RNF-5 — Extensible

La arquitectura debe soportar nuevos comandos y nuevos proveedores.

### RNF-6 — Alto rendimiento

Debe minimizar llamadas a la API.

---

# 5. **Arquitectura sugerida del proyecto**

```
next/
 ├── cmd/
 │    └── next/
 │         ├── root.go
 │         ├── login.go
 │         ├── list.go
 │         ├── create_version.go
 │         ├── versions.go
 │         └── utils.go
 │
 ├── internal/
 │    ├── api/
 │    │     ├── provider.go
 │    │     ├── gitlab.go
 │    │     └── github.go
 │    │
 │    ├── config/
 │    │     ├── manager.go
 │    │     └── model.go
 │    │
 │    ├── git/
 │    │     ├── local.go
 │    │     └── helpers.go
 │    │
 │    └── ui/
 │          └── colors.go
 │
 └── go.mod
```

---

# 6. **Especificación detallada de la interfaz de comandos**

## 6.1 `next login`

```
next login \
    --provider gitlab \
    --url https://gitlab.example.com \
    --token <PAT> \
    --name company
```

Salida esperada (con colores):

```
[✔] Cuenta 'company' agregada correctamente   (verde)
Proveedor: GitLab                            (magenta)
Dominio:   https://gitlab.example.com         (gris)
```

---

## 6.2 `next list`

```
next list --account company
```

Ejemplo de salida:

```
fundation                       (cyan)
authify-utils                   (cyan)
core-events                     (cyan)
provider: gitlab                (magenta)
dominio: gitlab.example.com     (gris)
```

---

## 6.3 `next create-version v1.4.0`

Validaciones:

* error si no es formato semver
* error si repo sucio

Salida:

```
[✔] Versión v1.4.0 creada exitosamente.    (verde)
Repositorio: fundation                     (cyan)
```

---

## 6.4 `next fundation versions`

```
v2.0.0     2025-11-28   (azul + gris)
v1.4.0     2025-11-22
v1.3.1     2025-11-10
v1.0.0     2025-09-01
```

---

# 7. **Flujo general del uso**

### 1. Usuario registra una cuenta:

```
next login --provider gitlab --url https://gitlab.miempresa.com --token X
```

### 2. Lista librerías:

```
next list
```

### 3. Crea un release:

```
cd fundation
next create-version v1.1.0
```

### 4. Otro proyecto lo instala:

```
go get gitlab.miempresa.com/fundation@v1.1.0
```

---

# 8. **Reglas adicionales para desarrollo**

### Convenciones:

* Todos los mensajes del CLI deben estar coloreados.
* Evitar dependencias grandes; Go puro + Cobra + Fatih Color.
* Usar interfaces limpias para Providers.
* El CLI nunca debe revelar tokens en pantalla.
* Los errores deben ser descriptivos.
* Código debe tener tests para:

  * lectura de config
  * escritura de config
  * detección de repos git
  * validación de semver

---

# 9. **Criterios de aceptación**

✓ El usuario puede autenticarse con múltiples dominios.
✓ Se pueden listar repos Go de manera coloreada.
✓ Se pueden crear tags válidos desde el CLI.
✓ Se pueden listar las versiones de una librería.
✓ El CLI funciona en Linux, macOS y Windows.
✓ Funciona sin necesidad de SDKs externos.

