package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// GitHubProvider implementa Provider para GitHub
type GitHubProvider struct {
	baseURL string
	apiURL  string
	token   string
	client  *http.Client
}

// NewGitHubProvider crea un nuevo proveedor GitHub
func NewGitHubProvider(baseURL, token string) *GitHubProvider {
	// Determinar API URL
	apiURL := "https://api.github.com"
	if baseURL != "" && baseURL != "https://github.com" {
		// GitHub Enterprise
		baseURL = strings.TrimSuffix(baseURL, "/")
		apiURL = baseURL + "/api/v3"
	} else {
		baseURL = "https://github.com"
	}

	return &GitHubProvider{
		baseURL: baseURL,
		apiURL:  apiURL,
		token:   token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetAPIURL retorna la URL de la API
func (g *GitHubProvider) GetAPIURL() string {
	return g.apiURL
}

// ValidateToken valida el token y retorna el nombre de usuario
func (g *GitHubProvider) ValidateToken() (string, error) {
	req, err := http.NewRequest("GET", g.apiURL+"/user", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error de conexión: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token inválido o sin permisos (status: %d)", resp.StatusCode)
	}

	var user struct {
		Login string `json:"login"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("error al decodificar respuesta: %w", err)
	}

	return user.Login, nil
}

// ListGoLibraries lista todas las librerías Go del usuario (público y privado)
func (g *GitHubProvider) ListGoLibraries() ([]Library, error) {
	return g.ListGoLibrariesWithOptions(ListOptions{Visibility: VisibilityAll})
}

// ListGoLibrariesWithOptions lista librerías con opciones de filtrado
func (g *GitHubProvider) ListGoLibrariesWithOptions(opts ListOptions) ([]Library, error) {
	var libraries []Library
	page := 1
	perPage := 100

	for {
		var apiURL string

		// Si hay owner, buscar repos de ese usuario/organización
		if opts.Owner != "" {
			// Primero intentar como organización
			apiURL = fmt.Sprintf("%s/orgs/%s/repos?per_page=%d&page=%d", g.apiURL, opts.Owner, perPage, page)

			// Agregar filtro de visibilidad
			switch opts.Visibility {
			case VisibilityPublic:
				apiURL += "&type=public"
			case VisibilityPrivate:
				apiURL += "&type=private"
			}
		} else {
			// Repos del usuario autenticado
			apiURL = fmt.Sprintf("%s/user/repos?per_page=%d&page=%d", g.apiURL, perPage, page)

			// Agregar filtro de visibilidad
			switch opts.Visibility {
			case VisibilityPublic:
				apiURL += "&visibility=public"
			case VisibilityPrivate:
				apiURL += "&visibility=private"
			}
		}

		req, err := http.NewRequest("GET", apiURL, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Authorization", "Bearer "+g.token)
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := g.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error de conexión: %w", err)
		}

		// Si falla como org y hay owner, intentar como usuario
		if resp.StatusCode == http.StatusNotFound && opts.Owner != "" {
			resp.Body.Close()
			apiURL = fmt.Sprintf("%s/users/%s/repos?per_page=%d&page=%d", g.apiURL, opts.Owner, perPage, page)
			switch opts.Visibility {
			case VisibilityPublic:
				apiURL += "&type=public"
			case VisibilityPrivate:
				apiURL += "&type=private"
			}

			req, err = http.NewRequest("GET", apiURL, nil)
			if err != nil {
				return nil, err
			}

			req.Header.Set("Authorization", "Bearer "+g.token)
			req.Header.Set("Accept", "application/vnd.github+json")

			resp, err = g.client.Do(req)
			if err != nil {
				return nil, fmt.Errorf("error de conexión: %w", err)
			}
		}

		var repos []struct {
			Name        string `json:"name"`
			FullName    string `json:"full_name"`
			Description string `json:"description"`
			HTMLURL     string `json:"html_url"`
			Private     bool   `json:"private"`
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err := json.Unmarshal(body, &repos); err != nil {
			return nil, fmt.Errorf("error al decodificar respuesta: %w", err)
		}

		if len(repos) == 0 {
			break
		}

		for _, r := range repos {
			// Verificar si tiene go.mod
			if g.hasGoMod(r.FullName) {
				visibility := "public"
				if r.Private {
					visibility = "private"
				}

				libraries = append(libraries, Library{
					Name:        r.Name,
					Description: r.Description,
					URL:         r.HTMLURL,
					Provider:    "github",
					Visibility:  visibility,
				})
			}
		}

		page++
	}

	return libraries, nil
}

// hasGoMod verifica si un repositorio tiene archivo go.mod
func (g *GitHubProvider) hasGoMod(fullName string) bool {
	url := fmt.Sprintf("%s/repos/%s/contents/go.mod", g.apiURL, fullName)

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// ListVersions lista todas las versiones de una librería
func (g *GitHubProvider) ListVersions(library string) ([]Version, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/tags", g.apiURL, library)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github+json")

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
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, fmt.Errorf("error al decodificar respuesta: %w", err)
	}

	var versions []Version
	for _, t := range tags {
		// Obtener fecha del commit
		date := g.getCommitDate(library, t.Commit.SHA)
		versions = append(versions, Version{
			Name: t.Name,
			Date: date,
		})
	}

	return versions, nil
}

// getCommitDate obtiene la fecha de un commit
func (g *GitHubProvider) getCommitDate(repo, sha string) string {
	url := fmt.Sprintf("%s/repos/%s/commits/%s", g.apiURL, repo, sha)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var commit struct {
		Commit struct {
			Author struct {
				Date time.Time `json:"date"`
			} `json:"author"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return ""
	}

	return commit.Commit.Author.Date.Format("2006-01-02")
}

// CreateTag crea un tag en un repositorio
func (g *GitHubProvider) CreateTag(repoPath, tag string) error {
	// Obtener el SHA del HEAD
	sha, err := g.getDefaultBranchSHA(repoPath)
	if err != nil {
		return err
	}

	// Crear la referencia del tag
	apiURL := fmt.Sprintf("%s/repos/%s/git/refs", g.apiURL, repoPath)

	payload := map[string]string{
		"ref": "refs/tags/" + tag,
		"sha": sha,
	}

	body, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("error de conexión: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("error al crear tag (status: %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// getDefaultBranchSHA obtiene el SHA de la rama por defecto
func (g *GitHubProvider) getDefaultBranchSHA(repoPath string) (string, error) {
	apiURL := fmt.Sprintf("%s/repos/%s", g.apiURL, repoPath)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error de conexión: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error al obtener repositorio (status: %d)", resp.StatusCode)
	}

	var repo struct {
		DefaultBranch string `json:"default_branch"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return "", fmt.Errorf("error al decodificar respuesta: %w", err)
	}

	// Obtener SHA de la rama por defecto
	branchURL := fmt.Sprintf("%s/repos/%s/branches/%s", g.apiURL, repoPath, repo.DefaultBranch)

	req, err = http.NewRequest("GET", branchURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err = g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error de conexión: %w", err)
	}
	defer resp.Body.Close()

	var branch struct {
		Commit struct {
			SHA string `json:"sha"`
		} `json:"commit"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&branch); err != nil {
		return "", fmt.Errorf("error al decodificar respuesta: %w", err)
	}

	return branch.Commit.SHA, nil
}
