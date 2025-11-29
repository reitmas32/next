package ui

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	// Success para mensajes de éxito (verde)
	Success = color.New(color.FgGreen)

	// Error para mensajes de error (rojo)
	Error = color.New(color.FgRed)

	// Warning para advertencias (amarillo)
	Warning = color.New(color.FgYellow)

	// Info para información general (azul)
	Info = color.New(color.FgBlue)

	// Name para nombres de librerías (cyan)
	Name = color.New(color.FgCyan)

	// Provider para proveedores (magenta)
	Provider = color.New(color.FgMagenta)

	// Meta para metadatos como fechas y URLs (gris/blanco)
	Meta = color.New(color.FgWhite)

	// Bold para texto en negrita
	Bold = color.New(color.Bold)
)

// SuccessIcon imprime un mensaje de éxito con checkmark
func SuccessIcon(format string, args ...interface{}) {
	Success.Printf("✔ "+format+"\n", args...)
}

// ErrorIcon imprime un mensaje de error con X
func ErrorIcon(format string, args ...interface{}) {
	Error.Printf("✗ "+format+"\n", args...)
}

// WarningIcon imprime una advertencia con signo de exclamación
func WarningIcon(format string, args ...interface{}) {
	Warning.Printf("! "+format+"\n", args...)
}

// InfoIcon imprime información con flecha
func InfoIcon(format string, args ...interface{}) {
	Info.Printf("→ "+format+"\n", args...)
}

// PrintLibrary imprime información de una librería con colores
func PrintLibrary(name, description, provider, domain string) {
	Name.Println(name)
	if description != "" {
		Meta.Printf("  %s\n", description)
	}
	Provider.Printf("  proveedor: %s\n", provider)
	Meta.Printf("  dominio: %s\n", domain)
	fmt.Println()
}

// PrintVersion imprime una versión con colores
func PrintVersion(version, date string) {
	Info.Printf("%-12s", version)
	Meta.Printf(" %s\n", date)
}

// PrintHeader imprime un encabezado en negrita
func PrintHeader(text string) {
	Bold.Println(text)
	fmt.Println()
}

