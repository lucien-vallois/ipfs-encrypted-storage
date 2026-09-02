// Package p2p provides a lightweight in-memory P2P stub.
package p2p

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	libp2pcrypto "github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

const (
	ProtocolStorage      = "encrypted-storage"
	ProtocolFileRequest  = "file-request"
	ProtocolFileTransfer = "file-transfer"
)

var DefaultBootstrapPeers = []string{
	"/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ",
}

// MessageHandler handles a message from a peer.
type MessageHandler func(string, []byte) error

// PeerInfo holds information about a peer.
type PeerInfo struct {
	ID           peer.ID
	Addresses    []string
	LastSeen     time.Time
	Capabilities []string
}

// Subscription is a cancellable topic subscription.
type Subscription struct {
	once   sync.Once
	cancel func()
}

func (s *Subscription) Cancel() {
	if s != nil && s.cancel != nil {
		s.once.Do(s.cancel)
	}
}

// P2PNode is an in-memory stub used by the CLI and tests.
type P2PNode struct {
	id              peer.ID
	addresses       []string
	mu              sync.RWMutex
	storage         map[string][]byte
	messageHandlers map[string]MessageHandler
	subscriptions   map[string]map[*Subscription]MessageHandler
	closed          bool
}

// NewP2PNode creates an in-memory node with a valid libp2p peer ID.
func NewP2PNode(listenAddr string) (*P2PNode, error) {
	if _, err := multiaddr.NewMultiaddr(listenAddr); err != nil {
		return nil, fmt.Errorf("invalid listen address: %w", err)
	}

	_, publicKey, err := libp2pcrypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate peer key: %w", err)
	}

	id, err := peer.IDFromPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("derive peer ID: %w", err)
	}

	return &P2PNode{
		id:              id,
		addresses:       []string{listenAddr},
		storage:         make(map[string][]byte),
		messageHandlers: make(map[string]MessageHandler),
		subscriptions:   make(map[string]map[*Subscription]MessageHandler),
	}, nil
}

func (n *P2PNode) GetID() peer.ID { return n.id }

func (n *P2PNode) GetAddresses() []string {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return append([]string(nil), n.addresses...)
}

func (n *P2PNode) ListPeers() []*PeerInfo { return []*PeerInfo{} }

// Bootstrap validates the configured bootstrap peer addresses.
func (n *P2PNode) Bootstrap(peers []string) error {
	for _, peerAddr := range peers {
		if err := n.ConnectToPeer(peerAddr); err != nil {
			return fmt.Errorf("bootstrap peer %q: %w", peerAddr, err)
		}
	}
	return nil
}

func (n *P2PNode) RegisterMessageHandler(protocol string, handler MessageHandler) {
	n.mu.Lock()
	defer n.mu.Unlock()
	if !n.closed && handler != nil {
		n.messageHandlers[protocol] = handler
	}
}

func (n *P2PNode) StoreValue(key string, value []byte) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	if n.closed {
		return fmt.Errorf("P2P node is closed")
	}
	n.storage[key] = append([]byte(nil), value...)
	return nil
}

func (n *P2PNode) GetValue(key string) ([]byte, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.closed {
		return nil, fmt.Errorf("P2P node is closed")
	}
	value, ok := n.storage[key]
	if !ok {
		return nil, fmt.Errorf("value not found for key %q", key)
	}
	return append([]byte(nil), value...), nil
}

func (n *P2PNode) PublishToTopic(topic string, data []byte) error {
	n.mu.RLock()
	if n.closed {
		n.mu.RUnlock()
		return fmt.Errorf("P2P node is closed")
	}
	handlers := make([]MessageHandler, 0, len(n.subscriptions[topic]))
	for _, handler := range n.subscriptions[topic] {
		handlers = append(handlers, handler)
	}
	publisherID := n.id.String()
	n.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(publisherID, append([]byte(nil), data...)); err != nil {
			return err
		}
	}
	return nil
}

func (n *P2PNode) SubscribeToTopic(topic string, handler MessageHandler) (*Subscription, error) {
	if topic == "" || handler == nil {
		return nil, fmt.Errorf("topic and handler are required")
	}

	subscription := &Subscription{}
	n.mu.Lock()
	if n.closed {
		n.mu.Unlock()
		return nil, fmt.Errorf("P2P node is closed")
	}
	if n.subscriptions[topic] == nil {
		n.subscriptions[topic] = make(map[*Subscription]MessageHandler)
	}
	n.subscriptions[topic][subscription] = handler
	n.mu.Unlock()

	subscription.cancel = func() {
		n.mu.Lock()
		delete(n.subscriptions[topic], subscription)
		if len(n.subscriptions[topic]) == 0 {
			delete(n.subscriptions, topic)
		}
		n.mu.Unlock()
	}
	return subscription, nil
}

func (n *P2PNode) ConnectToPeer(peerAddr string) error {
	addr, err := multiaddr.NewMultiaddr(peerAddr)
	if err != nil {
		return fmt.Errorf("invalid peer address: %w", err)
	}
	id, err := addr.ValueForProtocol(multiaddr.P_P2P)
	if err != nil {
		return fmt.Errorf("peer address has no peer ID: %w", err)
	}
	if _, err := peer.Decode(id); err != nil {
		return fmt.Errorf("invalid peer ID: %w", err)
	}
	return nil
}

func (n *P2PNode) GetNetworkStats() map[string]interface{} {
	return map[string]interface{}{
		"connected_peers":  0,
		"listen_addresses": n.GetAddresses(),
		"protocols":        []string{ProtocolStorage},
	}
}

func (n *P2PNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.closed = true
	clear(n.messageHandlers)
	clear(n.subscriptions)
	return nil
}
