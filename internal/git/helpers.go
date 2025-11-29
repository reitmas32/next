package git

import (
	"fmt"
	"regexp"
	"strings"
)

// ParseRemoteURL parsea una URL de remote y extrae proveedor, dominio y path del repo
func ParseRemoteURL(remoteURL string) (provider, domain, repoPath string, err error) {
	// Patrones para diferentes formatos de URL
	// SSH: git@github.com:owner/repo.git
	// HTTPS: https://github.com/owner/repo.git

	// Limpiar .git al final
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	// Patrón SSH
	sshPattern := regexp.MustCompile(`^git@([^:]+):(.+)$`)
	if matches := sshPattern.FindStringSubmatch(remoteURL); len(matches) == 3 {
		domain = matches[1]
		repoPath = matches[2]
		provider = detectProvider(domain)
		return provider, domain, repoPath, nil
	}

	// Patrón HTTPS
	httpsPattern := regexp.MustCompile(`^https?://([^/]+)/(.+)$`)
	if matches := httpsPattern.FindStringSubmatch(remoteURL); len(matches) == 3 {
		domain = matches[1]
		repoPath = matches[2]
		provider = detectProvider(domain)
		return provider, domain, repoPath, nil
	}

	return "", "", "", fmt.Errorf("formato de URL no reconocido: %s", remoteURL)
}

// detectProvider detecta el proveedor basándose en el dominio
func detectProvider(domain string) string {
	domain = strings.ToLower(domain)

	if strings.Contains(domain, "github") {
		return "github"
	}
	if strings.Contains(domain, "gitlab") {
		return "gitlab"
	}

	// Por defecto asumir GitLab para dominios personalizados
	return "gitlab"
}

// IsValidSemver valida que un tag siga el formato semántico vX.Y.Z
func IsValidSemver(tag string) bool {
	pattern := regexp.MustCompile(`^v\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?$`)
	return pattern.MatchString(tag)
}

// NormalizeRepoPath normaliza el path de un repositorio
func NormalizeRepoPath(path string) string {
	// Remover .git al final
	path = strings.TrimSuffix(path, ".git")
	// Remover barra inicial
	path = strings.TrimPrefix(path, "/")
	return path
}

