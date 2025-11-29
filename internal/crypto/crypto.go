package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"runtime"

	"github.com/zalando/go-keyring"
)

const (
	serviceName = "next-cli"
	keyName     = "encryption-key"
	keySize     = 32 // AES-256
)

// GetOrCreateKey obtiene la llave de encriptación del keychain o la crea
func GetOrCreateKey() ([]byte, error) {
	// Intentar obtener del keychain
	keyStr, err := keyring.Get(serviceName, keyName)
	if err == nil {
		return base64.StdEncoding.DecodeString(keyStr)
	}

	// No existe, crear nueva llave
	key := make([]byte, keySize)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("error generando llave: %w", err)
	}

	// Intentar guardar en keychain
	keyStr = base64.StdEncoding.EncodeToString(key)
	if err := keyring.Set(serviceName, keyName, keyStr); err != nil {
		// Fallback: derivar llave del sistema
		return deriveKeyFromSystem()
	}

	return key, nil
}

// deriveKeyFromSystem deriva una llave del machine-id como fallback
func deriveKeyFromSystem() ([]byte, error) {
	var machineID string

	// Intentar leer machine-id según el OS
	switch runtime.GOOS {
	case "linux":
		data, err := os.ReadFile("/etc/machine-id")
		if err != nil {
			data, err = os.ReadFile("/var/lib/dbus/machine-id")
		}
		if err == nil {
			machineID = string(data)
		}
	case "darwin":
		// macOS: usar IOPlatformUUID
		machineID = "macos-fallback-key"
	case "windows":
		machineID = "windows-fallback-key"
	}

	if machineID == "" {
		machineID = "next-cli-default-key"
	}

	// Agregar salt y usuario
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("USERNAME")
	}

	combined := fmt.Sprintf("%s:%s:next-cli-v1", machineID, user)

	// Derivar llave usando SHA-256
	hash := sha256.Sum256([]byte(combined))
	return hash[:], nil
}

// Encrypt encripta un texto usando AES-256-GCM
func Encrypt(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt desencripta un texto encriptado con AES-256-GCM
func Decrypt(encryptedText string, key []byte) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext demasiado corto")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// IsEncrypted verifica si un string parece estar encriptado (base64 válido)
func IsEncrypted(s string) bool {
	// Los tokens de GitHub empiezan con "ghp_" o "gho_"
	// Los tokens de GitLab empiezan con "glpat-"
	// Si empieza con estos prefijos, NO está encriptado
	if len(s) > 4 {
		prefix := s[:4]
		if prefix == "ghp_" || prefix == "gho_" || prefix == "glpa" {
			return false
		}
	}

	// Intentar decodificar como base64
	_, err := base64.StdEncoding.DecodeString(s)
	return err == nil && len(s) > 20
}

// DeleteKey elimina la llave del keychain
func DeleteKey() error {
	return keyring.Delete(serviceName, keyName)
}
