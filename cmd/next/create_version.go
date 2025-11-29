package next

import (
	"fmt"
	"regexp"

	"github.com/fatih/color"
	"github.com/rafa/next/internal/api"
	"github.com/rafa/next/internal/config"
	"github.com/rafa/next/internal/git"
	"github.com/spf13/cobra"
)

var forceVersion bool

var createVersionCmd = &cobra.Command{
	Use:   "create-version <tag>",
	Short: "Crea un tag Git semántico en el repositorio actual",
	Long: `Crea un tag Git semántico en el repositorio actual.

El tag debe seguir el formato semántico: vX.Y.Z

Validaciones:
  - El directorio actual debe ser un repositorio Git válido
  - No deben existir cambios sin commit (usar -f para forzar)
  - El tag debe seguir el formato vX.Y.Z

Ejemplo:
  next create-version v1.4.0`,
	Args: cobra.ExactArgs(1),
	RunE: runCreateVersion,
}

func init() {
	createVersionCmd.Flags().BoolVarP(&forceVersion, "force", "f", false, "Forzar creación aunque haya cambios sin commit")
}

func runCreateVersion(cmd *cobra.Command, args []string) error {
	tag := args[0]

	// Validar formato semver
	if !isValidSemver(tag) {
		color.Red("✗ Formato de versión inválido: %s", tag)
		color.Yellow("  Use el formato: vX.Y.Z (ejemplo: v1.0.0)")
		return fmt.Errorf("formato de versión inválido")
	}

	// Verificar que estamos en un repo git
	repoRoot, err := git.GetRepoRoot()
	if err != nil {
		color.Red("✗ No se encuentra en un repositorio Git")
		return err
	}

	// Verificar cambios sin commit
	if !forceVersion {
		hasChanges, err := git.HasUncommittedChanges()
		if err != nil {
			color.Red("✗ Error al verificar estado del repositorio: %v", err)
			return err
		}
		if hasChanges {
			color.Red("✗ Existen cambios sin commit")
			color.Yellow("  Use -f para forzar la creación del tag")
			return fmt.Errorf("cambios sin commit")
		}
	}

	// Obtener remote origin
	remoteURL, err := git.GetRemoteURL("origin")
	if err != nil {
		color.Red("✗ Error al obtener remote origin: %v", err)
		return err
	}

	// Detectar proveedor y dominio desde la URL
	provider, domain, repoPath, err := git.ParseRemoteURL(remoteURL)
	if err != nil {
		color.Red("✗ Error al parsear URL del remote: %v", err)
		return err
	}

	// Cargar configuración y buscar cuenta
	cfg, err := config.Load()
	if err != nil {
		color.Red("✗ Error al cargar configuración: %v", err)
		return err
	}

	account, err := cfg.GetAccountByDomain(domain)
	if err != nil {
		color.Red("✗ No se encontró cuenta para el dominio: %s", domain)
		color.Yellow("  Use 'next login' para agregar una cuenta")
		return err
	}

	// Crear cliente del proveedor
	apiProvider, err := api.NewProvider(provider, account.Domain, account.Token)
	if err != nil {
		color.Red("✗ Error al crear cliente: %v", err)
		return err
	}

	// Crear tag en el remote
	if err := apiProvider.CreateTag(repoPath, tag); err != nil {
		color.Red("✗ Error al crear tag: %v", err)
		return err
	}

	// Mostrar éxito
	color.Green("✔ Versión %s creada exitosamente", tag)
	color.Cyan("Repositorio: %s", repoPath)

	_ = repoRoot // Usamos la variable para evitar warning

	return nil
}

// isValidSemver valida que el tag siga el formato vX.Y.Z
func isValidSemver(tag string) bool {
	pattern := `^v\d+\.\d+\.\d+$`
	matched, _ := regexp.MatchString(pattern, tag)
	return matched
}
