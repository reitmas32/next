package config

// Account representa una cuenta configurada
type Account struct {
	Name     string `json:"name"`
	Provider string `json:"provider"`
	APIURL   string `json:"api_url"`
	Domain   string `json:"domain"`
	Token    string `json:"token"`
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
