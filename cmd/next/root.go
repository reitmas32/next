package next

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "next",
	Short: "CLI de gestión de librerías Go privadas y versionado",
	Long: `next es una herramienta CLI para administrar repositorios de librerías 
privadas o públicas, permitir autenticación con múltiples dominios 
GitLab y GitHub, listar librerías disponibles, listar versiones 
y crear nuevas versiones (tags semánticos).`,
}

// Execute ejecuta el comando raíz
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Configuración global de colores
	color.NoColor = false

	// Agregar comandos hijos
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(createVersionCmd)
	rootCmd.AddCommand(versionsCmd)
}
