package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	configDir  = ".next"
	configFile = "config.json"
	mu         sync.Mutex
)

// getConfigPath retorna la ruta completa del archivo de configuración
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("no se pudo obtener directorio home: %w", err)
	}

	return filepath.Join(homeDir, configDir, configFile), nil
}

// Load carga la configuración desde el archivo
func Load() (*Config, error) {
	mu.Lock()
	defer mu.Unlock()

	configPath, err := getConfigPath()
	if err != nil {
		return nil, err
	}

	// Si no existe el archivo, retornar config vacía
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return NewConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error al leer configuración: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("error al parsear configuración: %w", err)
	}

	return &cfg, nil
}

// Save guarda la configuración en el archivo
func (c *Config) Save() error {
	mu.Lock()
	defer mu.Unlock()

	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Crear directorio si no existe
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("error al crear directorio de configuración: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("error al serializar configuración: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("error al guardar configuración: %w", err)
	}

	return nil
}

// AddAccount agrega o actualiza una cuenta
func (c *Config) AddAccount(account Account) {
	// Buscar si ya existe
	for i, a := range c.Accounts {
		if a.Name == account.Name {
			c.Accounts[i] = account
			return
		}
	}

	// Agregar nueva
	c.Accounts = append(c.Accounts, account)
}

// GetAccount obtiene una cuenta por nombre
func (c *Config) GetAccount(name string) (*Account, error) {
	// Si no se especifica nombre y solo hay una cuenta, usarla
	if name == "" {
		if len(c.Accounts) == 0 {
			return nil, fmt.Errorf("no hay cuentas configuradas. Use 'next login' para agregar una")
		}
		if len(c.Accounts) == 1 {
			return &c.Accounts[0], nil
		}
		return nil, fmt.Errorf("hay múltiples cuentas configuradas. Use --account para especificar cuál usar")
	}

	// Buscar por nombre
	for _, a := range c.Accounts {
		if a.Name == name {
			return &a, nil
		}
	}

	return nil, fmt.Errorf("cuenta '%s' no encontrada", name)
}

// GetAccountByDomain obtiene una cuenta por dominio
func (c *Config) GetAccountByDomain(domain string) (*Account, error) {
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimSuffix(domain, "/")

	for _, a := range c.Accounts {
		accountDomain := strings.TrimPrefix(a.Domain, "https://")
		accountDomain = strings.TrimPrefix(accountDomain, "http://")
		accountDomain = strings.TrimSuffix(accountDomain, "/")

		if accountDomain == domain {
			return &a, nil
		}
	}

	return nil, fmt.Errorf("no se encontró cuenta para el dominio: %s", domain)
}

// RemoveAccount elimina una cuenta por nombre
func (c *Config) RemoveAccount(name string) error {
	for i, a := range c.Accounts {
		if a.Name == name {
			c.Accounts = append(c.Accounts[:i], c.Accounts[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("cuenta '%s' no encontrada", name)
}

// ListAccounts retorna todas las cuentas configuradas
func (c *Config) ListAccounts() []Account {
	return c.Accounts
}

