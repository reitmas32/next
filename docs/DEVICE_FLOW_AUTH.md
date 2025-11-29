# RFC: Autenticación con Device Flow

## Resumen

Implementar autenticación OAuth 2.0 Device Flow para que los usuarios no tengan que crear tokens manualmente.

---

## Problema actual

```bash
# Actualmente el usuario debe:
# 1. Ir a GitHub/GitLab settings
# 2. Crear un token manualmente
# 3. Copiar el token
# 4. Pegarlo en el comando

next login --provider github --url https://github.com --token ghp_xxxx --name personal
```

**Fricción:** Muchos pasos manuales, propenso a errores, tokens con permisos incorrectos.

---

## Solución propuesta

```bash
# Nuevo flujo simplificado:
next login --provider github --name personal
```

```
🔐 Iniciando autenticación con GitHub...

Para continuar, abre este enlace en tu navegador:

  👉 https://github.com/login/device

E ingresa el código:

  🔑 WDJB-MJHT

Esperando autorización... (presiona Ctrl+C para cancelar)

✔ Autenticado como reitmas32
✔ Cuenta 'personal' agregada correctamente
  Owners: reitmas32
```

---

## Cómo funciona el Device Flow

```
┌─────────────┐                              ┌─────────────┐
│   next CLI  │                              │   GitHub    │
└──────┬──────┘                              └──────┬──────┘
       │                                            │
       │  1. POST /login/device/code                │
       │  (client_id, scope)                        │
       │ ─────────────────────────────────────────► │
       │                                            │
       │  2. Respuesta:                             │
       │  {device_code, user_code, verification_url}│
       │ ◄───────────────────────────────────────── │
       │                                            │
       │  3. Mostrar código al usuario              │
       │  ══════════════════════════════════════    │
       │                                            │
       │         ┌─────────────┐                    │
       │         │   Usuario   │                    │
       │         └──────┬──────┘                    │
       │                │                           │
       │                │  4. Abre navegador        │
       │                │  Ingresa código           │
       │                │  Autoriza la app          │
       │                │ ────────────────────────► │
       │                                            │
       │  5. Poll: POST /login/oauth/access_token   │
       │  (device_code, client_id)                  │
       │ ─────────────────────────────────────────► │
       │                                            │
       │  6. Respuesta: {access_token}              │
       │ ◄───────────────────────────────────────── │
       │                                            │
       │  7. Guardar token en ~/.next/config.json   │
       │                                            │
```

---

## Compatibilidad

| Proveedor | Device Flow | URL de autorización |
|-----------|-------------|---------------------|
| GitHub.com | ✅ Sí | `https://github.com/login/device` |
| GitLab.com | ✅ Sí (v15.9+) | `https://gitlab.com/-/profile/device` |
| GitLab Self-hosted | ⚠️ v15.9+ | `https://{domain}/-/profile/device` |
| GitHub Enterprise | ✅ Sí | `https://{domain}/login/device` |

---

## Flujo con GitLab

```bash
next login --provider gitlab --url https://gitlab.com --name gitlab-personal
```

```
🔐 Iniciando autenticación con GitLab...

Para continuar, abre este enlace en tu navegador:

  👉 https://gitlab.com/-/profile/device

E ingresa el código:

  🔑 ABCD-1234

Esperando autorización...

✔ Autenticado como reitmas32
✔ Cuenta 'gitlab-personal' agregada correctamente
```

---

## Fallback para GitLab antiguo

Si GitLab no soporta Device Flow (versiones < 15.9):

```bash
next login --provider gitlab --url https://gitlab-viejo.company.com --name company
```

```
⚠️  Este servidor GitLab no soporta Device Flow

Se abrirá el navegador para crear un token manualmente...

Permisos necesarios:
  ✓ api
  ✓ read_repository
  ✓ write_repository

Presiona Enter para abrir el navegador...
[Se abre: https://gitlab-viejo.company.com/-/profile/personal_access_tokens]

Pega el token aquí: glpat-xxxxxxxxxxxx

✔ Token válido
✔ Cuenta 'company' agregada correctamente
```

---

## Permisos (Scopes) solicitados

### GitHub
```
repo        - Acceso completo a repositorios
read:org    - Leer membresías de organizaciones
```

### GitLab
```
api         - Acceso completo a la API
read_repository  - Leer repositorios
write_repository - Escribir repositorios (para crear tags)
```

---

## Requisitos de implementación

### 1. Registrar OAuth App en GitHub

Ir a: https://github.com/settings/developers

```
Application name: next-cli
Homepage URL: https://github.com/rafa/next
Authorization callback URL: http://localhost (no se usa en Device Flow)
Enable Device Flow: ✅
```

Obtener: `CLIENT_ID`

### 2. Registrar OAuth App en GitLab

Ir a: https://gitlab.com/-/profile/applications

```
Name: next-cli
Redirect URI: urn:ietf:wg:oauth:2.0:oob
Confidential: No
Scopes: api, read_repository, write_repository
```

Obtener: `APPLICATION_ID`

### 3. Almacenar Client IDs

Los client IDs se pueden embeber en el binario (son públicos, no secretos):

```go
const (
    GitHubClientID = "Iv1.xxxxxxxxxx"
    GitLabClientID = "xxxxxxxxxxxxxxx"
)
```

---

## API Endpoints

### GitHub

```bash
# 1. Solicitar código
POST https://github.com/login/device/code
Content-Type: application/x-www-form-urlencoded

client_id=CLIENT_ID&scope=repo,read:org

# Respuesta
{
  "device_code": "xxxx",
  "user_code": "WDJB-MJHT",
  "verification_uri": "https://github.com/login/device",
  "expires_in": 900,
  "interval": 5
}

# 2. Poll para token
POST https://github.com/login/oauth/access_token
Content-Type: application/x-www-form-urlencoded

client_id=CLIENT_ID&device_code=DEVICE_CODE&grant_type=urn:ietf:params:oauth:grant-type:device_code

# Respuesta (pendiente)
{"error": "authorization_pending"}

# Respuesta (éxito)
{"access_token": "gho_xxxx", "token_type": "bearer", "scope": "repo,read:org"}
```

### GitLab

```bash
# 1. Solicitar código
POST https://gitlab.com/oauth/authorize_device
Content-Type: application/x-www-form-urlencoded

client_id=APPLICATION_ID&scope=api+read_repository+write_repository

# Respuesta
{
  "device_code": "xxxx",
  "user_code": "ABCD-1234",
  "verification_uri": "https://gitlab.com/-/profile/device",
  "expires_in": 300,
  "interval": 5
}

# 2. Poll para token
POST https://gitlab.com/oauth/token
Content-Type: application/x-www-form-urlencoded

client_id=APPLICATION_ID&device_code=DEVICE_CODE&grant_type=urn:ietf:params:oauth:grant-type:device_code

# Respuesta (éxito)
{"access_token": "xxxx", "token_type": "Bearer", "scope": "api read_repository write_repository"}
```

---

## Estructura del código propuesto

```
internal/
  auth/
    device_flow.go      # Lógica común del Device Flow
    github_auth.go      # Implementación GitHub
    gitlab_auth.go      # Implementación GitLab
    fallback.go         # Fallback a token manual
```

### Interfaz propuesta

```go
type AuthProvider interface {
    // SupportsDeviceFlow verifica si el servidor soporta Device Flow
    SupportsDeviceFlow() bool
    
    // StartDeviceFlow inicia el flujo y retorna el código para el usuario
    StartDeviceFlow() (*DeviceFlowResponse, error)
    
    // PollForToken espera a que el usuario autorice
    PollForToken(deviceCode string, interval int) (string, error)
    
    // GetManualTokenURL retorna URL para crear token manualmente
    GetManualTokenURL() string
}

type DeviceFlowResponse struct {
    DeviceCode      string
    UserCode        string
    VerificationURL string
    ExpiresIn       int
    Interval        int
}
```

---

## Nuevo comando login

```bash
# Uso básico (Device Flow automático)
next login --provider github --name personal

# Con owners
next login --provider github --name trabajo --owners mi-empresa

# Forzar token manual
next login --provider github --name personal --manual

# GitLab
next login --provider gitlab --url https://gitlab.com --name gl

# GitLab self-hosted
next login --provider gitlab --url https://gitlab.company.com --name company
```

---

## Seguridad

### Ventajas del Device Flow:
- ✅ El token nunca pasa por el portapapeles
- ✅ El usuario ve exactamente qué permisos autoriza
- ✅ Tokens revocables desde GitHub/GitLab
- ✅ No requiere que el usuario entienda scopes

### Consideraciones:
- ⚠️ Los tokens siguen guardándose en `~/.next/config.json` (texto plano)
- 🔮 Futuro: Integrar con keychain del sistema

---

## Cronograma estimado

| Tarea | Tiempo |
|-------|--------|
| Registrar OAuth Apps | 30 min |
| Implementar Device Flow GitHub | 2-3 horas |
| Implementar Device Flow GitLab | 2-3 horas |
| Implementar fallback manual | 1 hora |
| Tests | 2 horas |
| Documentación | 1 hora |
| **Total** | **~10 horas** |

---

## Referencias

- [GitHub Device Flow](https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps#device-flow)
- [GitLab Device Flow](https://docs.gitlab.com/ee/api/oauth2.html#device-authorization-grant)
- [RFC 8628 - OAuth 2.0 Device Authorization Grant](https://datatracker.ietf.org/doc/html/rfc8628)

---

## Preguntas abiertas

1. **¿Crear una GitHub App o OAuth App?**
   - GitHub App: Más features, más complejo
   - OAuth App: Más simple, suficiente para nuestro caso
   - **Recomendación:** OAuth App

2. **¿Dónde hospedar la app OAuth?**
   - Puede estar en mi cuenta personal de GitHub
   - O en una organización dedicada al proyecto

3. **¿Qué hacer con GitLab self-hosted antiguo?**
   - Fallback a token manual
   - Mostrar instrucciones claras

4. **¿Guardar refresh tokens?**
   - Tokens de Device Flow pueden ser de larga duración
   - Evaluar si vale la pena implementar refresh

---

## Decisión

[ ] Aprobar implementación
[ ] Modificaciones necesarias
[ ] Rechazar / diferir

