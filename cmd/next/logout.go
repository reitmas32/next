package next

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/reitmas32/next/internal/config"
	"github.com/spf13/cobra"
)

var (
	logoutAll   bool
	logoutForce bool
)

var logoutCmd = &cobra.Command{
	Use:   "logout [nombre-cuenta]",
	Short: "Elimina una cuenta registrada",
	Long: `Elimina una cuenta del archivo de configuración.

Ejemplos:
  # Eliminar cuenta específica
  next logout personal

  # Eliminar con confirmación
  next logout trabajo

  # Eliminar sin confirmación
  next logout trabajo -f

  # Eliminar todas las cuentas
  next logout --all`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLogout,
}

func init() {
	logoutCmd.Flags().BoolVarP(&logoutAll, "all", "a", false, "Eliminar todas las cuentas")
	logoutCmd.Flags().BoolVarP(&logoutForce, "force", "f", false, "No pedir confirmación")

	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)

	// Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		color.Red("✗ Error al cargar configuración: %v", err)
		return err
	}

	if len(cfg.Accounts) == 0 {
		yellow.Println("No hay cuentas configuradas")
		return nil
	}

	// Eliminar todas las cuentas
	if logoutAll {
		if !logoutForce {
			fmt.Printf("¿Eliminar TODAS las cuentas (%d)? [s/N]: ", len(cfg.Accounts))
			if !confirmAction() {
				yellow.Println("Operación cancelada")
				return nil
			}
		}

		count := len(cfg.Accounts)
		cfg.Accounts = []config.Account{}

		if err := cfg.Save(); err != nil {
			color.Red("✗ Error al guardar configuración: %v", err)
			return err
		}

		green.Printf("✔ %d cuenta(s) eliminada(s)\n", count)
		return nil
	}

	// Verificar que se proporcionó nombre de cuenta
	if len(args) == 0 {
		// Mostrar cuentas disponibles
		yellow.Println("Especifica la cuenta a eliminar:")
		fmt.Println()
		for _, acc := range cfg.Accounts {
			cyan.Printf("  • %s", acc.Name)
			color.White(" (%s - %s)", acc.Provider, acc.Domain)
			if len(acc.Owners) > 0 {
				color.White(" [owners: %s]", strings.Join(acc.Owners, ", "))
			}
			fmt.Println()
		}
		fmt.Println()
		color.White("Uso: next logout <nombre-cuenta>")
		color.White("     next logout --all  (eliminar todas)")
		return nil
	}

	accountName := args[0]

	// Buscar la cuenta
	var accountToDelete *config.Account
	for _, acc := range cfg.Accounts {
		if acc.Name == accountName {
			accountToDelete = &acc
			break
		}
	}

	if accountToDelete == nil {
		color.Red("✗ Cuenta '%s' no encontrada", accountName)
		fmt.Println()
		yellow.Println("Cuentas disponibles:")
		for _, acc := range cfg.Accounts {
			cyan.Printf("  • %s\n", acc.Name)
		}
		return fmt.Errorf("cuenta no encontrada")
	}

	// Confirmar eliminación
	if !logoutForce {
		fmt.Println()
		cyan.Printf("Cuenta: %s\n", accountToDelete.Name)
		color.White("  Proveedor: %s\n", accountToDelete.Provider)
		color.White("  Dominio:   %s\n", accountToDelete.Domain)
		if len(accountToDelete.Owners) > 0 {
			color.White("  Owners:    %s\n", strings.Join(accountToDelete.Owners, ", "))
		}
		fmt.Println()
		fmt.Print("¿Eliminar esta cuenta? [s/N]: ")

		if !confirmAction() {
			yellow.Println("Operación cancelada")
			return nil
		}
	}

	// Eliminar cuenta
	if err := cfg.RemoveAccount(accountName); err != nil {
		color.Red("✗ Error al eliminar cuenta: %v", err)
		return err
	}

	if err := cfg.Save(); err != nil {
		color.Red("✗ Error al guardar configuración: %v", err)
		return err
	}

	green.Printf("✔ Cuenta '%s' eliminada correctamente\n", accountName)

	// Mostrar cuentas restantes
	if len(cfg.Accounts) > 0 {
		fmt.Println()
		color.White("Cuentas restantes: %d", len(cfg.Accounts))
		for _, acc := range cfg.Accounts {
			cyan.Printf("  • %s\n", acc.Name)
		}
	} else {
		yellow.Println("\nNo quedan cuentas configuradas")
		color.White("Usa 'next login' para agregar una nueva cuenta")
	}

	return nil
}

// confirmAction lee la entrada del usuario y retorna true si confirma
func confirmAction() bool {
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "s" || response == "si" || response == "sí" || response == "y" || response == "yes"
}
