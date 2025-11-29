package next

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/rafa/next/internal/api"
	"github.com/rafa/next/internal/config"
	"github.com/spf13/cobra"
)

var listAccount string

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista todas las librerías Go disponibles en una cuenta",
	Long: `Lista todas las librerías Go disponibles en una cuenta.
Solo muestra repositorios que contengan un archivo go.mod.

Ejemplo:
  next list --account gitlab-main`,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVarP(&listAccount, "account", "a", "", "Nombre de la cuenta a usar")
}

func runList(cmd *cobra.Command, args []string) error {
	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		color.Red("✗ Error al cargar configuración: %v", err)
		return err
	}

	// Obtener cuenta
	account, err := cfg.GetAccount(listAccount)
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

	// Obtener librerías Go
	libraries, err := provider.ListGoLibraries()
	if err != nil {
		color.Red("✗ Error al listar librerías: %v", err)
		return err
	}

	if len(libraries) == 0 {
		color.Yellow("No se encontraron librerías Go en esta cuenta")
		return nil
	}

	// Mostrar librerías
	cyan := color.New(color.FgCyan)
	magenta := color.New(color.FgMagenta)
	gray := color.New(color.FgWhite)

	fmt.Println()
	for _, lib := range libraries {
		cyan.Println(lib.Name)
		if lib.Description != "" {
			gray.Printf("  %s\n", lib.Description)
		}
	}

	fmt.Println()
	magenta.Printf("proveedor: %s\n", account.Provider)
	gray.Printf("dominio: %s\n", account.Domain)

	return nil
}

