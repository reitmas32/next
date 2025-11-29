package api

import (
	"fmt"
)

// Library representa una librería Go
type Library struct {
	Name        string
	Description string
	URL         string
	Provider    string
}

// Version representa una versión/tag de una librería
type Version struct {
	Name string
	Date string
}

// Provider define la interfaz para interactuar con proveedores Git
type Provider interface {
	// ValidateToken valida el token y retorna el nombre de usuario
	ValidateToken() (string, error)

	// GetAPIURL retorna la URL de la API
	GetAPIURL() string

	// ListGoLibraries lista todas las librerías Go del usuario
	ListGoLibraries() ([]Library, error)

	// ListVersions lista todas las versiones de una librería
	ListVersions(library string) ([]Version, error)

	// CreateTag crea un tag en un repositorio
	CreateTag(repoPath, tag string) error
}

// NewProvider crea un nuevo proveedor según el tipo
func NewProvider(providerType, baseURL, token string) (Provider, error) {
	switch providerType {
	case "gitlab":
		return NewGitLabProvider(baseURL, token), nil
	case "github":
		return NewGitHubProvider(baseURL, token), nil
	default:
		return nil, fmt.Errorf("proveedor no soportado: %s", providerType)
	}
}

