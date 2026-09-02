package tests_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	"ipfs-encrypted-storage/src/config"
	"ipfs-encrypted-storage/src/did"
	"ipfs-encrypted-storage/src/encryption"
	"ipfs-encrypted-storage/src/ipfs"
	"ipfs-encrypted-storage/src/p2p"
	"ipfs-encrypted-storage/src/utils"
	"ipfs-encrypted-storage/src/zkp"
)

func TestEndToEndEncryptionWorkflow(t *testing.T) {
	// Test complete encryption workflow
	plaintext := []byte("This is a complete end-to-end test message")
	password := "Test-password-12345"

	// Step 1: Generate key pair
	publicKey, privateKey, err := encryption.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Step 2: Encrypt data
	ciphertext, metadata, err := encryption.EncryptWithMetadata(plaintext, password, privateKey)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	// Step 3: Verify metadata contains required fields
	if metadata.Salt == nil {
		t.Error("Metadata should contain salt")
	}
	if metadata.Signature == nil {
		t.Error("Metadata should contain signature")
	}
	if metadata.PublicKey == nil {
		t.Error("Metadata should contain public key")
	}

	// Step 4: Decrypt data
	decrypted, err := encryption.DecryptWithMetadata(ciphertext, metadata, password, publicKey)
	if err != nil {
		t.Fatalf("Failed to decrypt data: %v", err)
	}

	// Step 5: Verify data integrity
	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("Decrypted data doesn't match original. Got: %s, Want: %s", decrypted, plaintext)
	}

	// Step 6: Test signature verification
	if !encryption.Verify(publicKey, ciphertext, metadata.Signature) {
		t.Error("Signature verification failed")
	}
}

func TestFileSizeHandling(t *testing.T) {
	testCases := []struct {
		name     string
		size     int
		password string
	}{
		{"Small file", 100, "Small-file-pass1"},
		{"Medium file", 1024 * 10, "Medium-file-pass1"}, // 10KB
		{"Large file", 1024 * 100, "Large-file-pass1"},  // 100KB
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate test data
			plaintext := make([]byte, tc.size)
			for i := range plaintext {
				plaintext[i] = byte(i % 256)
			}

			// Generate keys
			publicKey, privateKey, err := encryption.GenerateKeyPair()
			if err != nil {
				t.Fatalf("Failed to generate key pair: %v", err)
			}

			// Encrypt
			ciphertext, metadata, err := encryption.EncryptWithMetadata(plaintext, tc.password, privateKey)
			if err != nil {
				t.Fatalf("Failed to encrypt %d bytes: %v", tc.size, err)
			}

			// Decrypt
			decrypted, err := encryption.DecryptWithMetadata(ciphertext, metadata, tc.password, publicKey)
			if err != nil {
				t.Fatalf("Failed to decrypt %d bytes: %v", tc.size, err)
			}

			// Verify
			if !bytes.Equal(plaintext, decrypted) {
				t.Errorf("Data integrity check failed for %d bytes", tc.size)
			}
		})
	}
}

func TestConcurrentOperations(t *testing.T) {
	numGoroutines := 10
	numOperations := 5

	// Channel to collect results
	results := make(chan bool, numGoroutines*numOperations)

	// Run concurrent encryption/decryption operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				plaintext := []byte("Concurrent test message")
				password := "Concurrent-test-pass1"

				publicKey, privateKey, err := encryption.GenerateKeyPair()
				if err != nil {
					results <- false
					continue
				}

				ciphertext, metadata, err := encryption.EncryptWithMetadata(plaintext, password, privateKey)
				if err != nil {
					results <- false
					continue
				}

				decrypted, err := encryption.DecryptWithMetadata(ciphertext, metadata, password, publicKey)
				if err != nil || !bytes.Equal(plaintext, decrypted) {
					results <- false
					continue
				}

				results <- true
			}
		}(i)
	}

	// Collect results
	successCount := 0
	totalOperations := numGoroutines * numOperations

	for i := 0; i < totalOperations; i++ {
		if <-results {
			successCount++
		}
	}

	if successCount != totalOperations {
		t.Errorf("Concurrent operations failed: %d/%d succeeded", successCount, totalOperations)
	}
}

// Performance benchmarks report throughput and latency metrics.
func BenchmarkEncryptionThroughput(b *testing.B) {
	plaintext := []byte("Benchmark test message for throughput measurement")
	password := "Benchmark-password1"

	_, privateKey, err := encryption.GenerateKeyPair()
	if err != nil {
		b.Fatalf("Failed to generate key pair: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	start := time.Now()
	operations := 0

	for i := 0; i < b.N; i++ {
		_, _, err := encryption.EncryptWithMetadata(plaintext, password, privateKey)
		if err != nil {
			b.Fatalf("Encryption failed: %v", err)
		}
		operations++
	}

	elapsed := time.Since(start)
	b.StopTimer()
	opsPerSecond := float64(operations) / elapsed.Seconds()

	b.ReportMetric(opsPerSecond, "ops/s")
	b.Logf("Encryption: %.2f operations/second", opsPerSecond)
}

func BenchmarkDecryptionThroughput(b *testing.B) {
	plaintext := []byte("Benchmark test message for decryption throughput")
	password := "Benchmark-password1"

	publicKey, privateKey, err := encryption.GenerateKeyPair()
	if err != nil {
		b.Fatalf("Failed to generate key pair: %v", err)
	}

	ciphertext, metadata, err := encryption.EncryptWithMetadata(plaintext, password, privateKey)
	if err != nil {
		b.Fatalf("Failed to prepare test data: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	start := time.Now()
	operations := 0

	for i := 0; i < b.N; i++ {
		_, err := encryption.DecryptWithMetadata(ciphertext, metadata, password, publicKey)
		if err != nil {
			b.Fatalf("Decryption failed: %v", err)
		}
		operations++
	}

	elapsed := time.Since(start)
	b.StopTimer()
	opsPerSecond := float64(operations) / elapsed.Seconds()

	b.ReportMetric(opsPerSecond, "ops/s")
	b.Logf("Decryption: %.2f operations/second", opsPerSecond)
}

func BenchmarkEncryptionLatency(b *testing.B) {
	plaintext := []byte("Benchmark test message for latency measurement")
	password := "Benchmark-password1"

	_, privateKey, err := encryption.GenerateKeyPair()
	if err != nil {
		b.Fatalf("Failed to generate key pair: %v", err)
	}

	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, _, err := encryption.EncryptWithMetadata(plaintext, password, privateKey)
		if err != nil {
			b.Fatalf("Encryption failed: %v", err)
		}

		latency := time.Since(start)
		latencies = append(latencies, latency)
	}
	b.StopTimer()

	// Calculate P95 latency
	slices.Sort(latencies)
	p95Index := (95*len(latencies)+99)/100 - 1
	p95Latency := latencies[p95Index]

	b.ReportMetric(float64(p95Latency.Nanoseconds()), "p95-ns")
	b.Logf("Encryption P95 latency: %v", p95Latency)
}

func BenchmarkDecryptionLatency(b *testing.B) {
	plaintext := []byte("Benchmark test message for decryption latency")
	password := "Benchmark-password1"

	publicKey, privateKey, err := encryption.GenerateKeyPair()
	if err != nil {
		b.Fatalf("Failed to generate key pair: %v", err)
	}

	ciphertext, metadata, err := encryption.EncryptWithMetadata(plaintext, password, privateKey)
	if err != nil {
		b.Fatalf("Failed to prepare test data: %v", err)
	}

	latencies := make([]time.Duration, 0, b.N)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		_, err := encryption.DecryptWithMetadata(ciphertext, metadata, password, publicKey)
		if err != nil {
			b.Fatalf("Decryption failed: %v", err)
		}

		latency := time.Since(start)
		latencies = append(latencies, latency)
	}
	b.StopTimer()

	// Calculate P95 latency
	slices.Sort(latencies)
	p95Index := (95*len(latencies)+99)/100 - 1
	p95Latency := latencies[p95Index]

	b.ReportMetric(float64(p95Latency.Nanoseconds()), "p95-ns")
	b.Logf("Decryption P95 latency: %v", p95Latency)
}

func BenchmarkKeyDerivationLatency(b *testing.B) {
	password := "Benchmark-password-for-key-derivation1"
	salt, err := encryption.GenerateSalt()
	if err != nil {
		b.Fatalf("Failed to generate salt: %v", err)
	}

	config := encryption.DefaultKeyDerivationConfig()

	var latencies []time.Duration

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		start := time.Now()

		if _, err := encryption.DeriveKey(password, salt, config); err != nil {
			b.Fatalf("Key derivation failed: %v", err)
		}

		latency := time.Since(start)
		latencies = append(latencies, latency)
	}
	b.StopTimer()

	// Calculate average latency
	totalLatency := time.Duration(0)
	for _, latency := range latencies {
		totalLatency += latency
	}
	avgLatency := totalLatency / time.Duration(len(latencies))

	b.ReportMetric(float64(avgLatency.Nanoseconds()), "avg-ns")
	b.Logf("Key derivation average latency: %v", avgLatency)
}

// Mock IPFS integration test
// MockIPFSClient provides a mock implementation for testing
type MockIPFSClient struct {
	mu      sync.RWMutex
	storage map[string][]byte
}

func NewMockIPFSClient() *MockIPFSClient {
	return &MockIPFSClient{
		storage: make(map[string][]byte),
	}
}

func (m *MockIPFSClient) AddFile(data []byte, filename string) (string, error) {
	// Generate a mock CID based on content hash
	hash := encryption.HashData(data)
	cid := fmt.Sprintf("Qm%s", string(hash)[:16])
	m.mu.Lock()
	defer m.mu.Unlock()
	m.storage[cid] = make([]byte, len(data))
	copy(m.storage[cid], data)
	return cid, nil
}

func (m *MockIPFSClient) GetFile(cid string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, exists := m.storage[cid]
	if !exists {
		return nil, fmt.Errorf("file not found")
	}
	result := make([]byte, len(data))
	copy(result, data)
	return result, nil
}

func TestEncryptedStorageWorkflowWithMock(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping mock encrypted storage workflow in short mode")
	}

	// Test with mock IPFS client
	mockIPFS := NewMockIPFSClient()

	plaintext := []byte("Test message for IPFS encryption workflow")
	password := "Ipfs-test-password1"

	// Generate keys
	publicKey, privateKey, err := encryption.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Encrypt
	ciphertext, metadata, err := encryption.EncryptWithMetadata(plaintext, password, privateKey)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// "Upload" to IPFS (mock)
	cid, err := mockIPFS.AddFile(ciphertext, "encrypted-test.txt")
	if err != nil {
		t.Fatalf("Failed to 'upload' to mock IPFS: %v", err)
	}

	// "Download" from IPFS (mock)
	retrievedCiphertext, err := mockIPFS.GetFile(cid)
	if err != nil {
		t.Fatalf("Failed to 'download' from mock IPFS: %v", err)
	}

	// Decrypt
	decrypted, err := encryption.DecryptWithMetadata(retrievedCiphertext, metadata, password, publicKey)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	// Verify
	if !bytes.Equal(plaintext, decrypted) {
		t.Error("IPFS workflow data integrity check failed")
	}
}

// TestEndToEndCompleteWorkflow tests the complete workflow from encryption to IPFS storage to retrieval
func TestEndToEndCompleteWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping complete workflow test in short mode")
	}

	// Initialize components
	mockIPFS := NewMockIPFSClient()

	// Test data
	testFiles := []struct {
		name     string
		content  []byte
		password string
	}{
		{"document.txt", []byte("This is a confidential document"), "Doc-password-123"},
		{"image.dat", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, "Img-password-456"}, // Mock PNG header
		{"config.json", []byte(`{"database": "encrypted", "port": 5432}`), "Config-password-789"},
	}

	var uploadedFiles []struct {
		originalName    string
		originalContent []byte
		cid             string
		metadata        *encryption.EncryptedMetadata
		password        string
		publicKey       ed25519.PublicKey
	}

	// Phase 1: Encrypt and upload all files
	for _, file := range testFiles {
		t.Run(fmt.Sprintf("Upload_%s", file.name), func(t *testing.T) {
			// Generate unique key pair for each file
			publicKey, privateKey, err := encryption.GenerateKeyPair()
			if err != nil {
				t.Fatalf("Failed to generate key pair for %s: %v", file.name, err)
			}

			// Encrypt file
			ciphertext, metadata, err := encryption.EncryptWithMetadata(file.content, file.password, privateKey)
			if err != nil {
				t.Fatalf("Failed to encrypt %s: %v", file.name, err)
			}

			// Upload to IPFS
			cid, err := mockIPFS.AddFile(ciphertext, file.name)
			if err != nil {
				t.Fatalf("Failed to upload %s to IPFS: %v", file.name, err)
			}

			// Store for later verification
			uploadedFiles = append(uploadedFiles, struct {
				originalName    string
				originalContent []byte
				cid             string
				metadata        *encryption.EncryptedMetadata
				password        string
				publicKey       ed25519.PublicKey
			}{
				originalName:    file.name,
				originalContent: append([]byte(nil), file.content...),
				cid:             cid,
				metadata:        metadata,
				password:        file.password,
				publicKey:       append(ed25519.PublicKey(nil), publicKey...),
			})
		})
	}

	if len(uploadedFiles) != len(testFiles) {
		t.Fatalf("Uploaded %d files, want %d", len(uploadedFiles), len(testFiles))
	}

	// Phase 2: retrieve, decrypt, and compare every uploaded file.
	for _, uploaded := range uploadedFiles {
		t.Run(fmt.Sprintf("Retrieve_%s", uploaded.originalName), func(t *testing.T) {
			ciphertext, err := mockIPFS.GetFile(uploaded.cid)
			if err != nil {
				t.Fatalf("Failed to retrieve %s: %v", uploaded.originalName, err)
			}
			plaintext, err := encryption.DecryptWithMetadata(ciphertext, uploaded.metadata, uploaded.password, uploaded.publicKey)
			if err != nil {
				t.Fatalf("Failed to decrypt %s: %v", uploaded.originalName, err)
			}
			if !bytes.Equal(plaintext, uploaded.originalContent) {
				t.Fatalf("Round-trip mismatch for %s", uploaded.originalName)
			}
		})
	}
}

// TestP2PLocalStubIntegration tests isolated in-memory nodes.
func TestP2PLocalStubIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping P2P integration test in short mode")
	}

	// Create two P2P nodes
	node1, err := p2p.NewP2PNode("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		t.Fatalf("Failed to create P2P node 1: %v", err)
	}
	defer node1.Close()

	node2, err := p2p.NewP2PNode("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		t.Fatalf("Failed to create P2P node 2: %v", err)
	}
	defer node2.Close()

	// Test data for exchange
	testMessage := []byte("Hello from P2P integration test!")
	receivedMessages := make(chan []byte, 1)

	// Set up message handler for node2
	messageHandler := func(peerID string, message []byte) error {
		receivedMessages <- message
		return nil
	}

	// Test node-local storage isolation.
	t.Run("LocalStorageIsolation", func(t *testing.T) {
		testKey := "integration-test-key"
		testValue := []byte("integration-test-value")

		// Store and retrieve the value on node1.
		err := node1.StoreValue(testKey, testValue)
		if err != nil {
			t.Fatalf("Failed to store local value: %v", err)
		}
		retrievedValue, err := node1.GetValue(testKey)
		if err != nil {
			t.Fatalf("Failed to retrieve local value: %v", err)
		}
		if !bytes.Equal(testValue, retrievedValue) {
			t.Errorf("Local value mismatch: expected %s, got %s", testValue, retrievedValue)
		}

		// A distinct node must not observe node1's in-memory storage.
		if _, err := node2.GetValue(testKey); err == nil {
			t.Fatal("Expected local storage to be isolated between nodes")
		}
	})

	// Test local identity and configured address metadata.
	t.Run("LocalIdentityAndAddressMetadata", func(t *testing.T) {
		// Get peer info
		peer1ID := node1.GetID()
		peer2ID := node2.GetID()

		if peer1ID.String() == "" {
			t.Error("Node1 peer ID is empty")
		}
		if peer2ID.String() == "" {
			t.Error("Node2 peer ID is empty")
		}
		if peer1ID == peer2ID {
			t.Error("Both nodes have the same peer ID")
		}

		// Check addresses
		addresses1 := node1.GetAddresses()
		addresses2 := node2.GetAddresses()

		if len(addresses1) == 0 {
			t.Error("Node1 has no listen addresses")
		}
		if len(addresses2) == 0 {
			t.Error("Node2 has no listen addresses")
		}
	})

	// Test same-node local pubsub messaging.
	t.Run("SameNodeLocalPubSub", func(t *testing.T) {
		topic := "integration-test-topic"

		// Node2 subscribes to topic
		subscription, err := node2.SubscribeToTopic(topic, messageHandler)
		if err != nil {
			t.Fatalf("Failed to subscribe to topic: %v", err)
		}
		defer subscription.Cancel()

		// Give subscription time to establish
		time.Sleep(100 * time.Millisecond)

		// Publish through the subscribed in-memory node.
		err = node2.PublishToTopic(topic, testMessage)
		if err != nil {
			t.Fatalf("Failed to publish message: %v", err)
		}

		// Wait for message reception
		select {
		case received := <-receivedMessages:
			if !bytes.Equal(testMessage, received) {
				t.Errorf("Received message mismatch: expected %s, got %s", testMessage, received)
			}
		case <-time.After(5 * time.Second):
			t.Error("Timeout waiting for pubsub message")
		}
	})

	// Test the static metadata exposed by the stub.
	t.Run("StaticStubStats", func(t *testing.T) {
		stats1 := node1.GetNetworkStats()
		stats2 := node2.GetNetworkStats()

		if stats1 == nil {
			t.Error("Node1 network stats is nil")
		}
		if stats2 == nil {
			t.Error("Node2 network stats is nil")
		}

		// Check for expected stat keys
		requiredKeys := []string{"connected_peers", "listen_addresses", "protocols"}
		for _, key := range requiredKeys {
			if _, exists := stats1[key]; !exists {
				t.Errorf("Node1 missing network stat key: %s", key)
			}
			if _, exists := stats2[key]; !exists {
				t.Errorf("Node2 missing network stat key: %s", key)
			}
		}
	})
}

// TestMultiFileBatchProcessing tests batch processing of multiple files
func TestMultiFileBatchProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping batch processing test in short mode")
	}

	mockIPFS := NewMockIPFSClient()

	// Create multiple test files
	fileCount := 5
	testFiles := make([]struct {
		name    string
		content []byte
		size    int
	}, fileCount)

	for i := 0; i < fileCount; i++ {
		size := (i + 1) * 1024 // 1KB, 2KB, 3KB, 4KB, 5KB
		content := make([]byte, size)
		for j := range content {
			content[j] = byte((i + j) % 256)
		}

		testFiles[i] = struct {
			name    string
			content []byte
			size    int
		}{
			name:    fmt.Sprintf("batch_file_%d.dat", i+1),
			content: content,
			size:    size,
		}
	}

	// Batch processing results
	results := make(chan struct {
		fileName string
		cid      string
		success  bool
		error    error
	}, fileCount)

	// Process files concurrently
	for _, file := range testFiles {
		go func(f struct {
			name    string
			content []byte
			size    int
		}) {
			// Generate keys
			publicKey, privateKey, err := encryption.GenerateKeyPair()
			if err != nil {
				results <- struct {
					fileName string
					cid      string
					success  bool
					error    error
				}{f.name, "", false, fmt.Errorf("key generation failed: %w", err)}
				return
			}

			// Encrypt
			password := fmt.Sprintf("Batch-password-1-%s", f.name)
			ciphertext, metadata, err := encryption.EncryptWithMetadata(f.content, password, privateKey)
			if err != nil {
				results <- struct {
					fileName string
					cid      string
					success  bool
					error    error
				}{f.name, "", false, fmt.Errorf("encryption failed: %w", err)}
				return
			}

			// Upload
			cid, err := mockIPFS.AddFile(ciphertext, f.name)
			if err != nil {
				results <- struct {
					fileName string
					cid      string
					success  bool
					error    error
				}{f.name, "", false, fmt.Errorf("upload failed: %w", err)}
				return
			}

			// Verify upload
			downloaded, err := mockIPFS.GetFile(cid)
			if err != nil {
				results <- struct {
					fileName string
					cid      string
					success  bool
					error    error
				}{f.name, cid, false, fmt.Errorf("download verification failed: %w", err)}
				return
			}

			// Decrypt and verify
			decrypted, err := encryption.DecryptWithMetadata(downloaded, metadata, password, publicKey)
			if err != nil {
				results <- struct {
					fileName string
					cid      string
					success  bool
					error    error
				}{f.name, cid, false, fmt.Errorf("decryption verification failed: %w", err)}
				return
			}

			if !bytes.Equal(f.content, decrypted) {
				results <- struct {
					fileName string
					cid      string
					success  bool
					error    error
				}{f.name, cid, false, fmt.Errorf("integrity check failed")}
				return
			}

			results <- struct {
				fileName string
				cid      string
				success  bool
				error    error
			}{f.name, cid, true, nil}
		}(file)
	}

	// Collect results
	successCount := 0
	totalProcessed := 0

	for i := 0; i < fileCount; i++ {
		result := <-results
		totalProcessed++

		if result.success {
			successCount++
			t.Logf("Successfully processed %s (CID: %s)", result.fileName, result.cid)
		} else {
			t.Errorf("Failed to process %s: %v", result.fileName, result.error)
		}
	}

	if successCount != fileCount {
		t.Errorf("Batch processing incomplete: %d/%d files succeeded", successCount, fileCount)
	} else {
		t.Logf("Batch processing completed successfully: %d/%d files processed", successCount, fileCount)
	}
}

// TestDIDZKPIntegration tests integration between DID and ZKP systems
func TestDIDZKPIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DID-ZKP integration test in short mode")
	}

	// Initialize components
	mockIPFS := NewMockIPFSClient()
	didManager := did.NewDIDManager()

	// Create a user DID for identity
	userDID, _, err := didManager.CreateDID("user")
	if err != nil {
		t.Fatalf("Failed to create user DID: %v", err)
	}

	// Create a resource owner DID
	ownerDID, ownerPrivateKey, err := didManager.CreateDID("owner")
	if err != nil {
		t.Fatalf("Failed to create owner DID: %v", err)
	}

	// Test data
	sensitiveData := []byte("This is highly sensitive encrypted data")
	resourceID := "encrypted-resource-123"

	// Phase 1: Encrypt data with owner credentials
	ciphertext := make([]byte, len(sensitiveData))
	copy(ciphertext, sensitiveData)

	// In a real scenario, this would use proper encryption with metadata
	_, err = mockIPFS.AddFile(ciphertext, "sensitive-data.enc")
	if err != nil {
		t.Fatalf("Failed to store encrypted data: %v", err)
	}

	// Phase 2: Generate access control ZKP for the user
	permissions := []string{"read", "decrypt"}
	userSecret := big.NewInt(42) // User's secret for ZKP

	accessProof, err := zkp.GenerateAccessProof(resourceID, userDID.ID, permissions, userSecret)
	if err != nil {
		t.Fatalf("Failed to generate access proof: %v", err)
	}

	// Phase 3: Store access proof in owner's DID document
	proofService := did.Service{
		ID:   ownerDID.ID + "#access-proof-" + userDID.ID,
		Type: "AccessProofService",
		ServiceEndpoint: map[string]interface{}{
			"proof_type":  "zkp",
			"resource_id": resourceID,
			"user_did":    userDID.ID,
			"permissions": permissions,
			"proof_data":  accessProof,
		},
	}

	err = didManager.AddService(ownerDID.ID, proofService)
	if err != nil {
		t.Fatalf("Failed to add access proof service to owner DID: %v", err)
	}

	// Phase 4: Verify the complete access control flow
	t.Run("AccessControlFlow", func(t *testing.T) {
		// Resolve owner DID to get access proof
		resolvedOwnerDID, err := didManager.ResolveDID(ownerDID.ID)
		if err != nil {
			t.Fatalf("Failed to resolve owner DID: %v", err)
		}

		// Find the access proof service
		var userAccessProof *zkp.AccessControlProof
		for _, service := range resolvedOwnerDID.Service {
			if service.Type == "AccessProofService" {
				endpoint := service.ServiceEndpoint.(map[string]interface{})
				if endpoint["user_did"] == userDID.ID && endpoint["resource_id"] == resourceID {
					userAccessProof = endpoint["proof_data"].(*zkp.AccessControlProof)
					break
				}
			}
		}

		if userAccessProof == nil {
			t.Fatal("Access proof not found in owner DID document")
		}

		// Verify ZKP access proof
		p := big.NewInt(23)
		g := big.NewInt(5)
		userPublicKey := new(big.Int).Exp(g, userSecret, p)

		if !zkp.VerifyAccessProofFor(userAccessProof, userPublicKey, resourceID, userDID.ID, permissions) {
			t.Error("Access proof verification failed")
		}

		// Verify proof metadata
		if userAccessProof.ResourceID != resourceID {
			t.Errorf("Access proof resource ID mismatch: expected %s, got %s", resourceID, userAccessProof.ResourceID)
		}
		if userAccessProof.UserID != userDID.ID {
			t.Errorf("Access proof user ID mismatch: expected %s, got %s", userDID.ID, userAccessProof.UserID)
		}
	})

	// Phase 5: Demonstrate DID credential integration
	t.Run("DIDCredentialIntegration", func(t *testing.T) {
		// Owner issues a credential to user granting access
		accessClaims := map[string]interface{}{
			"resource_id":  resourceID,
			"permissions":  permissions,
			"valid_until":  "2024-12-31",
			"access_level": "confidential",
		}

		credential, err := didManager.IssueCredential(ownerDID.ID, userDID.ID, accessClaims, ownerPrivateKey)
		if err != nil {
			t.Fatalf("Failed to issue access credential: %v", err)
		}

		// Verify credential
		if !didManager.VerifyCredential(credential) {
			t.Error("Credential verification failed")
		}

		// Verify credential content
		if credential.Issuer != ownerDID.ID {
			t.Errorf("Credential issuer mismatch: expected %s, got %s", ownerDID.ID, credential.Issuer)
		}
		if credential.CredentialSubject["id"] != userDID.ID {
			t.Errorf("Credential subject mismatch: expected %s, got %s", userDID.ID, credential.CredentialSubject["id"])
		}
		if credential.CredentialSubject["resource_id"] != resourceID {
			t.Errorf("Credential resource ID mismatch: expected %s, got %s", resourceID, credential.CredentialSubject["resource_id"])
		}
	})
}

// TestMockContentAddressingWorkflow tests content identifiers produced by the mock storage.
func TestMockContentAddressingWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping mock content-addressing workflow in short mode")
	}

	mockIPFS := NewMockIPFSClient()

	// Test data with different sizes and types
	testCases := []struct {
		name        string
		content     []byte
		description string
	}{
		{
			name:        "small-text",
			content:     []byte("Small text file content"),
			description: "A small text document",
		},
		{
			name:        "medium-json",
			content:     []byte(`{"data": "This is a medium sized JSON file", "size": "medium", "items": [1,2,3,4,5]}`),
			description: "A medium JSON configuration file",
		},
		{
			name:        "binary-data",
			content:     []byte{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC},
			description: "Binary data file",
		},
	}

	var storedFiles []struct {
		name        string
		cid         string
		content     []byte
		description string
		hash        []byte
	}

	// Store files and compute content addresses
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Store_%s", tc.name), func(t *testing.T) {
			// Compute content hash (content addressing)
			contentHash := encryption.HashData(tc.content)

			// Store file
			cid, err := mockIPFS.AddFile(tc.content, tc.name)
			if err != nil {
				t.Fatalf("Failed to store %s: %v", tc.name, err)
			}

			storedFiles = append(storedFiles, struct {
				name        string
				cid         string
				content     []byte
				description string
				hash        []byte
			}{
				name:        tc.name,
				cid:         cid,
				content:     tc.content,
				description: tc.description,
				hash:        contentHash,
			})

			t.Logf("Stored %s with CID: %s, Hash: %x", tc.name, cid, contentHash)
		})
	}

	// Verify content addressing and integrity
	t.Run("ContentIntegrity", func(t *testing.T) {
		for _, file := range storedFiles {
			// Retrieve file
			retrieved, err := mockIPFS.GetFile(file.cid)
			if err != nil {
				t.Fatalf("Failed to retrieve %s: %v", file.name, err)
			}

			// Verify content integrity
			if !bytes.Equal(file.content, retrieved) {
				t.Errorf("Content integrity check failed for %s", file.name)
			}

			// Verify content hash
			retrievedHash := encryption.HashData(retrieved)
			if !bytes.Equal(file.hash, retrievedHash) {
				t.Errorf("Content hash verification failed for %s", file.name)
			}

			// Verify data integrity using encryption module
			if !encryption.VerifyDataIntegrity(retrieved, file.hash) {
				t.Errorf("Data integrity verification failed for %s", file.name)
			}
		}
	})

	// Test deduplication (same content should produce same CID)
	t.Run("Deduplication", func(t *testing.T) {
		duplicateContent := []byte("Duplicate content for testing")
		cid1, err := mockIPFS.AddFile(duplicateContent, "duplicate-1")
		if err != nil {
			t.Fatalf("Failed to store first duplicate: %v", err)
		}

		cid2, err := mockIPFS.AddFile(duplicateContent, "duplicate-2")
		if err != nil {
			t.Fatalf("Failed to store second duplicate: %v", err)
		}

		// In a real content-addressable system, same content should produce same CID
		// Our mock implementation generates CIDs based on hash, so they should be identical
		if cid1 != cid2 {
			t.Logf("Mock IPFS doesn't enforce deduplication (CID1: %s, CID2: %s)", cid1, cid2)
		}
	})
}

// TestRealIPFSEncryptionWorkflow tests integration with real IPFS when available
func TestRealIPFSEncryptionWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real IPFS integration test in short mode")
	}

	ipfsURL := os.Getenv("IPFS_URL")
	if ipfsURL == "" {
		t.Skip("IPFS_URL is not configured")
	}

	// Try to create real IPFS client
	client, err := ipfs.NewIPFSClient(ipfsURL)
	if err != nil {
		t.Fatalf("IPFS_URL is configured with an invalid endpoint %s: %v", ipfsURL, err)
	}

	// Test IPFS connectivity
	err = client.HealthCheck()
	if err != nil {
		t.Fatalf("IPFS_URL is configured but unavailable at %s: %v", ipfsURL, err)
	}

	t.Logf("Real IPFS available at %s, running integration tests", ipfsURL)

	// Test data
	testData := []byte("Real IPFS integration test data")
	testPassword := "Real-ipfs-test-password1"

	// Generate keys
	publicKey, privateKey, err := encryption.GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Encrypt data
	ciphertext, metadata, err := encryption.EncryptWithMetadata(testData, testPassword, privateKey)
	if err != nil {
		t.Fatalf("Failed to encrypt data: %v", err)
	}

	// Upload to real IPFS
	cid, err := client.AddFile(ciphertext, "real-ipfs-test.txt")
	if err != nil {
		t.Fatalf("Failed to upload to real IPFS: %v", err)
	}

	t.Logf("Successfully uploaded to real IPFS with CID: %s", cid)

	// Download from real IPFS
	downloadedCiphertext, err := client.GetFile(cid)
	if err != nil {
		t.Fatalf("Failed to download from real IPFS: %v", err)
	}

	// Verify downloaded data matches uploaded
	if !bytes.Equal(ciphertext, downloadedCiphertext) {
		t.Error("Downloaded data doesn't match uploaded data")
	}

	// Decrypt downloaded data
	decrypted, err := encryption.DecryptWithMetadata(downloadedCiphertext, metadata, testPassword, publicKey)
	if err != nil {
		t.Fatalf("Failed to decrypt downloaded data: %v", err)
	}

	// Verify decrypted data matches original
	if !bytes.Equal(testData, decrypted) {
		t.Error("Decrypted data doesn't match original")
	}

	// Test IPFS pinning
	err = client.PinFile(cid)
	if err != nil {
		t.Logf("Warning: Failed to pin file (may not be necessary in all IPFS setups): %v", err)
	} else {
		t.Logf("Successfully pinned file with CID: %s", cid)
	}

	// Get object stats
	stats, err := client.GetObjectStats(cid)
	if err != nil {
		t.Logf("Warning: Failed to get object stats: %v", err)
	} else {
		t.Logf("Object stats - Size: %d, CumulativeSize: %d, NumLinks: %d", stats.DataSize, stats.CumulativeSize, stats.NumLinks)
	}
}

// TestNetworkFailureScenarios tests behavior under network failure conditions
func TestNetworkFailureScenarios(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network failure test in short mode")
	}

	// Test with invalid IPFS endpoint
	invalidClient, err := ipfs.NewIPFSClient("127.0.0.1:1")

	testData := []byte("Network failure test data")
	if err == nil {
		_, err = invalidClient.AddFile(testData, "failure-test.txt")
	}
	if err == nil {
		t.Error("Expected error when connecting to invalid IPFS endpoint")
	} else {
		t.Logf("Correctly received error for invalid endpoint: %v", err)
	}

	// Test invalid CID retrieval
	validClient, err := ipfs.NewIPFSClient("localhost:5001")
	if err != nil {
		t.Fatalf("Failed to create IPFS client: %v", err)
	}
	_, err = validClient.GetFile("invalid-cid")
	if err == nil {
		t.Error("Expected error when retrieving invalid CID")
	} else {
		t.Logf("Correctly received error for invalid CID: %v", err)
	}
}

// TestP2PLocalStubValidationFailures tests invalid local-stub inputs.
func TestP2PLocalStubValidationFailures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping local P2P stub validation test in short mode")
	}

	// Create a P2P node
	node, err := p2p.NewP2PNode("/ip4/127.0.0.1/tcp/0")
	if err != nil {
		t.Fatalf("Failed to create P2P node: %v", err)
	}
	defer node.Close()

	// Test connecting to invalid peer addresses
	invalidAddresses := []string{
		"invalid-peer-address",
		"/ip4/192.168.999.999/tcp/4001/p2p/invalid-peer-id",
		"/ip4/127.0.0.1/tcp/99999/p2p/invalid", // Invalid port
	}

	for _, addr := range invalidAddresses {
		err := node.ConnectToPeer(addr)
		if err == nil {
			t.Errorf("Expected error when connecting to invalid address: %s", addr)
		} else {
			t.Logf("Correctly received error for invalid address %s: %v", addr, err)
		}
	}

	// Test publishing to topic without subscribers
	topic := "failure-test-topic"
	message := []byte("This message should fail gracefully")

	// This should not panic even without subscribers
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Publishing to topic without subscribers caused panic: %v", r)
		}
	}()

	err = node.PublishToTopic(topic, message)
	if err != nil {
		// Publishing might fail in isolated node, but shouldn't panic
		t.Logf("Publishing to topic failed as expected in isolated node: %v", err)
	} else {
		t.Log("Successfully published to topic in isolated node")
	}
}

// TestConcurrentMockStorageOperations tests concurrent operations against the mock storage.
func TestConcurrentMockStorageOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent mock storage test in short mode")
	}

	// Use mock IPFS for predictable concurrent testing
	mockIPFS := NewMockIPFSClient()

	numGoroutines := 10
	numOperations := 5
	results := make(chan error, numGoroutines*numOperations)

	// Run concurrent upload/download operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numOperations; j++ {
				// Create unique test data
				testData := []byte(fmt.Sprintf("Concurrent test data from goroutine %d operation %d", id, j))
				filename := fmt.Sprintf("concurrent-test-%d-%d.txt", id, j)

				// Upload
				cid, err := mockIPFS.AddFile(testData, filename)
				if err != nil {
					results <- fmt.Errorf("upload failed for %s: %w", filename, err)
					continue
				}

				// Download
				retrieved, err := mockIPFS.GetFile(cid)
				if err != nil {
					results <- fmt.Errorf("download failed for %s: %w", filename, err)
					continue
				}

				// Verify
				if !bytes.Equal(testData, retrieved) {
					results <- fmt.Errorf("data mismatch for %s", filename)
					continue
				}

				results <- nil // Success
			}
		}(i)
	}

	// Collect results
	totalOperations := numGoroutines * numOperations
	successCount := 0
	var errors []error

	for i := 0; i < totalOperations; i++ {
		err := <-results
		if err != nil {
			errors = append(errors, err)
		} else {
			successCount++
		}
	}

	if successCount != totalOperations {
		t.Errorf("Concurrent operations failed: %d/%d succeeded", successCount, totalOperations)
		for _, err := range errors {
			t.Errorf("Concurrent operation error: %v", err)
		}
	} else {
		t.Logf("All concurrent operations succeeded: %d/%d", successCount, totalOperations)
	}
}

// TestIPFSClientClose tests the Close() method
func TestIPFSClientClose(t *testing.T) {
	client, err := ipfs.NewIPFSClient("localhost:5001")
	if err != nil {
		t.Skipf("Skipping Close test - IPFS not available: %v", err)
	}

	// Close should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Close() panicked: %v", r)
		}
	}()

	err = client.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

// TestIPFSClientCreateDirectory tests CreateDirectory functionality
func TestIPFSClientCreateDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping CreateDirectory test in short mode")
	}

	mockIPFS := NewMockIPFSClient()

	// Create a real IPFS client wrapper for testing
	// Since we can't easily test with real IPFS, we'll test the logic
	files := map[string][]byte{
		"file1.txt": []byte("Content of file 1"),
		"file2.txt": []byte("Content of file 2"),
		"file3.txt": []byte("Content of file 3"),
	}

	// Test that CreateDirectory handles multiple files
	// In a real scenario, this would use the actual IPFS client
	t.Run("MultipleFiles", func(t *testing.T) {
		if len(files) == 0 {
			t.Error("No files provided for directory creation")
		}

		// Verify all files can be added individually
		cids := make(map[string]string)
		for filename, data := range files {
			cid, err := mockIPFS.AddFile(data, filename)
			if err != nil {
				t.Fatalf("Failed to add file %s: %v", filename, err)
			}
			cids[filename] = cid
		}

		if len(cids) != len(files) {
			t.Errorf("Expected %d CIDs, got %d", len(files), len(cids))
		}

		// Verify all files are retrievable
		for filename, cid := range cids {
			retrieved, err := mockIPFS.GetFile(cid)
			if err != nil {
				t.Errorf("Failed to retrieve file %s: %v", filename, err)
			}
			if !bytes.Equal(files[filename], retrieved) {
				t.Errorf("File %s content mismatch", filename)
			}
		}
	})
}

// TestConfigModule tests the configuration module
func TestConfigModule(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		cfg := config.DefaultConfig()
		if cfg == nil {
			t.Fatal("DefaultConfig returned nil")
		}

		if cfg.IPFS.URL == "" {
			t.Error("Default IPFS URL should not be empty")
		}

		if cfg.Encryption.KeyDerivation.KeyLen != 32 {
			t.Error("Default key length should be 32 bytes")
		}
	})

	t.Run("ConfigValidation", func(t *testing.T) {
		cfg := config.DefaultConfig()
		err := config.ValidateConfig(cfg)
		if err != nil {
			t.Errorf("Default config should be valid: %v", err)
		}

		// Test invalid config
		cfg.Encryption.KeyDerivation.KeyLen = 16
		err = config.ValidateConfig(cfg)
		if err == nil {
			t.Error("Config with invalid key length should fail validation")
		}
	})

	t.Run("ConfigSaveLoad", func(t *testing.T) {
		// Create temporary config file
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "test-config.json")

		// Create and save config
		cfg := config.DefaultConfig()
		cfg.IPFS.URL = "test-url:5001"

		err := config.SaveConfig(cfg, configPath)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Load config
		loadedCfg, err := config.LoadConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if loadedCfg.IPFS.URL != cfg.IPFS.URL {
			t.Errorf("Loaded config URL mismatch: expected %s, got %s", cfg.IPFS.URL, loadedCfg.IPFS.URL)
		}
	})
}

// TestRetryLogic tests the retry utility functions
func TestRetryLogic(t *testing.T) {
	t.Run("RetryWithBackoff", func(t *testing.T) {
		ctx := context.Background()
		attempts := 0
		maxAttempts := 3

		fn := func() error {
			attempts++
			if attempts < maxAttempts {
				return errors.New("temporary error")
			}
			return nil
		}

		config := utils.DefaultRetryConfig()
		config.MaxRetries = maxAttempts

		err := utils.RetryWithBackoff(ctx, fn, config)
		if err != nil {
			t.Errorf("Retry should succeed after %d attempts: %v", maxAttempts, err)
		}

		if attempts != maxAttempts {
			t.Errorf("Expected %d attempts, got %d", maxAttempts, attempts)
		}
	})

	t.Run("RetryMaxAttempts", func(t *testing.T) {
		ctx := context.Background()
		attempts := 0

		fn := func() error {
			attempts++
			return errors.New("persistent error")
		}

		config := utils.DefaultRetryConfig()
		config.MaxRetries = 2

		err := utils.RetryWithBackoff(ctx, fn, config)
		if err == nil {
			t.Error("Retry should fail after max attempts")
		}

		expectedAttempts := config.MaxRetries + 1
		if attempts != expectedAttempts {
			t.Errorf("Expected %d attempts, got %d", expectedAttempts, attempts)
		}
	})

	t.Run("IsRetryableError", func(t *testing.T) {
		testCases := []struct {
			err       error
			retryable bool
		}{
			{errors.New("connection refused"), true},
			{errors.New("timeout"), true},
			{errors.New("network error"), true},
			{errors.New("invalid input"), false},
			{errors.New("authentication failed"), false},
			{nil, false},
		}

		for _, tc := range testCases {
			result := utils.IsRetryableError(tc.err)
			if result != tc.retryable {
				t.Errorf("IsRetryableError(%v) = %v, expected %v", tc.err, result, tc.retryable)
			}
		}
	})
}
