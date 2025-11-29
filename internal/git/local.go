package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// GetRepoRoot obtiene el directorio raíz del repositorio Git
func GetRepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("no es un repositorio Git: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// HasUncommittedChanges verifica si hay cambios sin commit
func HasUncommittedChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("error al verificar estado: %w", err)
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
}

// GetRemoteURL obtiene la URL de un remote
func GetRemoteURL(remote string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", remote)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error al obtener remote '%s': %w", remote, err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCurrentBranch obtiene la rama actual
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error al obtener rama actual: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetCurrentCommit obtiene el hash del commit actual
func GetCurrentCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error al obtener commit actual: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

