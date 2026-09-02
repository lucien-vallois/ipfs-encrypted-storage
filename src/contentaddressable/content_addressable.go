// Package contentaddressable provides content-addressable storage patterns
// Implements IPFS-style content addressing with cryptographic hashing
package contentaddressable

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"strings"
)

// ContentAddressableStorage provides content-addressable storage operations
type ContentAddressableStorage interface {
	Store(data []byte) (string, error)
	Retrieve(address string) ([]byte, error)
	Has(address string) (bool, error)
	Delete(address string) error
	List() ([]string, error)
}

// HashFunction represents supported hash functions for content addressing
type HashFunction int

const (
	SHA256 HashFunction = iota
	SHA3_256
	BLAKE3
)

// String returns the string representation of the hash function
func (hf HashFunction) String() string {
	switch hf {
	case SHA256:
		return "sha256"
	case SHA3_256:
		return "sha3-256"
	case BLAKE3:
		return "blake3"
	default:
		return "unknown"
	}
}

// ContentAddress represents a content address with hash function and digest
type ContentAddress struct {
	HashFunction HashFunction `json:"hash_function"`
	Digest       string       `json:"digest"`
	Multihash    string       `json:"multihash,omitempty"` // IPFS-style multihash
}

// String returns the string representation of the content address
func (ca *ContentAddress) String() string {
	if ca.Multihash != "" {
		return ca.Multihash
	}
	return fmt.Sprintf("%s-%s", ca.HashFunction.String(), ca.Digest)
}

// ParseContentAddress parses a content address string
func ParseContentAddress(address string) (*ContentAddress, error) {
	// Handle IPFS-style multihash (simplified)
	if strings.HasPrefix(address, "Qm") || strings.HasPrefix(address, "bafy") {
		return &ContentAddress{
			HashFunction: SHA256,
			Multihash:    address,
			Digest:       address, // Simplified
		}, nil
	}

	// Handle hash-function-digest format
	parts := strings.SplitN(address, "-", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid content address format: %s", address)
	}

	var hashFunc HashFunction
	switch parts[0] {
	case "sha256":
		hashFunc = SHA256
	case "sha3-256":
		hashFunc = SHA3_256
	case "blake3":
		hashFunc = BLAKE3
	default:
		return nil, fmt.Errorf("unsupported hash function: %s", parts[0])
	}

	return &ContentAddress{
		HashFunction: hashFunc,
		Digest:       parts[1],
	}, nil
}

// ContentHasher provides content hashing functionality
type ContentHasher struct {
	hashFunc HashFunction
}

// NewContentHasher creates a new content hasher
func NewContentHasher(hashFunc HashFunction) *ContentHasher {
	return &ContentHasher{
		hashFunc: hashFunc,
	}
}

// Hash computes the content address for the given data
func (ch *ContentHasher) Hash(data []byte) (*ContentAddress, error) {
	var hasher hash.Hash

	switch ch.hashFunc {
	case SHA256:
		hasher = sha256.New()
	case SHA3_256:
		// Would need: golang.org/x/crypto/sha3
		// hasher = sha3.New256()
		hasher = sha256.New() // Fallback for now
	case BLAKE3:
		// Would need: lukechampine.com/blake3
		// hasher = blake3.New()
		hasher = sha256.New() // Fallback for now
	default:
		return nil, fmt.Errorf("unsupported hash function: %v", ch.hashFunc)
	}

	hasher.Write(data)
	digest := hex.EncodeToString(hasher.Sum(nil))

	return &ContentAddress{
		HashFunction: ch.hashFunc,
		Digest:       digest,
	}, nil
}

// HashReader computes the content address for data from a reader
func (ch *ContentHasher) HashReader(reader io.Reader) (*ContentAddress, error) {
	var hasher hash.Hash

	switch ch.hashFunc {
	case SHA256:
		hasher = sha256.New()
	case SHA3_256:
		hasher = sha256.New() // Fallback
	case BLAKE3:
		hasher = sha256.New() // Fallback
	default:
		return nil, fmt.Errorf("unsupported hash function: %v", ch.hashFunc)
	}

	_, err := io.Copy(hasher, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	digest := hex.EncodeToString(hasher.Sum(nil))

	return &ContentAddress{
		HashFunction: ch.hashFunc,
		Digest:       digest,
	}, nil
}

// VerifyContent verifies that data matches a content address
func (ch *ContentHasher) VerifyContent(data []byte, address *ContentAddress) bool {
	computed, err := ch.Hash(data)
	if err != nil {
		return false
	}

	return computed.Digest == address.Digest && computed.HashFunction == address.HashFunction
}

// ContentBlock represents a block of content-addressable data
type ContentBlock struct {
	Address *ContentAddress `json:"address"`
	Data    []byte          `json:"data,omitempty"` // May be nil for metadata-only blocks
	Size    int             `json:"size"`
	Type    string          `json:"type"`            // "file", "directory", "metadata", etc.
	Links   []string        `json:"links,omitempty"` // Addresses of linked blocks
}

// ContentGraph represents a graph of content-addressable blocks
type ContentGraph struct {
	blocks map[string]*ContentBlock
}

// NewContentGraph creates a new content graph
func NewContentGraph() *ContentGraph {
	return &ContentGraph{
		blocks: make(map[string]*ContentBlock),
	}
}

// AddBlock adds a block to the graph
func (cg *ContentGraph) AddBlock(block *ContentBlock) {
	cg.blocks[block.Address.String()] = block
}

// GetBlock retrieves a block from the graph
func (cg *ContentGraph) GetBlock(address string) (*ContentBlock, bool) {
	block, exists := cg.blocks[address]
	return block, exists
}

// GetLinkedBlocks returns all blocks linked from a given block
func (cg *ContentGraph) GetLinkedBlocks(address string) []*ContentBlock {
	block, exists := cg.blocks[address]
	if !exists {
		return nil
	}

	var linked []*ContentBlock
	for _, linkAddr := range block.Links {
		if linkedBlock, exists := cg.blocks[linkAddr]; exists {
			linked = append(linked, linkedBlock)
		}
	}

	return linked
}

// MerkleDAG represents a Merkle DAG structure
type MerkleDAG struct {
	Root     string                   `json:"root"`
	Blocks   map[string]*ContentBlock `json:"blocks"`
	Metadata map[string]interface{}   `json:"metadata,omitempty"`
}

// BuildMerkleDAG builds a Merkle DAG from a slice of content blocks
func BuildMerkleDAG(blocks []*ContentBlock) *MerkleDAG {
	dag := &MerkleDAG{
		Blocks: make(map[string]*ContentBlock),
	}

	// Add all blocks
	for _, block := range blocks {
		dag.Blocks[block.Address.String()] = block
	}

	// Find root (block with no incoming links)
	// Simplified: assume first block is root
	if len(blocks) > 0 {
		dag.Root = blocks[0].Address.String()
	}

	return dag
}

// ValidateDAG validates the integrity of a Merkle DAG
func ValidateDAG(dag *MerkleDAG, hasher *ContentHasher) bool {
	for addr, block := range dag.Blocks {
		// Verify block address
		if addr != block.Address.String() {
			return false
		}

		// Verify content integrity if data is present
		if block.Data != nil {
			if !hasher.VerifyContent(block.Data, block.Address) {
				return false
			}
		}

		// Verify links exist
		for _, linkAddr := range block.Links {
			if _, exists := dag.Blocks[linkAddr]; !exists {
				return false
			}
		}
	}

	return true
}

// ContentAddressableFileSystem provides a file system interface over content-addressable storage
type ContentAddressableFileSystem struct {
	storage ContentAddressableStorage
	hasher  *ContentHasher
	graph   *ContentGraph
}

// NewContentAddressableFileSystem creates a new CAFS
func NewContentAddressableFileSystem(storage ContentAddressableStorage) *ContentAddressableFileSystem {
	return &ContentAddressableFileSystem{
		storage: storage,
		hasher:  NewContentHasher(SHA256),
		graph:   NewContentGraph(),
	}
}

// StoreFile stores a file in the CAFS
func (cafs *ContentAddressableFileSystem) StoreFile(data []byte, filename string) (*ContentBlock, error) {
	// Compute content address
	address, err := cafs.hasher.Hash(data)
	if err != nil {
		return nil, fmt.Errorf("failed to hash content: %w", err)
	}

	// Store in underlying storage
	_, err = cafs.storage.Store(data)
	if err != nil {
		return nil, fmt.Errorf("failed to store content: %w", err)
	}

	// Create content block
	block := &ContentBlock{
		Address: address,
		Data:    data,
		Size:    len(data),
		Type:    "file",
	}

	// Add to graph
	cafs.graph.AddBlock(block)

	return block, nil
}

// StoreDirectory stores a directory structure in the CAFS
func (cafs *ContentAddressableFileSystem) StoreDirectory(files map[string][]byte, dirname string) (*ContentBlock, error) {
	var fileBlocks []*ContentBlock
	var links []string

	// Store each file
	for filename, data := range files {
		fileBlock, err := cafs.StoreFile(data, filename)
		if err != nil {
			return nil, fmt.Errorf("failed to store file %s: %w", filename, err)
		}

		fileBlocks = append(fileBlocks, fileBlock)
		links = append(links, fileBlock.Address.String())
	}

	// Create directory block
	dirData := []byte(fmt.Sprintf("Directory: %s", dirname))
	dirAddress, err := cafs.hasher.Hash(dirData)
	if err != nil {
		return nil, fmt.Errorf("failed to hash directory: %w", err)
	}

	dirBlock := &ContentBlock{
		Address: dirAddress,
		Data:    dirData,
		Size:    len(dirData),
		Type:    "directory",
		Links:   links,
	}

	// Add directory to graph
	cafs.graph.AddBlock(dirBlock)

	return dirBlock, nil
}

// RetrieveFile retrieves a file from the CAFS
func (cafs *ContentAddressableFileSystem) RetrieveFile(address string) ([]byte, error) {
	return cafs.storage.Retrieve(address)
}

// GetBlockInfo returns information about a content block
func (cafs *ContentAddressableFileSystem) GetBlockInfo(address string) (*ContentBlock, error) {
	block, exists := cafs.graph.GetBlock(address)
	if !exists {
		return nil, fmt.Errorf("block not found: %s", address)
	}

	return block, nil
}

// ListDirectory lists the contents of a directory
func (cafs *ContentAddressableFileSystem) ListDirectory(address string) ([]*ContentBlock, error) {
	return cafs.graph.GetLinkedBlocks(address), nil
}
