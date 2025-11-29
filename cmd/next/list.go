package next

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/rafa/next/internal/api"
	"github.com/rafa/next/internal/config"
	"github.com/spf13/cobra"
)

var (
	listAccount    string
	listVisibility string
	listOwner      string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lista todas las librerías Go disponibles en una cuenta",
	Long: `Lista todas las librerías Go disponibles en una cuenta.
Solo muestra repositorios que contengan un archivo go.mod.

Soporta tanto repositorios públicos como privados.

Ejemplo:
  next list --account gitlab-main
  next list --visibility public
  next list --owner myorg --visibility private`,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVarP(&listAccount, "account", "a", "", "Nombre de la cuenta a usar")
	listCmd.Flags().StringVarP(&listVisibility, "visibility", "v", "all", "Filtrar por visibilidad: all, public, private")
	listCmd.Flags().StringVarP(&listOwner, "owner", "o", "", "Filtrar por usuario/organización específico")
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

	// Configurar opciones de listado
	opts := api.ListOptions{
		Owner: listOwner,
	}

	switch listVisibility {
	case "public":
		opts.Visibility = api.VisibilityPublic
	case "private":
		opts.Visibility = api.VisibilityPrivate
	default:
		opts.Visibility = api.VisibilityAll
	}

	// Obtener librerías Go
	libraries, err := provider.ListGoLibrariesWithOptions(opts)
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
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)

	fmt.Println()
	for _, lib := range libraries {
		cyan.Printf("%-30s", lib.Name)

		// Mostrar badge de visibilidad
		if lib.Visibility == "public" {
			green.Printf(" [público]")
		} else {
			yellow.Printf(" [privado]")
		}
		fmt.Println()

		if lib.Description != "" {
			gray.Printf("  %s\n", lib.Description)
		}
	}

	fmt.Println()
	magenta.Printf("proveedor: %s\n", account.Provider)
	gray.Printf("dominio: %s\n", account.Domain)

	if listVisibility != "all" {
		gray.Printf("filtro: %s\n", listVisibility)
	}

	return nil
}
