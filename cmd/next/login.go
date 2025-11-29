package next

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/rafa/next/internal/api"
	"github.com/rafa/next/internal/config"
	"github.com/spf13/cobra"
)

var (
	loginProvider string
	loginURL      string
	loginToken    string
	loginName     string
	loginOwners   string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Autenticar con un dominio GitLab o GitHub",
	Long: `Permite autenticar un dominio y guardarlo como una cuenta.
Soporta múltiples dominios registrados simultáneamente.

Para manejar múltiples cuentas del mismo dominio (ej: GitHub personal y trabajo),
usa --owners para especificar qué usuarios/organizaciones maneja cada cuenta.

Ejemplos:
  # Cuenta básica (acepta todos los repos del dominio)
  next login --provider github --url https://github.com --token <PAT> --name personal

  # Cuenta con owners específicos (solo para repos de esos usuarios/orgs)
  next login --provider github --url https://github.com --token <PAT> --name trabajo --owners mi-empresa,empresa-tools

  # GitLab con owners
  next login --provider gitlab --url https://gitlab.com --token <PAT> --name gitlab-work --owners company-group`,
	RunE: runLogin,
}

func init() {
	loginCmd.Flags().StringVarP(&loginProvider, "provider", "p", "", "Proveedor: gitlab o github (requerido)")
	loginCmd.Flags().StringVarP(&loginURL, "url", "u", "", "URL del dominio o API endpoint (requerido)")
	loginCmd.Flags().StringVarP(&loginToken, "token", "t", "", "Token de acceso (PAT o Deploy Token) (requerido)")
	loginCmd.Flags().StringVarP(&loginName, "name", "n", "", "Nombre/alias de la cuenta (opcional)")
	loginCmd.Flags().StringVarP(&loginOwners, "owners", "o", "", "Usuarios/organizaciones que maneja esta cuenta, separados por coma (opcional)")

	loginCmd.MarkFlagRequired("provider")
	loginCmd.MarkFlagRequired("url")
	loginCmd.MarkFlagRequired("token")
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Validar proveedor
	if loginProvider != "gitlab" && loginProvider != "github" {
		return fmt.Errorf("proveedor inválido: %s (use 'gitlab' o 'github')", loginProvider)
	}

	// Crear cliente del proveedor
	provider, err := api.NewProvider(loginProvider, loginURL, loginToken)
	if err != nil {
		color.Red("✗ Error al crear cliente: %v", err)
		return err
	}

	// Validar token
	user, err := provider.ValidateToken()
	if err != nil {
		color.Red("✗ Error de autenticación: %v", err)
		return err
	}

	// Determinar nombre de la cuenta
	accountName := loginName
	if accountName == "" {
		accountName = fmt.Sprintf("%s-%s", loginProvider, user)
	}

	// Parsear owners
	var owners []string
	if loginOwners != "" {
		for _, o := range strings.Split(loginOwners, ",") {
			o = strings.TrimSpace(o)
			if o != "" {
				owners = append(owners, o)
			}
		}
	}

	// Crear cuenta
	account := config.Account{
		Name:     accountName,
		Provider: loginProvider,
		APIURL:   provider.GetAPIURL(),
		Domain:   loginURL,
		Token:    loginToken,
		Owners:   owners,
	}

	// Guardar en configuración
	cfg, err := config.Load()
	if err != nil {
		color.Red("✗ Error al cargar configuración: %v", err)
		return err
	}

	cfg.AddAccount(account)

	if err := cfg.Save(); err != nil {
		color.Red("✗ Error al guardar configuración: %v", err)
		return err
	}

	// Mostrar éxito
	color.Green("✔ Cuenta '%s' agregada correctamente", accountName)
	color.Magenta("Proveedor: %s", loginProvider)
	color.White("Dominio:   %s", loginURL)

	if len(owners) > 0 {
		color.Cyan("Owners:    %s", strings.Join(owners, ", "))
	} else {
		color.White("Owners:    * (todos)")
	}

	return nil
}
