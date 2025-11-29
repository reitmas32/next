package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GitLabProvider implementa Provider para GitLab
type GitLabProvider struct {
	baseURL string
	apiURL  string
	token   string
	client  *http.Client
}

// NewGitLabProvider crea un nuevo proveedor GitLab
func NewGitLabProvider(baseURL, token string) *GitLabProvider {
	// Normalizar URL
	baseURL = strings.TrimSuffix(baseURL, "/")
	apiURL := baseURL + "/api/v4"

	return &GitLabProvider{
		baseURL: baseURL,
		apiURL:  apiURL,
		token:   token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAPIURL retorna la URL de la API
func (g *GitLabProvider) GetAPIURL() string {
	return g.apiURL
}

// ValidateToken valida el token y retorna el nombre de usuario
func (g *GitLabProvider) ValidateToken() (string, error) {
	req, err := http.NewRequest("GET", g.apiURL+"/user", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("PRIVATE-TOKEN", g.token)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error de conexión: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token inválido o sin permisos (status: %d)", resp.StatusCode)
	}

	var user struct {
		Username string `json:"username"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("error al decodificar respuesta: %w", err)
	}

	return user.Username, nil
}

// ListGoLibraries lista todas las librerías Go del usuario (público y privado)
func (g *GitLabProvider) ListGoLibraries() ([]Library, error) {
	return g.ListGoLibrariesWithOptions(ListOptions{Visibility: VisibilityAll})
}

// ListGoLibrariesWithOptions lista librerías con opciones de filtrado
func (g *GitLabProvider) ListGoLibrariesWithOptions(opts ListOptions) ([]Library, error) {
	var libraries []Library
	page := 1
	perPage := 100

	for {
		// Construir URL con parámetros
		apiURL := fmt.Sprintf("%s/projects?per_page=%d&page=%d", g.apiURL, perPage, page)

		// Filtrar por visibilidad
		switch opts.Visibility {
		case VisibilityPublic:
			apiURL += "&visibility=public"
		case VisibilityPrivate:
			apiURL += "&visibility=private"
		}

		// Si hay owner, buscar por grupo/usuario específico
		if opts.Owner != "" {
			// Buscar proyectos de un grupo o usuario específico
			encodedOwner := url.PathEscape(opts.Owner)
			apiURL = fmt.Sprintf("%s/groups/%s/projects?per_page=%d&page=%d", g.apiURL, encodedOwner, perPage, page)
			if opts.Visibility == VisibilityPublic {
				apiURL += "&visibility=public"
			} else if opts.Visibility == VisibilityPrivate {
				apiURL += "&visibility=private"
			}
		} else {
			// Para proyectos propios, incluir membership
			apiURL += "&membership=true"
		}

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("PRIVATE-TOKEN", g.token)

		resp, err := g.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error de conexión: %w", err)
		}

		var projects []struct {
			ID          int    `json:"id"`
			Name        string `json:"name"`
			Description string `json:"description"`
			WebURL      string `json:"web_url"`
			PathWithNS  string `json:"path_with_namespace"`
			Visibility  string `json:"visibility"` // "public", "internal", "private"
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err := json.Unmarshal(body, &projects); err != nil {
			return nil, fmt.Errorf("error al decodificar respuesta: %w", err)
		}

		if len(projects) == 0 {
			break
		}

		for _, p := range projects {
			// Verificar si tiene go.mod
			if g.hasGoMod(p.ID) {
				visibility := p.Visibility
				if visibility == "internal" {
					visibility = "private"
				}

				libraries = append(libraries, Library{
					Name:        p.Name,
					Description: p.Description,
					URL:         p.WebURL,
					Provider:    "gitlab",
					Visibility:  visibility,
				})
			}
		}

		page++
	}

	return libraries, nil
}

// hasGoMod verifica si un proyecto tiene archivo go.mod
func (g *GitLabProvider) hasGoMod(projectID int) bool {
	url := fmt.Sprintf("%s/projects/%d/repository/files/go.mod?ref=main", g.apiURL, projectID)

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("PRIVATE-TOKEN", g.token)

	resp, err := g.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true
	}

	// Intentar con master si main no funciona
	url = fmt.Sprintf("%s/projects/%d/repository/files/go.mod?ref=master", g.apiURL, projectID)

	req, err = http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("PRIVATE-TOKEN", g.token)

	resp, err = g.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// ListVersions lista todas las versiones de una librería
func (g *GitLabProvider) ListVersions(library string) ([]Version, error) {
	// Codificar el path del proyecto
	encodedPath := url.PathEscape(library)
	apiURL := fmt.Sprintf("%s/projects/%s/repository/tags", g.apiURL, encodedPath)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("PRIVATE-TOKEN", g.token)

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error de conexión: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error al obtener tags (status: %d)", resp.StatusCode)
	}

	var tags []struct {
		Name   string `json:"name"`
		Commit struct {
			CreatedAt time.Time `json:"created_at"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %w", err)
	}

	var versions []Version
	for _, t := range tags {
		versions = append(versions, Version{
			Name: t.Name,
			Date: t.Commit.CreatedAt.Format("2006-01-02"),
		})
	}

	return versions, nil
}

// CreateTag crea un tag en un repositorio
func (g *GitLabProvider) CreateTag(repoPath, tag string) error {
	// Obtener el commit actual (HEAD)
	ref, err := g.getDefaultBranchRef(repoPath)
	if err != nil {
		return err
	}

	// Codificar el path del proyecto
	encodedPath := url.PathEscape(repoPath)
	apiURL := fmt.Sprintf("%s/projects/%s/repository/tags", g.apiURL, encodedPath)

	data := url.Values{}
	data.Set("tag_name", tag)
	data.Set("ref", ref)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("PRIVATE-TOKEN", g.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("error de conexión: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error al crear tag (status: %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// getDefaultBranchRef obtiene la referencia de la rama por defecto
func (g *GitLabProvider) getDefaultBranchRef(repoPath string) (string, error) {
	encodedPath := url.PathEscape(repoPath)
	apiURL := fmt.Sprintf("%s/projects/%s", g.apiURL, encodedPath)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("PRIVATE-TOKEN", g.token)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error de conexión: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error al obtener proyecto (status: %d)", resp.StatusCode)
	}

	var project struct {
		DefaultBranch string `json:"default_branch"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&project); err != nil {
		return "", fmt.Errorf("error al decodificar respuesta: %w", err)
	}

	return project.DefaultBranch, nil
}
