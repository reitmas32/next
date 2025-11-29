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

var (
	forceVersion bool
	skipPush     bool
)

var createVersionCmd = &cobra.Command{
	Use:   "create-version <tag>",
	Short: "Crea un tag Git semántico en el repositorio actual",
	Long: `Crea un tag Git semántico en el repositorio actual.

El tag debe seguir el formato semántico: vX.Y.Z

Validaciones:
  - El directorio actual debe ser un repositorio Git válido
  - No deben existir cambios sin commit (usar -f para forzar)
  - El tag debe seguir el formato vX.Y.Z
  - Si hay commits pendientes de push, los sube automáticamente

Ejemplo:
  next create-version v1.4.0`,
	Args: cobra.ExactArgs(1),
	RunE: runCreateVersion,
}

func init() {
	createVersionCmd.Flags().BoolVarP(&forceVersion, "force", "f", false, "Forzar creación aunque haya cambios sin commit")
	createVersionCmd.Flags().BoolVar(&skipPush, "skip-push", false, "No hacer push automático de commits pendientes")
}

func runCreateVersion(cmd *cobra.Command, args []string) error {
	tag := args[0]

	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)

	// Validar formato semver
	if !isValidSemver(tag) {
		color.Red("✗ Formato de versión inválido: %s", tag)
		color.Yellow("  Use el formato: vX.Y.Z (ejemplo: v1.0.0)")
		return fmt.Errorf("formato de versión inválido")
	}

	// Verificar que estamos en un repo git
	_, err := git.GetRepoRoot()
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
			yellow.Println("  Haga commit de sus cambios o use -f para forzar")
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

	// Verificar estado de sincronización con el remote
	cyan.Println("🔍 Verificando sincronización con origin...")

	status, err := git.GetBranchStatus("origin")
	if err != nil {
		color.Red("✗ Error al verificar estado de la rama: %v", err)
		return err
	}

	// Si hay commits detrás, advertir
	if status.NeedsPull {
		yellow.Printf("! La rama '%s' está %d commit(s) detrás de origin\n", status.Branch, status.Behind)
		yellow.Println("  Considere hacer 'git pull' antes de crear la versión")
		if !forceVersion {
			return fmt.Errorf("rama desactualizada")
		}
		yellow.Println("  Continuando por -f (force)...")
	}

	// Si hay commits pendientes de push, subirlos
	if status.NeedsPush && !skipPush {
		if status.IsNew {
			cyan.Printf("📤 La rama '%s' es nueva, subiendo al remote...\n", status.Branch)
		} else {
			cyan.Printf("📤 Subiendo %d commit(s) pendiente(s) a origin/%s...\n", status.Ahead, status.Branch)
		}

		var pushErr error
		if status.IsNew {
			pushErr = git.PushBranchSetUpstream("origin", status.Branch)
		} else {
			pushErr = git.PushBranch("origin", status.Branch)
		}

		if pushErr != nil {
			color.Red("✗ Error al hacer push: %v", pushErr)
			return pushErr
		}
		green.Printf("✔ Código subido exitosamente\n")
	} else if status.IsSynced {
		green.Printf("✔ Rama '%s' sincronizada con origin\n", status.Branch)
	}

	// Crear cliente del proveedor
	apiProvider, err := api.NewProvider(provider, account.Domain, account.Token)
	if err != nil {
		color.Red("✗ Error al crear cliente: %v", err)
		return err
	}

	// Crear tag en el remote
	cyan.Printf("🏷️  Creando tag %s...\n", tag)

	if err := apiProvider.CreateTag(repoPath, tag); err != nil {
		color.Red("✗ Error al crear tag: %v", err)
		return err
	}

	// Mostrar éxito
	fmt.Println()
	green.Printf("✔ Versión %s creada exitosamente\n", tag)
	cyan.Printf("  Repositorio: %s\n", repoPath)
	cyan.Printf("  Rama: %s\n", status.Branch)
	fmt.Println()

	// Mostrar cómo instalar
	color.White("Para instalar esta versión:")
	cyan.Printf("  go get %s/%s@%s\n", domain, repoPath, tag)
	fmt.Println()

	return nil
}

// isValidSemver valida que el tag siga el formato vX.Y.Z
func isValidSemver(tag string) bool {
	pattern := `^v\d+\.\d+\.\d+$`
	matched, _ := regexp.MatchString(pattern, tag)
	return matched
}
