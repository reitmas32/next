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

	// Detectar dependencias privadas (que coincidan con dominios configurados)
	var privateDeps []privateDependency
	var goprivateList []string

	for _, dep := range dependencies {
		domain := extractDomain(dep)
		account := findAccountForDomain(cfg, domain)

		if account != nil {
			privateDeps = append(privateDeps, privateDependency{
				Module:  dep,
				Domain:  domain,
				Account: account,
			})
			goprivateList = append(goprivateList, domain+"/*")
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
		gray.Printf("    cuenta: %s (%s)\n", dep.Account.Name, dep.Account.Provider)
	}
	fmt.Println()

	// Configurar GOPRIVATE
	goprivateValue := strings.Join(unique(goprivateList), ",")
	cyan.Println("⚙️  Configurando GOPRIVATE...")

	if err := setGOPRIVATE(goprivateValue); err != nil {
		color.Red("✗ Error al configurar GOPRIVATE: %v", err)
		return err
	}
	green.Printf("✔ GOPRIVATE=%s\n\n", goprivateValue)

	// Configurar credenciales de git para cada dominio
	cyan.Println("🔐 Configurando credenciales de git...")

	for _, dep := range privateDeps {
		if err := configureGitCredentials(dep.Domain, dep.Account); err != nil {
			color.Yellow("! Advertencia al configurar %s: %v", dep.Domain, err)
		} else {
			green.Printf("✔ Credenciales configuradas para %s\n", dep.Domain)
		}
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

// findAccountForDomain busca una cuenta configurada que coincida con el dominio
func findAccountForDomain(cfg *config.Config, domain string) *config.Account {
	// Mapeo de dominios conocidos
	domainMap := map[string]string{
		"github.com": "github.com",
		"gitlab.com": "gitlab.com",
	}

	for _, account := range cfg.Accounts {
		accountDomain := strings.TrimPrefix(account.Domain, "https://")
		accountDomain = strings.TrimPrefix(accountDomain, "http://")
		accountDomain = strings.TrimSuffix(accountDomain, "/")

		// Coincidencia directa
		if accountDomain == domain {
			return &account
		}

		// Coincidencia con dominio mapeado
		if mapped, ok := domainMap[domain]; ok && accountDomain == mapped {
			return &account
		}

		// Para github.com y gitlab.com públicos
		if (domain == "github.com" && account.Provider == "github") ||
			(domain == "gitlab.com" && account.Provider == "gitlab") {
			return &account
		}
	}

	return nil
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

// unique elimina duplicados de un slice
func unique(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

