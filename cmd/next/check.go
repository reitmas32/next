package next

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/rafa/next/internal/config"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Verifica y configura dependencias privadas del proyecto",
	Long: `Analiza el archivo go.mod del proyecto actual, detecta dependencias 
privadas y configura automáticamente GOPRIVATE y las credenciales 
necesarias para que 'go mod tidy' funcione correctamente.

Soporta múltiples cuentas del mismo dominio (ej: GitHub personal y trabajo).
Usa el owner del módulo para seleccionar la cuenta correcta.

Ejemplo:
  next check`,
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	gray := color.New(color.FgWhite)

	fmt.Println()
	cyan.Println("🔍 Analizando dependencias del proyecto...")
	fmt.Println()

	// Buscar go.mod en el directorio actual
	goModPath := "go.mod"
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		color.Red("✗ No se encontró go.mod en el directorio actual")
		return fmt.Errorf("go.mod no encontrado")
	}

	// Leer dependencias del go.mod
	dependencies, err := parseGoMod(goModPath)
	if err != nil {
		color.Red("✗ Error al leer go.mod: %v", err)
		return err
	}

	if len(dependencies) == 0 {
		yellow.Println("No se encontraron dependencias en go.mod")
		return nil
	}

	gray.Printf("Dependencias encontradas: %d\n\n", len(dependencies))

	// Cargar configuración de cuentas
	cfg, err := config.Load()
	if err != nil {
		color.Red("✗ Error al cargar configuración: %v", err)
		return err
	}

	// Detectar dependencias privadas usando GetAccountForModule
	var privateDeps []privateDependency
	var goprivatePatterns []string
	configuredDomains := make(map[string]bool)

	for _, dep := range dependencies {
		// Usar GetAccountForModule para encontrar la cuenta correcta
		account, err := cfg.GetAccountForModule(dep)
		if err != nil {
			continue // No hay cuenta para este módulo
		}

		domain := extractDomain(dep)
		owner := extractOwner(dep)

		privateDeps = append(privateDeps, privateDependency{
			Module:  dep,
			Domain:  domain,
			Owner:   owner,
			Account: account,
		})

		// Agregar patrón a GOPRIVATE
		if !configuredDomains[domain] {
			goprivatePatterns = append(goprivatePatterns, domain+"/*")
			configuredDomains[domain] = true
		}
	}

	if len(privateDeps) == 0 {
		green.Println("✔ No se detectaron dependencias privadas")
		gray.Println("  Todas las dependencias son públicas o no están en cuentas configuradas")
		return nil
	}

	// Mostrar dependencias privadas detectadas
	yellow.Printf("📦 Dependencias privadas detectadas: %d\n\n", len(privateDeps))

	for _, dep := range privateDeps {
		cyan.Printf("  • %s\n", dep.Module)
		if len(dep.Account.Owners) > 0 {
			gray.Printf("    cuenta: %s (owners: %s)\n", dep.Account.Name, strings.Join(dep.Account.Owners, ", "))
		} else {
			gray.Printf("    cuenta: %s (wildcard)\n", dep.Account.Name)
		}
	}
	fmt.Println()

	// Configurar GOPRIVATE
	goprivateValue := strings.Join(goprivatePatterns, ",")
	cyan.Println("⚙️  Configurando GOPRIVATE...")

	if err := setGOPRIVATE(goprivateValue); err != nil {
		color.Red("✗ Error al configurar GOPRIVATE: %v", err)
		return err
	}
	green.Printf("✔ GOPRIVATE=%s\n\n", goprivateValue)

	// Configurar credenciales de git para cada dependencia
	cyan.Println("🔐 Configurando credenciales de git...")

	// Agrupar por dominio+cuenta para no configurar múltiples veces
	configuredCredentials := make(map[string]bool)

	for _, dep := range privateDeps {
		key := fmt.Sprintf("%s:%s", dep.Domain, dep.Account.Name)
		if configuredCredentials[key] {
			continue
		}

		if err := configureGitCredentials(dep.Domain, dep.Account); err != nil {
			color.Yellow("! Advertencia al configurar %s: %v", dep.Domain, err)
		} else {
			green.Printf("✔ Credenciales configuradas para %s (cuenta: %s)\n", dep.Domain, dep.Account.Name)
		}
		configuredCredentials[key] = true
	}

	fmt.Println()
	green.Println("✔ Configuración completada")
	fmt.Println()
	gray.Println("Ahora puedes ejecutar:")
	cyan.Println("  go mod tidy")
	cyan.Println("  go build")
	fmt.Println()

	return nil
}

type privateDependency struct {
	Module  string
	Domain  string
	Owner   string
	Account *config.Account
}

// parseGoMod lee el archivo go.mod y extrae las dependencias
func parseGoMod(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var dependencies []string
	inRequireBlock := false

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Detectar inicio de bloque require
		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}

		// Detectar fin de bloque
		if inRequireBlock && line == ")" {
			inRequireBlock = false
			continue
		}

		// Require en una línea
		if strings.HasPrefix(line, "require ") && !strings.Contains(line, "(") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				dependencies = append(dependencies, parts[1])
			}
			continue
		}

		// Dependencias dentro del bloque
		if inRequireBlock && line != "" && !strings.HasPrefix(line, "//") {
			parts := strings.Fields(line)
			if len(parts) >= 1 {
				dependencies = append(dependencies, parts[0])
			}
		}
	}

	return dependencies, scanner.Err()
}

// extractDomain extrae el dominio de un módulo Go
func extractDomain(module string) string {
	parts := strings.Split(module, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return module
}

// extractOwner extrae el owner de un módulo Go
func extractOwner(module string) string {
	parts := strings.Split(module, "/")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

// setGOPRIVATE configura la variable GOPRIVATE
func setGOPRIVATE(value string) error {
	cmd := exec.Command("go", "env", "-w", "GOPRIVATE="+value)
	return cmd.Run()
}

// configureGitCredentials configura las credenciales de git para un dominio
func configureGitCredentials(domain string, account *config.Account) error {
	// Configurar git para usar el token
	var urlPattern string
	if account.Provider == "github" {
		urlPattern = fmt.Sprintf("url.https://%s:x-oauth-basic@%s/.insteadOf", account.Token, domain)
	} else {
		// GitLab usa oauth2 como username
		urlPattern = fmt.Sprintf("url.https://oauth2:%s@%s/.insteadOf", account.Token, domain)
	}

	originalURL := fmt.Sprintf("https://%s/", domain)

	cmd := exec.Command("git", "config", "--global", urlPattern, originalURL)
	if err := cmd.Run(); err != nil {
		// Intentar con .netrc como fallback
		return configureNetrc(domain, account)
	}

	return nil
}

// configureNetrc configura el archivo .netrc como fallback
func configureNetrc(domain string, account *config.Account) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	netrcPath := filepath.Join(homeDir, ".netrc")

	// Leer contenido existente
	var content string
	if data, err := os.ReadFile(netrcPath); err == nil {
		content = string(data)
	}

	// Verificar si ya existe entrada para este dominio
	if strings.Contains(content, "machine "+domain) {
		return nil // Ya configurado
	}

	// Agregar nueva entrada
	var username string
	if account.Provider == "github" {
		username = "x-oauth-basic"
	} else {
		username = "oauth2"
	}

	entry := fmt.Sprintf("\nmachine %s login %s password %s\n", domain, username, account.Token)

	f, err := os.OpenFile(netrcPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(entry)
	return err
}
