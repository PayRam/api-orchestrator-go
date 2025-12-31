package models

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"sync"
	"time"
)

var (
	// encryptionKey is set by the library consumer during initialization
	encryptionKey   []byte
	encryptionMutex sync.RWMutex

	// ErrEncryptionKeyNotSet is returned when encryption is attempted without setting a key
	ErrEncryptionKeyNotSet = errors.New("encryption key not set: call SetEncryptionKey() before using credential encryption")
)

// SetEncryptionKey sets the encryption key for the library.
// This MUST be called by the library consumer before using credential encryption.
// The key must be exactly 32 bytes for AES-256 encryption.
//
// Example usage:
//
//	func main() {
//	    key := []byte(os.Getenv("MY_APP_ENCRYPTION_KEY"))
//	    if err := models.SetEncryptionKey(key); err != nil {
//	        log.Fatal(err)
//	    }
//	}
func SetEncryptionKey(key []byte) error {
	if len(key) == 0 {
		return ErrEncryptionKeyNotSet
	}

	encryptionMutex.Lock()
	defer encryptionMutex.Unlock()

	// Ensure key is exactly 32 bytes for AES-256
	if len(key) < 32 {
		// Pad if shorter
		padded := make([]byte, 32)
		copy(padded, key)
		encryptionKey = padded
	} else {
		encryptionKey = key[:32]
	}

	return nil
}

// SetEncryptionKeyFromString is a convenience method that accepts a string key.
// The string should be at least 32 characters for full AES-256 security.
func SetEncryptionKeyFromString(key string) error {
	return SetEncryptionKey([]byte(key))
}

// IsEncryptionKeySet returns true if an encryption key has been configured.
func IsEncryptionKeySet() bool {
	encryptionMutex.RLock()
	defer encryptionMutex.RUnlock()
	return len(encryptionKey) == 32
}

// ClearEncryptionKey removes the encryption key from memory.
// Useful for testing or security-sensitive cleanup.
func ClearEncryptionKey() {
	encryptionMutex.Lock()
	defer encryptionMutex.Unlock()
	// Zero out the key before clearing
	for i := range encryptionKey {
		encryptionKey[i] = 0
	}
	encryptionKey = nil
}

// getEncryptionKey retrieves the configured encryption key.
// Returns an error if no key has been set.
func getEncryptionKey() ([]byte, error) {
	encryptionMutex.RLock()
	defer encryptionMutex.RUnlock()

	if len(encryptionKey) != 32 {
		return nil, ErrEncryptionKeyNotSet
	}

	// Return a copy to prevent external modification
	keyCopy := make([]byte, 32)
	copy(keyCopy, encryptionKey)
	return keyCopy, nil
}

// Credential stores API credentials per provider securely.
// This model supports field-level encryption for sensitive data like API keys and secrets.
type Credential struct {
	// ID is the unique identifier for the credential (UUID format).
	ID string `gorm:"type:uuid;primaryKey" json:"id" db:"id"`

	// ProviderID is the foreign key reference to the Provider.
	ProviderID string `gorm:"type:uuid;not null;index:idx_credential_provider_id" json:"provider_id" db:"provider_id"`

	// Provider is the associated Provider entity (foreign key relationship).
	Provider Provider `gorm:"foreignKey:ProviderID;constraint:OnDelete:CASCADE" json:"-"`

	// Key is the credential key name (e.g., "API_KEY", "SECRET", "API_SECRET").
	Key string `gorm:"type:varchar(255);not null;index:idx_credential_key" json:"key" db:"key"`

	// Value stores the credential value (encrypted).
	// Use EncryptValue/DecryptValue helper methods for secure access.
	Value string `gorm:"type:text;not null" json:"value" db:"value"`

	// Required indicates whether this credential is mandatory for the provider.
	Required bool `gorm:"default:false" json:"required" db:"required"`

	// Description provides additional context about the credential.
	Description string `gorm:"type:text" json:"description" db:"description"`

	// CreatedAt is the timestamp when the credential was created.
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at" db:"created_at"`

	// UpdatedAt is the timestamp when the credential was last updated.
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at" db:"updated_at"`
}

// TableName overrides the table name for GORM.
// Returns the table name with configured prefix.
func (Credential) TableName() string {
	return GetTableName("credentials")
}

// EncryptValue encrypts the given plain text value using AES-256-GCM.
// Returns base64-encoded ciphertext that includes the nonce.
func EncryptValue(plainText string) (string, error) {
	if plainText == "" {
		return "", nil
	}

	key, err := getEncryptionKey()
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

	// Create a unique nonce for this encryption
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and prepend nonce to ciphertext
	cipherText := gcm.Seal(nonce, nonce, []byte(plainText), nil)

	// Encode to base64 for storage
	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// DecryptValue decrypts the stored encrypted value using AES-256-GCM.
// Expects base64-encoded ciphertext with prepended nonce.
func DecryptValue(cipherText string) (string, error) {
	if cipherText == "" {
		return "", nil
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		// If it's not base64, assume it's plain text (for backward compatibility)
		return cipherText, nil
	}

	key, err := getEncryptionKey()
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
	if len(data) < nonceSize {
		// Data too short, assume it's plain text (backward compatibility)
		return cipherText, nil
	}

	// Extract nonce and ciphertext
	nonce, cipherBytes := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plainText, err := gcm.Open(nil, nonce, cipherBytes, nil)
	if err != nil {
		// Decryption failed, might be plain text (backward compatibility)
		return cipherText, nil
	}

	return string(plainText), nil
}

// SetEncryptedValue sets the Value field after encrypting the plain text.
// Use this method when storing new credentials.
func (c *Credential) SetEncryptedValue(plainText string) error {
	if plainText == "" {
		return errors.New("value cannot be empty")
	}
	encrypted, err := EncryptValue(plainText)
	if err != nil {
		return err
	}
	c.Value = encrypted
	return nil
}

// GetDecryptedValue returns the decrypted value of the credential.
// Use this method when retrieving credentials for use.
func (c *Credential) GetDecryptedValue() (string, error) {
	return DecryptValue(c.Value)
}

// IsEncrypted checks if the value appears to be encrypted (base64 encoded).
func (c *Credential) IsEncrypted() bool {
	if c.Value == "" {
		return false
	}
	_, err := base64.StdEncoding.DecodeString(c.Value)
	return err == nil
}
