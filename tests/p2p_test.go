package tests

import (
	"bytes"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"
	"ipfs-encrypted-storage/src/p2p"
)

const testPeerID = "QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ"

func TestP2PNodeCreation(t *testing.T) {
	// Test P2P node creation (without actually starting the node)
	// This tests the configuration and setup logic

	listenAddr := "/ip4/127.0.0.1/tcp/0" // Use localhost and random port

	node, err := p2p.NewP2PNode(listenAddr)
	if err != nil {
		t.Fatalf("Failed to create P2P node: %v", err)
	}
	defer node.Close()

	// Verify node properties
	if node.GetID() == "" {
		t.Error("Node ID should not be empty")
	}

	addresses := node.GetAddresses()
	if len(addresses) == 0 {
		t.Error("Node should have at least one listen address")
	}

	// Test peer ID validation
	peerID := node.GetID()
	_, err = peer.Decode(peerID.String())
	if err != nil {
		t.Errorf("Invalid peer ID format: %v", err)
	}
}

func TestPeerInfo(t *testing.T) {
	peerID, err := peer.Decode(testPeerID)
	if err != nil {
		t.Fatalf("Failed to decode test peer ID: %v", err)
	}

	peerInfo := &p2p.PeerInfo{
		ID:           peerID,
		LastSeen:     time.Now(),
		Capabilities: []string{"storage", "relay"},
	}

	if peerInfo.ID != peerID {
		t.Error("Peer ID not set correctly")
	}

	if len(peerInfo.Capabilities) != 2 {
		t.Error("Capabilities not set correctly")
	}

	if peerInfo.Capabilities[0] != "storage" {
		t.Error("First capability should be storage")
	}
}

func TestBootstrapPeersList(t *testing.T) {
	peers := p2p.DefaultBootstrapPeers

	if len(peers) == 0 {
		t.Error("Should have default bootstrap peers")
	}

	// Verify that bootstrap peers are valid multiaddrs
	for _, addr := range peers {
		if addr == "" {
			t.Error("Bootstrap peer address should not be empty")
		}

		// Should contain IPFS protocol identifier
		if !contains(addr, "/ip4/") && !contains(addr, "/ip6/") {
			t.Errorf("Invalid bootstrap peer format: %s", addr)
		}
	}
}

func TestProtocolConstants(t *testing.T) {
	// Test that protocol constants are properly defined
	if p2p.ProtocolStorage == "" {
		t.Error("ProtocolStorage should not be empty")
	}

	if p2p.ProtocolFileRequest == "" {
		t.Error("ProtocolFileRequest should not be empty")
	}

	if p2p.ProtocolFileTransfer == "" {
		t.Error("ProtocolFileTransfer should not be empty")
	}

	// Test that protocols contain expected strings
	if !contains(string(p2p.ProtocolStorage), "encrypted-storage") {
		t.Error("ProtocolStorage should contain encrypted-storage")
	}

	if !contains(string(p2p.ProtocolFileRequest), "file-request") {
		t.Error("ProtocolFileRequest should contain file-request")
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr ||
		len(s) > len(substr) && s[len(s)-len(substr):] == substr ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}()
}

// Mock P2P node for testing
type MockP2PNode struct {
	id       peer.ID
	peers    map[peer.ID]*p2p.PeerInfo
	messages map[string][]byte
}

func NewMockP2PNode() *MockP2PNode {
	mockID, err := peer.Decode(testPeerID)
	if err != nil {
		panic(err)
	}
	return &MockP2PNode{
		id:       mockID,
		peers:    make(map[peer.ID]*p2p.PeerInfo),
		messages: make(map[string][]byte),
	}
}

func (m *MockP2PNode) GetID() peer.ID {
	return m.id
}

func (m *MockP2PNode) GetAddresses() []string {
	return []string{"/ip4/127.0.0.1/tcp/4001/p2p/" + m.id.String()}
}

func (m *MockP2PNode) ListPeers() []*p2p.PeerInfo {
	var peers []*p2p.PeerInfo
	for _, info := range m.peers {
		peers = append(peers, info)
	}
	return peers
}

func TestMockP2POperations(t *testing.T) {
	node := NewMockP2PNode()

	// Test basic properties
	if node.GetID().String() == "" {
		t.Error("Mock node should have an ID")
	}

	addresses := node.GetAddresses()
	if len(addresses) == 0 {
		t.Error("Mock node should have addresses")
	}

	// Test peer listing (should be empty initially)
	peers := node.ListPeers()
	if len(peers) != 0 {
		t.Error("New mock node should have no peers")
	}
}

// Benchmark P2P operations (mock)
func BenchmarkPeerInfoCreation(b *testing.B) {
	peerID, _ := peer.Decode(testPeerID)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = &p2p.PeerInfo{
			ID:           peerID,
			LastSeen:     time.Now(),
			Capabilities: []string{"storage", "relay"},
		}
	}
}

// TestLocalStoreValueOperations verifies node-local key/value storage.
func TestLocalStoreValueOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping local storage test in short mode")
	}

	listenAddr := "/ip4/127.0.0.1/tcp/0"

	node, err := p2p.NewP2PNode(listenAddr)
	if err != nil {
		t.Fatalf("Failed to create local P2P stub: %v", err)
	}
	defer node.Close()

	testKey := "test-key"
	testValue := []byte("test-value")

	if err := node.StoreValue(testKey, testValue); err != nil {
		t.Fatalf("StoreValue failed: %v", err)
	}
	got, err := node.GetValue(testKey)
	if err != nil {
		t.Fatalf("GetValue failed: %v", err)
	}
	if !bytes.Equal(got, testValue) {
		t.Fatalf("GetValue = %q, want %q", got, testValue)
	}
}
