package next

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/rafa/next/internal/config"
	"github.com/spf13/cobra"
)

var accountsCmd = &cobra.Command{
	Use:     "accounts",
	Aliases: []string{"acc", "whoami"},
	Short:   "Lista las cuentas configuradas",
	Long: `Muestra todas las cuentas configuradas en next.

Ejemplos:
  next accounts
  next acc
  next whoami`,
	RunE: runAccounts,
}

func init() {
	rootCmd.AddCommand(accountsCmd)
}

func runAccounts(cmd *cobra.Command, args []string) error {
	cyan := color.New(color.FgCyan)
	magenta := color.New(color.FgMagenta)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	gray := color.New(color.FgWhite)

	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		color.Red("✗ Error al cargar configuración: %v", err)
		return err
	}

	if len(cfg.Accounts) == 0 {
		yellow.Println("No hay cuentas configuradas")
		fmt.Println()
		gray.Println("Usa 'next login' para agregar una cuenta:")
		cyan.Println("  next login --provider github --url https://github.com --token <TOKEN> --name personal")
		return nil
	}

	fmt.Println()
	green.Printf("📋 Cuentas configuradas: %d\n", len(cfg.Accounts))
	fmt.Println()

	for i, acc := range cfg.Accounts {
		// Nombre de la cuenta
		cyan.Printf("  %d. %s\n", i+1, acc.Name)

		// Proveedor
		magenta.Printf("     Proveedor: %s\n", acc.Provider)

		// Dominio
		gray.Printf("     Dominio:   %s\n", acc.Domain)

		// Owners
		if len(acc.Owners) > 0 {
			gray.Printf("     Owners:    %s\n", strings.Join(acc.Owners, ", "))
		} else {
			gray.Printf("     Owners:    * (todos)\n")
		}

		// Token (oculto)
		tokenPreview := maskToken(acc.Token)
		gray.Printf("     Token:     %s\n", tokenPreview)

		fmt.Println()
	}

	return nil
}

// maskToken oculta la mayor parte del token mostrando solo inicio y fin
func maskToken(token string) string {
	if len(token) <= 8 {
		return "****"
	}

	prefix := token[:4]
	suffix := token[len(token)-4:]
	masked := strings.Repeat("*", 8)

	return fmt.Sprintf("%s%s%s", prefix, masked, suffix)
}

