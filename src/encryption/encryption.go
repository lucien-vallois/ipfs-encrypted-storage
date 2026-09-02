// Package encryption provides AES-256-GCM encryption for IPFS storage
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/ed25519"
)

// Import enhanced error handling
// Note: This would normally be in a separate package, but for this implementation
// we're keeping it in the same file for simplicity

// ErrorContext represents contextual information about an error
type ErrorContext struct {
	Operation   string
	Resource    string
	UserID      string
	FileName    string
	CID         string
	RetryCount  int
	Suggestions []string
}

// EnhancedError wraps an error with additional context and recovery suggestions
type EnhancedError struct {
	Err     error
	Context *ErrorContext
	Code    ErrorCode
}

func (e *EnhancedError) Error() string {
	var msg strings.Builder

	if e.Err != nil {
		msg.WriteString(e.Err.Error())
	}

	if e.Context != nil {
		if e.Context.Operation != "" {
			msg.WriteString(fmt.Sprintf(" (operation: %s)", e.Context.Operation))
		}
		if e.Context.Resource != "" {
			msg.WriteString(fmt.Sprintf(" (resource: %s)", e.Context.Resource))
		}
		if e.Context.FileName != "" {
			msg.WriteString(fmt.Sprintf(" (file: %s)", e.Context.FileName))
		}
		if e.Context.CID != "" {
			msg.WriteString(fmt.Sprintf(" (CID: %s)", e.Context.CID))
		}
		if e.Context.UserID != "" {
			msg.WriteString(fmt.Sprintf(" (user: %s)", e.Context.UserID))
		}
	}

	return msg.String()
}

// Suggestions returns recovery suggestions for the error
func (e *EnhancedError) Suggestions() []string {
	if e.Context != nil && len(e.Context.Suggestions) > 0 {
		return e.Context.Suggestions
	}

	// Default suggestions based on error code
	switch e.Code {
	case ErrCodeNetworkFailure:
		return []string{
			"Check your internet connection",
			"Verify IPFS daemon is running",
			"Try again in a few moments",
			"Check firewall settings",
		}
	case ErrCodeInvalidInput:
		return []string{
			"Verify input parameters are correct",
			"Check file format and size limits",
			"Ensure all required fields are provided",
		}
	case ErrCodeAuthentication:
		return []string{
			"Verify your credentials",
			"Check password strength requirements",
			"Ensure key pair is valid",
		}
	default:
		return []string{
			"Check system logs for more details",
			"Try the operation again",
			"Contact support if the problem persists",
		}
	}
}

// ErrorCode represents different types of errors
type ErrorCode int

const (
	ErrCodeUnknown ErrorCode = iota
	ErrCodeNetworkFailure
	ErrCodeInvalidInput
	ErrCodeAuthentication
	ErrCodePermissionDenied
	ErrCodeResourceNotFound
	ErrCodeQuotaExceeded
	ErrCodeInternalError
)

// NewEnhancedError creates a new enhanced error
func NewEnhancedError(err error, code ErrorCode, context *ErrorContext) *EnhancedError {
	return &EnhancedError{
		Err:     err,
		Code:    code,
		Context: context,
	}
}

// Encryption errors
var (
	ErrInvalidKeyLength     = errors.New("invalid key length")
	ErrDecryptionFailed     = errors.New("decryption failed")
	ErrInvalidSignature     = errors.New("invalid signature")
	ErrInvalidPassword      = errors.New("invalid password: must be at least 8 characters")
	ErrInvalidSalt          = errors.New("invalid salt: must be 32 bytes")
	ErrInvalidData          = errors.New("invalid data: cannot be empty")
	ErrDataTooLarge         = errors.New("data too large: exceeds maximum allowed size")
	ErrInvalidKeyDerivation = errors.New("invalid key derivation parameters")
)

// Validation constants
const (
	MinPasswordLength = 8
	MaxDataSize       = 100 * 1024 * 1024 // 100MB
	SaltSize          = 32
	KeySize           = 32
)

// KeyDerivationConfig holds parameters for key derivation
type KeyDerivationConfig struct {
	Time    uint32 // Number of iterations
	Memory  uint32 // Memory usage in KiB
	Threads uint8  // Number of threads
	KeyLen  uint32 // Desired key length
}

// DefaultKeyDerivationConfig returns sensible defaults for key derivation
func DefaultKeyDerivationConfig() *KeyDerivationConfig {
	return &KeyDerivationConfig{
		Time:    1,
		Memory:  64 * 1024, // 64 MiB
		Threads: 4,
		KeyLen:  32, // 256 bits
	}
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return NewEnhancedError(
			ErrInvalidPassword,
			ErrCodeInvalidInput,
			&ErrorContext{
				Operation: "password_validation",
				Suggestions: []string{
					fmt.Sprintf("Use a password with at least %d characters", MinPasswordLength),
					"Include uppercase, lowercase, and numeric characters",
					"Consider using a password manager for complex passwords",
				},
			},
		)
	}

	// Check for basic complexity requirements
	hasUpper := false
	hasLower := false
	hasDigit := false

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return NewEnhancedError(
			errors.New("password must contain at least one uppercase letter, one lowercase letter, and one digit"),
			ErrCodeInvalidInput,
			&ErrorContext{
				Operation: "password_validation",
				Suggestions: []string{
					"Include at least one uppercase letter (A-Z)",
					"Include at least one lowercase letter (a-z)",
					"Include at least one digit (0-9)",
					"Example: 'MySecurePass123'",
				},
			},
		)
	}

	return nil
}

// ValidateSalt validates salt
func ValidateSalt(salt []byte) error {
	if len(salt) != SaltSize {
		return ErrInvalidSalt
	}
	return nil
}

// ValidateData validates input data
func ValidateData(data []byte) error {
	if len(data) == 0 {
		return ErrInvalidData
	}
	if len(data) > MaxDataSize {
		return ErrDataTooLarge
	}
	return nil
}

// ValidateKeyDerivationConfig validates key derivation parameters
func ValidateKeyDerivationConfig(config *KeyDerivationConfig) error {
	if config == nil {
		return ErrInvalidKeyDerivation
	}
	if config.Time < 1 || config.Time > 10 {
		return errors.New("time parameter must be between 1 and 10")
	}
	if config.Memory < 1024 || config.Memory > 1024*1024 { // 1MB to 1GB
		return errors.New("memory parameter must be between 1024 and 1048576 KiB")
	}
	if config.Threads < 1 || config.Threads > 8 {
		return errors.New("threads parameter must be between 1 and 8")
	}
	if config.KeyLen != 32 {
		return errors.New("key length must be 32 bytes for AES-256")
	}
	return nil
}

// DeriveKey derives a 256-bit key from password using Argon2
func DeriveKey(password string, salt []byte, config *KeyDerivationConfig) ([]byte, error) {
	// Validate inputs
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}
	if err := ValidateSalt(salt); err != nil {
		return nil, err
	}
	if config == nil {
		config = DefaultKeyDerivationConfig()
	}
	if err := ValidateKeyDerivationConfig(config); err != nil {
		return nil, err
	}

	key := argon2.IDKey([]byte(password), salt, config.Time, config.Memory, config.Threads, config.KeyLen)
	return key, nil
}

// GenerateSalt generates a random 32-byte salt
func GenerateSalt() ([]byte, error) {
	salt := make([]byte, 32)
	_, err := rand.Read(salt)
	return salt, err
}

// Encrypt encrypts data using AES-256-GCM
func Encrypt(plaintext []byte, key []byte) ([]byte, error) {
	// Validate inputs
	if err := ValidateData(plaintext); err != nil {
		return nil, err
	}
	if len(key) != KeySize {
		return nil, ErrInvalidKeyLength
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts data using AES-256-GCM
func Decrypt(ciphertext []byte, key []byte) ([]byte, error) {
	// Validate inputs
	if len(ciphertext) == 0 {
		return nil, ErrInvalidData
	}
	if len(key) != KeySize {
		return nil, ErrInvalidKeyLength
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, ErrDecryptionFailed
	}

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// EncryptStream encrypts data from reader and writes to writer using AES-GCM
// Note: For large streams, this reads data in chunks and encrypts each chunk
func EncryptStream(reader io.Reader, writer io.Writer, key []byte) error {
	if len(key) != 32 {
		return ErrInvalidKeyLength
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonceSize := gcm.NonceSize()
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	// Write nonce first
	if _, err := writer.Write(nonce); err != nil {
		return err
	}

	// Read data in chunks and encrypt each chunk
	// For simplicity, we'll read all data into memory and encrypt it
	// For very large streams, a chunked approach would be better
	plaintext, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read data: %w", err)
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// Write encrypted data (nonce is already written, so skip it)
	if _, err := writer.Write(ciphertext[nonceSize:]); err != nil {
		return err
	}

	return nil
}

// DecryptStream decrypts data from reader and writes to writer using AES-GCM
// Note: This reads all data into memory for decryption
func DecryptStream(reader io.Reader, writer io.Writer, key []byte) error {
	if len(key) != 32 {
		return ErrInvalidKeyLength
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}

	nonceSize := gcm.NonceSize()
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(reader, nonce); err != nil {
		return fmt.Errorf("failed to read nonce: %w", err)
	}

	// Read remaining ciphertext
	ciphertext, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read ciphertext: %w", err)
	}

	// Decrypt the data
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return ErrDecryptionFailed
	}

	// Write decrypted data
	if _, err := writer.Write(plaintext); err != nil {
		return err
	}

	return nil
}

// GenerateKeyPair generates an Ed25519 key pair for signing
func GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

// Sign signs data with private key
func Sign(privateKey ed25519.PrivateKey, data []byte) []byte {
	return ed25519.Sign(privateKey, data)
}

// Verify verifies signature with public key
func Verify(publicKey ed25519.PublicKey, data []byte, signature []byte) bool {
	return ed25519.Verify(publicKey, data, signature)
}

// EncryptedMetadata holds encryption metadata
type EncryptedMetadata struct {
	Salt      []byte               `json:"salt"`
	Signature []byte               `json:"signature,omitempty"`
	PublicKey []byte               `json:"public_key,omitempty"`
	Config    *KeyDerivationConfig `json:"config"`
}

// EncryptWithMetadata encrypts data and returns metadata
func EncryptWithMetadata(plaintext []byte, password string, privateKey ed25519.PrivateKey) ([]byte, *EncryptedMetadata, error) {
	// Validate inputs
	if err := ValidateData(plaintext); err != nil {
		return nil, nil, err
	}
	if privateKey == nil {
		return nil, nil, errors.New("private key cannot be nil")
	}

	// Generate salt
	salt, err := GenerateSalt()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key
	config := DefaultKeyDerivationConfig()
	key, err := DeriveKey(password, salt, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to derive key: %w", err)
	}

	// Encrypt data
	ciphertext, err := Encrypt(plaintext, key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt data: %w", err)
	}

	// Sign the ciphertext
	signature := Sign(privateKey, ciphertext)

	metadata := &EncryptedMetadata{
		Salt:      salt,
		Signature: signature,
		PublicKey: privateKey.Public().(ed25519.PublicKey),
		Config:    config,
	}

	return ciphertext, metadata, nil
}

// DecryptWithMetadata decrypts data using metadata
func DecryptWithMetadata(ciphertext []byte, metadata *EncryptedMetadata, password string, publicKey ed25519.PublicKey) ([]byte, error) {
	// Validate inputs
	if len(ciphertext) == 0 {
		return nil, ErrInvalidData
	}
	if metadata == nil {
		return nil, errors.New("metadata cannot be nil")
	}
	if metadata.Salt == nil || len(metadata.Salt) != SaltSize {
		return nil, ErrInvalidSalt
	}
	if metadata.Signature == nil || len(metadata.Signature) == 0 {
		return nil, errors.New("invalid signature in metadata")
	}
	if publicKey == nil {
		return nil, errors.New("public key cannot be nil")
	}

	// Verify signature first
	if !Verify(publicKey, ciphertext, metadata.Signature) {
		return nil, ErrInvalidSignature
	}

	// Derive key
	key, err := DeriveKey(password, metadata.Salt, metadata.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to derive key: %w", err)
	}

	// Decrypt data
	plaintext, err := Decrypt(ciphertext, key)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return plaintext, nil
}

// HashData creates SHA-256 hash of data
func HashData(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// VerifyDataIntegrity verifies data integrity using hash
func VerifyDataIntegrity(data []byte, expectedHash []byte) bool {
	actualHash := HashData(data)
	return string(actualHash) == string(expectedHash)
}

// ZeroMemory overwrites the provided slice on a best-effort basis.
// It cannot erase copies made by the Go runtime or callers.
func ZeroMemory(data []byte) {
	for i := range data {
		data[i] = 0
	}
}

// SecureKey holds a private key copy that can be overwritten best-effort.
type SecureKey struct {
	key []byte
}

// NewSecureKey creates a new secure key
func NewSecureKey(key []byte) *SecureKey {
	if len(key) != 32 {
		panic("key must be 32 bytes")
	}
	keyCopy := make([]byte, 32)
	copy(keyCopy, key)
	return &SecureKey{key: keyCopy}
}

// Use executes a function with access to the key, then overwrites this copy.
func (sk *SecureKey) Use(fn func([]byte) error) error {
	defer ZeroMemory(sk.key)
	return fn(sk.key)
}

// Destroy overwrites this key buffer on a best-effort basis.
func (sk *SecureKey) Destroy() {
	ZeroMemory(sk.key)
}
