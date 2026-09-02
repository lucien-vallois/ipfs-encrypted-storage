package tests

import (
	"os"
	"strings"
	"testing"

	"ipfs-encrypted-storage/src/ipfs"
)

func TestIPFSClientInitialization(t *testing.T) {
	// Test client creation with valid URL
	client, err := ipfs.NewIPFSClient("localhost:5001")
	if err != nil {
		t.Skipf("Skipping test - IPFS not available: %v", err)
	}
	if client == nil {
		t.Error("Failed to create IPFS client")
	}

	// Test Close method
	err = client.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}

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

	// Close again should be safe
	err = client.Close()
	if err != nil {
		t.Errorf("Second Close() returned error: %v", err)
	}
}

func TestCIDValidation(t *testing.T) {
	validCIDs := []string{
		"QmYjtig7VJQ6XsnUjqqJvj7QaMcCAwtrgNdahSiFofrE79",
		"bafybeigdyrzt5sfp7udm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi",
		"QmYwAPJzv5CZsnAzt8auVZRnGmTZcyuT3HCeJZ2R6V6q7E",
	}

	for _, cid := range validCIDs {
		if !ipfs.IsValidCID(cid) {
			t.Errorf("CID %s should be valid", cid)
		}
	}

	invalidCIDs := []string{
		"",
		"invalid-cid",
		"QmYjtig7VJQ6XsnUjqqJvj7QaMcCAwtrgNdahSiFofrE7",   // too short
		"QmYjtig7VJQ6XsnUjqqJvj7QaMcCAwtrgNdahSiFofrE790", // invalid base58
	}

	for _, cid := range invalidCIDs {
		if ipfs.IsValidCID(cid) {
			t.Errorf("CID %s should be invalid", cid)
		}
	}
}

func TestCIDParsing(t *testing.T) {
	testCID := "QmYjtig7VJQ6XsnUjqqJvj7QaMcCAwtrgNdahSiFofrE79"

	cid, err := ipfs.ParseCID(testCID)
	if err != nil {
		t.Fatalf("Failed to parse valid CID: %v", err)
	}

	if cid.String() != testCID {
		t.Errorf("Parsed CID doesn't match. Got: %s, Want: %s", cid.String(), testCID)
	}

	// Test invalid CID
	_, err = ipfs.ParseCID("invalid-cid")
	if err == nil {
		t.Error("Should fail to parse invalid CID")
	}
}

// Note: These tests require a running IPFS daemon
// They are marked as integration tests and should be run separately
func TestIPFSClientIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ipfsURL := os.Getenv("IPFS_URL")
	if ipfsURL == "" {
		t.Skip("IPFS_URL is not configured")
	}

	client, err := ipfs.NewIPFSClient(ipfsURL)
	if err != nil {
		t.Fatalf("Failed to create IPFS client: %v", err)
	}

	// Test health check (requires running IPFS daemon)
	err = client.HealthCheck()
	if err != nil {
		t.Fatalf("IPFS_URL is configured but its health check failed: %v", err)
	}
}

func BenchmarkCIDValidation(b *testing.B) {
	validCID := "QmYjtig7VJQ6XsnUjqqJvj7QaMcCAwtrgNdahSiFofrE79"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = ipfs.IsValidCID(validCID)
	}
}

func BenchmarkCIDParsing(b *testing.B) {
	testCID := "QmYjtig7VJQ6XsnUjqqJvj7QaMcCAwtrgNdahSiFofrE79"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = ipfs.ParseCID(testCID)
	}
}

// Mock IPFS client for testing without daemon
type MockIPFSClient struct {
	storedData map[string][]byte
}

func NewMockIPFSClient() *MockIPFSClient {
	return &MockIPFSClient{
		storedData: make(map[string][]byte),
	}
}

func (m *MockIPFSClient) AddFile(data []byte, filename string) (string, error) {
	cid := "QmMockCID" + strings.ReplaceAll(filename, ".", "")
	m.storedData[cid] = data
	return cid, nil
}

func (m *MockIPFSClient) GetFile(cid string) ([]byte, error) {
	data, exists := m.storedData[cid]
	if !exists {
		return nil, &IPFSError{Message: "file not found"}
	}
	return data, nil
}

type IPFSError struct {
	Message string
}

func (e *IPFSError) Error() string {
	return e.Message
}

// Test with mock client
func TestMockIPFSOperations(t *testing.T) {
	mockClient := NewMockIPFSClient()

	testData := []byte("Test data for mock IPFS")
	filename := "test.txt"

	// Test upload
	cid, err := mockClient.AddFile(testData, filename)
	if err != nil {
		t.Fatalf("Mock upload failed: %v", err)
	}

	if cid == "" {
		t.Error("CID should not be empty")
	}

	// Test download
	retrievedData, err := mockClient.GetFile(cid)
	if err != nil {
		t.Fatalf("Mock download failed: %v", err)
	}

	if string(retrievedData) != string(testData) {
		t.Errorf("Retrieved data doesn't match. Got: %s, Want: %s", retrievedData, testData)
	}
}
