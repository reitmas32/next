package config

// Account representa una cuenta configurada
type Account struct {
	Name     string   `json:"name"`
	Provider string   `json:"provider"`
	APIURL   string   `json:"api_url"`
	Domain   string   `json:"domain"`
	Token    string   `json:"token"`
	Owners   []string `json:"owners,omitempty"` // Usuarios/orgs que maneja esta cuenta (opcional)
}

// Config representa la configuración completa del CLI
type Config struct {
	Accounts []Account `json:"accounts"`
}

// NewConfig crea una nueva configuración vacía
func NewConfig() *Config {
	return &Config{
		Accounts: []Account{},
	}
}

// HasOwner verifica si la cuenta tiene un owner específico configurado
func (a *Account) HasOwner(owner string) bool {
	for _, o := range a.Owners {
		if o == owner {
			return true
		}
	}
	return false
}

// IsWildcard retorna true si la cuenta no tiene owners (acepta todos)
func (a *Account) IsWildcard() bool {
	return len(a.Owners) == 0
}
