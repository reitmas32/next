package next

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/rafa/next/internal/api"
	"github.com/rafa/next/internal/config"
	"github.com/spf13/cobra"
)

var versionsAccount string

var versionsCmd = &cobra.Command{
	Use:   "versions <library>",
	Short: "Lista todas las versiones (tags) de una librería",
	Long: `Lista todas las versiones (tags) de una librería.
Los tags se muestran ordenados por fecha, de más reciente a más antiguo.

Ejemplo:
  next versions fundation --account gitlab-main`,
	Args: cobra.ExactArgs(1),
	RunE: runVersions,
}

func init() {
	versionsCmd.Flags().StringVarP(&versionsAccount, "account", "a", "", "Nombre de la cuenta a usar")
}

func runVersions(cmd *cobra.Command, args []string) error {
	library := args[0]

	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		color.Red("✗ Error al cargar configuración: %v", err)
		return err
	}

	// Obtener cuenta
	account, err := cfg.GetAccount(versionsAccount)
	if err != nil {
		color.Red("✗ %v", err)
		return err
	}

	// Crear cliente del proveedor
	provider, err := api.NewProvider(account.Provider, account.Domain, account.Token)
	if err != nil {
		color.Red("✗ Error al crear cliente: %v", err)
		return err
	}

	// Obtener versiones
	versions, err := provider.ListVersions(library)
	if err != nil {
		color.Red("✗ Error al obtener versiones: %v", err)
		return err
	}

	if len(versions) == 0 {
		color.Yellow("No se encontraron versiones para '%s'", library)
		return nil
	}

	// Mostrar versiones
	blue := color.New(color.FgBlue)
	gray := color.New(color.FgWhite)

	fmt.Println()
	for _, v := range versions {
		blue.Printf("%-12s", v.Name)
		gray.Printf(" %s\n", v.Date)
	}
	fmt.Println()

	return nil
}
