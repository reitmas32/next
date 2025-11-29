package git

import (
	"fmt"
	"os/exec"
	"strconv"
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

// FetchRemote hace fetch del remote para tener la información actualizada
func FetchRemote(remote string) error {
	cmd := exec.Command("git", "fetch", remote)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error al hacer fetch de '%s': %w", remote, err)
	}
	return nil
}

// GetCommitsAhead retorna la cantidad de commits que la rama local está adelante del remote
func GetCommitsAhead(remote, branch string) (int, error) {
	// Formato: origin/main
	remoteBranch := fmt.Sprintf("%s/%s", remote, branch)

	// Verificar si existe la rama remota
	checkCmd := exec.Command("git", "rev-parse", "--verify", remoteBranch)
	if err := checkCmd.Run(); err != nil {
		// La rama remota no existe, todos los commits son nuevos
		cmd := exec.Command("git", "rev-list", "--count", "HEAD")
		output, err := cmd.Output()
		if err != nil {
			return 0, fmt.Errorf("error al contar commits: %w", err)
		}
		count, _ := strconv.Atoi(strings.TrimSpace(string(output)))
		return count, nil
	}

	// Contar commits adelante
	cmd := exec.Command("git", "rev-list", "--count", fmt.Sprintf("%s..HEAD", remoteBranch))
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("error al contar commits adelante: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, fmt.Errorf("error al parsear conteo: %w", err)
	}

	return count, nil
}

// GetCommitsBehind retorna la cantidad de commits que la rama local está detrás del remote
func GetCommitsBehind(remote, branch string) (int, error) {
	remoteBranch := fmt.Sprintf("%s/%s", remote, branch)

	// Verificar si existe la rama remota
	checkCmd := exec.Command("git", "rev-parse", "--verify", remoteBranch)
	if err := checkCmd.Run(); err != nil {
		// La rama remota no existe
		return 0, nil
	}

	cmd := exec.Command("git", "rev-list", "--count", fmt.Sprintf("HEAD..%s", remoteBranch))
	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("error al contar commits detrás: %w", err)
	}

	count, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, fmt.Errorf("error al parsear conteo: %w", err)
	}

	return count, nil
}

// PushBranch hace push de la rama actual al remote
func PushBranch(remote, branch string) error {
	cmd := exec.Command("git", "push", remote, branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error al hacer push: %s", string(output))
	}
	return nil
}

// PushBranchSetUpstream hace push y configura el upstream
func PushBranchSetUpstream(remote, branch string) error {
	cmd := exec.Command("git", "push", "-u", remote, branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error al hacer push: %s", string(output))
	}
	return nil
}

// HasRemoteBranch verifica si existe la rama en el remote
func HasRemoteBranch(remote, branch string) bool {
	remoteBranch := fmt.Sprintf("%s/%s", remote, branch)
	cmd := exec.Command("git", "rev-parse", "--verify", remoteBranch)
	return cmd.Run() == nil
}

// BranchStatus representa el estado de sincronización de una rama
type BranchStatus struct {
	Branch       string
	Remote       string
	Ahead        int
	Behind       int
	IsNew        bool // true si la rama no existe en el remote
	NeedsPush    bool
	NeedsPull    bool
	IsSynced     bool
}

// GetBranchStatus obtiene el estado completo de sincronización
func GetBranchStatus(remote string) (*BranchStatus, error) {
	branch, err := GetCurrentBranch()
	if err != nil {
		return nil, err
	}

	status := &BranchStatus{
		Branch: branch,
		Remote: remote,
	}

	// Hacer fetch para tener info actualizada
	_ = FetchRemote(remote)

	// Verificar si la rama existe en el remote
	if !HasRemoteBranch(remote, branch) {
		status.IsNew = true
		status.NeedsPush = true

		// Contar commits locales
		cmd := exec.Command("git", "rev-list", "--count", "HEAD")
		output, err := cmd.Output()
		if err == nil {
			status.Ahead, _ = strconv.Atoi(strings.TrimSpace(string(output)))
		}
		return status, nil
	}

	// Obtener commits adelante y detrás
	ahead, err := GetCommitsAhead(remote, branch)
	if err != nil {
		return nil, err
	}
	status.Ahead = ahead

	behind, err := GetCommitsBehind(remote, branch)
	if err != nil {
		return nil, err
	}
	status.Behind = behind

	status.NeedsPush = ahead > 0
	status.NeedsPull = behind > 0
	status.IsSynced = ahead == 0 && behind == 0

	return status, nil
}
