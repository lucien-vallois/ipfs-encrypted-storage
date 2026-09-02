package tests

import (
	"bytes"
	"testing"

	"ipfs-encrypted-storage/src/encryption"
)

func TestAES256GCMEncryption(t *testing.T) {
	plaintext := []byte("This is a test message for encryption")
	password := "Test-password-123"

	// Generate key pair for signing
	publicKey, privateKey, err := encryption.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Encrypt
	ciphertext, metadata, err := encryption.EncryptWithMetadata(plaintext, password, privateKey)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Decrypt
	decrypted, err := encryption.DecryptWithMetadata(ciphertext, metadata, password, publicKey)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("Decrypted text doesn't match original. Got: %s, Want: %s", decrypted, plaintext)
	}
}

func TestKeyDerivation(t *testing.T) {
	password := "Test-password1"
	salt, err := encryption.GenerateSalt()
	if err != nil {
		t.Fatalf("Failed to generate salt: %v", err)
	}

	config := encryption.DefaultKeyDerivationConfig()
	key1, err := encryption.DeriveKey(password, salt, config)
	if err != nil {
		t.Fatalf("Failed to derive first key: %v", err)
	}

	key2, err := encryption.DeriveKey(password, salt, config)
	if err != nil {
		t.Fatalf("Failed to derive second key: %v", err)
	}

	if !bytes.Equal(key1, key2) {
		t.Error("Key derivation should be deterministic with same inputs")
	}

	// Test different password produces different key
	key3, err := encryption.DeriveKey("Different-password1", salt, config)
	if err != nil {
		t.Fatalf("Failed to derive key with different password: %v", err)
	}
	if bytes.Equal(key1, key3) {
		t.Error("Different passwords should produce different keys")
	}
}

func TestEd25519Signatures(t *testing.T) {
	message := []byte("Test message for signing")

	publicKey, privateKey, err := encryption.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Sign message
	signature := encryption.Sign(privateKey, message)

	// Verify signature
	if !encryption.Verify(publicKey, message, signature) {
		t.Error("Signature verification failed")
	}

	// Test with tampered message
	tamperedMessage := []byte("Tampered message")
	if encryption.Verify(publicKey, tamperedMessage, signature) {
		t.Error("Signature verification should fail for tampered message")
	}
}

func TestStreamEncryption(t *testing.T) {
	plaintext := []byte("This is a longer test message for stream encryption testing")
	key := []byte("32-byte-key-for-testing-12345678")

	// Test stream encryption/decryption
	var encrypted bytes.Buffer
	reader := bytes.NewReader(plaintext)

	err := encryption.EncryptStream(reader, &encrypted, key)
	if err != nil {
		t.Fatalf("Failed to encrypt stream: %v", err)
	}

	var decrypted bytes.Buffer
	encryptedReader := bytes.NewReader(encrypted.Bytes())

	err = encryption.DecryptStream(encryptedReader, &decrypted, key)
	if err != nil {
		t.Fatalf("Failed to decrypt stream: %v", err)
	}

	if !bytes.Equal(plaintext, decrypted.Bytes()) {
		t.Errorf("Stream decrypted text doesn't match original. Got: %s, Want: %s", decrypted.Bytes(), plaintext)
	}
}

func TestSecureKeyHandling(t *testing.T) {
	keyData := []byte("32-byte-key-for-testing-12345678")

	secureKey := encryption.NewSecureKey(keyData)

	// Test using the key
	var keyView []byte
	err := secureKey.Use(func(key []byte) error {
		if !bytes.Equal(keyData, key) {
			t.Error("Secure key use failed")
		}
		keyView = key
		return nil
	})

	if err != nil {
		t.Fatalf("Failed to use secure key: %v", err)
	}

	// Test key destruction
	secureKey.Destroy()

	// After destruction, the key should be zeroed
	for _, b := range keyView {
		if b != 0 {
			t.Error("Key was not properly zeroed after destruction")
		}
	}
}

func BenchmarkAES256GCMEncryption(b *testing.B) {
	plaintext := []byte("This is a benchmark test message for AES-256-GCM encryption")
	password := "Benchmark-password1"

	_, privateKey, err := encryption.GenerateKeyPair()
	if err != nil {
		b.Fatalf("Failed to generate key pair: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, err := encryption.EncryptWithMetadata(plaintext, password, privateKey)
		if err != nil {
			b.Fatalf("Encryption failed: %v", err)
		}
	}
}

func BenchmarkAES256GCMDecryption(b *testing.B) {
	plaintext := []byte("This is a benchmark test message for AES-256-GCM decryption")
	password := "Benchmark-password1"

	publicKey, privateKey, err := encryption.GenerateKeyPair()
	if err != nil {
		b.Fatalf("Failed to generate key pair: %v", err)
	}

	ciphertext, metadata, err := encryption.EncryptWithMetadata(plaintext, password, privateKey)
	if err != nil {
		b.Fatalf("Failed to encrypt: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := encryption.DecryptWithMetadata(ciphertext, metadata, password, publicKey)
		if err != nil {
			b.Fatalf("Decryption failed: %v", err)
		}
	}
}

func BenchmarkKeyDerivation(b *testing.B) {
	password := "Benchmark-password1"
	salt, err := encryption.GenerateSalt()
	if err != nil {
		b.Fatalf("Failed to generate salt: %v", err)
	}

	config := encryption.DefaultKeyDerivationConfig()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if _, err := encryption.DeriveKey(password, salt, config); err != nil {
			b.Fatalf("Key derivation failed: %v", err)
		}
	}
}
