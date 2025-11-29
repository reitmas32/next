package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/reitmas32/next/internal/crypto"
)

var (
	configDir  = ".next"
	configFile = "config.json"
	mu         sync.Mutex

	// Cache de la llave de encriptación
	encryptionKey []byte
)

// getConfigPath retorna la ruta completa del archivo de configuración
func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("no se pudo obtener directorio home: %w", err)
	}

	return filepath.Join(homeDir, configDir, configFile), nil
}

// getEncryptionKey obtiene o crea la llave de encriptación
func getEncryptionKey() ([]byte, error) {
	if encryptionKey != nil {
		return encryptionKey, nil
	}

	key, err := crypto.GetOrCreateKey()
	if err != nil {
		return nil, err
	}

	encryptionKey = key
	return key, nil
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

	// Desencriptar tokens
	key, err := getEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo llave de encriptación: %w", err)
	}

	for i := range cfg.Accounts {
		if crypto.IsEncrypted(cfg.Accounts[i].Token) {
			decrypted, err := crypto.Decrypt(cfg.Accounts[i].Token, key)
			if err != nil {
				// Si falla la desencriptación, el token podría estar en texto plano (migración)
				continue
			}
			cfg.Accounts[i].Token = decrypted
		}
	}

	return &cfg, nil
}

// Save guarda la configuración en el archivo con tokens encriptados
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

	// Obtener llave de encriptación
	key, err := getEncryptionKey()
	if err != nil {
		return fmt.Errorf("error obteniendo llave de encriptación: %w", err)
	}

	// Crear copia para encriptar tokens
	configToSave := &Config{
		Accounts: make([]Account, len(c.Accounts)),
	}

	for i, acc := range c.Accounts {
		configToSave.Accounts[i] = acc

		// Solo encriptar si no está ya encriptado
		if !crypto.IsEncrypted(acc.Token) {
			encrypted, err := crypto.Encrypt(acc.Token, key)
			if err != nil {
				return fmt.Errorf("error encriptando token: %w", err)
			}
			configToSave.Accounts[i].Token = encrypted
		}
	}

	data, err := json.MarshalIndent(configToSave, "", "  ")
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

// GetAccountByDomain obtiene una cuenta por dominio (usa la primera que coincida)
// Para mejor precisión con múltiples cuentas del mismo dominio, usar GetAccountForModule
func (c *Config) GetAccountByDomain(domain string) (*Account, error) {
	domain = normalizeDomain(domain)

	// Buscar primera cuenta que coincida con el dominio
	for i := range c.Accounts {
		accountDomain := normalizeDomain(c.Accounts[i].Domain)
		if accountDomain == domain {
			return &c.Accounts[i], nil
		}
	}

	return nil, fmt.Errorf("no se encontró cuenta para el dominio: %s", domain)
}

// GetAccountForModule obtiene la cuenta correcta para un módulo Go específico
// Prioridad: 1) Cuenta con owner específico, 2) Cuenta wildcard (sin owners)
func (c *Config) GetAccountForModule(module string) (*Account, error) {
	domain, owner := parseModule(module)
	domain = normalizeDomain(domain)

	var wildcardAccount *Account

	// Buscar cuenta con owner específico primero
	for i := range c.Accounts {
		accountDomain := normalizeDomain(c.Accounts[i].Domain)

		if accountDomain != domain {
			continue
		}

		// Si tiene el owner específico, es la cuenta correcta
		if c.Accounts[i].HasOwner(owner) {
			return &c.Accounts[i], nil
		}

		// Guardar cuenta wildcard como fallback
		if c.Accounts[i].IsWildcard() && wildcardAccount == nil {
			wildcardAccount = &c.Accounts[i]
		}
	}

	// Si encontramos cuenta wildcard, usarla
	if wildcardAccount != nil {
		return wildcardAccount, nil
	}

	return nil, fmt.Errorf("no se encontró cuenta para el módulo: %s", module)
}

// GetAccountByDomainAndOwner obtiene la cuenta para un dominio y owner específico
func (c *Config) GetAccountByDomainAndOwner(domain, owner string) (*Account, error) {
	domain = normalizeDomain(domain)

	var wildcardAccount *Account

	for i := range c.Accounts {
		accountDomain := normalizeDomain(c.Accounts[i].Domain)

		if accountDomain != domain {
			continue
		}

		// Match exacto con owner
		if c.Accounts[i].HasOwner(owner) {
			return &c.Accounts[i], nil
		}

		// Guardar wildcard como fallback
		if c.Accounts[i].IsWildcard() && wildcardAccount == nil {
			wildcardAccount = &c.Accounts[i]
		}
	}

	if wildcardAccount != nil {
		return wildcardAccount, nil
	}

	return nil, fmt.Errorf("no se encontró cuenta para %s/%s", domain, owner)
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

// normalizeDomain normaliza un dominio eliminando protocolo y trailing slash
func normalizeDomain(domain string) string {
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimSuffix(domain, "/")
	return domain
}

// parseModule extrae el dominio y owner de un módulo Go
// Ejemplo: "github.com/reitmas32/mathutils" -> "github.com", "reitmas32"
func parseModule(module string) (domain, owner string) {
	parts := strings.Split(module, "/")
	if len(parts) >= 1 {
		domain = parts[0]
	}
	if len(parts) >= 2 {
		owner = parts[1]
	}
	return
}
