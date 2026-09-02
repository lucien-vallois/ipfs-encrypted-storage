// Package models provides request and response structures for the REST API
package models

import (
	"time"
)

// FileUploadRequest represents a file upload request
type FileUploadRequest struct {
	Password    string                 `json:"password,omitempty" binding:"omitempty,min=8"`
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// FileUploadResponse represents a file upload response
type FileUploadResponse struct {
	CID                string                 `json:"cid"`
	Filename           string                 `json:"filename"`
	Size               int64                  `json:"size"`
	Description        string                 `json:"description,omitempty"`
	Encrypted          bool                   `json:"encrypted"`
	MetadataFile       string                 `json:"metadata_file,omitempty"`
	UploadedAt         int64                  `json:"uploaded_at"`
	EncryptionMetadata map[string]interface{} `json:"encryption_metadata,omitempty"`
}

// FileDownloadResponse represents file information for download
type FileDownloadResponse struct {
	CID         string `json:"cid"`
	Filename    string `json:"filename"`
	Size        int64  `json:"size"`
	Encrypted   bool   `json:"encrypted"`
	UploadedAt  int64  `json:"uploaded_at"`
	Description string `json:"description,omitempty"`
}

// FileListResponse represents a list of files
type FileListResponse struct {
	Files []FileInfo `json:"files"`
	Count int        `json:"count"`
}

// FileInfo represents basic file information
type FileInfo struct {
	CID       string `json:"cid"`
	Filename  string `json:"filename"`
	Size      int64  `json:"size"`
	PinnedAt  int64  `json:"pinned_at"`
	Encrypted bool   `json:"encrypted"`
}

// FileDeleteResponse represents a file deletion response
type FileDeleteResponse struct {
	CID     string `json:"cid"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// P2PConnectRequest represents a P2P connection request
type P2PConnectRequest struct {
	PeerAddress string `json:"peer_address" binding:"required"`
	Timeout     int    `json:"timeout,omitempty"`
}

// P2PConnectResponse represents a P2P connection response
type P2PConnectResponse struct {
	Success bool    `json:"success"`
	PeerID  string  `json:"peer_id,omitempty"`
	Address string  `json:"address,omitempty"`
	Latency float64 `json:"latency,omitempty"`
	Error   string  `json:"error,omitempty"`
}

// P2PPeersResponse represents connected P2P peers
type P2PPeersResponse struct {
	Peers []PeerInfo `json:"peers"`
	Count int        `json:"count"`
}

// PeerInfo represents information about a connected peer
type PeerInfo struct {
	PeerID      string   `json:"peer_id"`
	Address     string   `json:"address"`
	ConnectedAt int64    `json:"connected_at"`
	Protocols   []string `json:"protocols"`
}

// HealthResponse represents system health status
type HealthResponse struct {
	Status        string    `json:"status"` // "healthy", "degraded", "unhealthy"
	IPFSConnected bool      `json:"ipfs_connected"`
	P2PPeers      int       `json:"p2p_peers"`
	Uptime        float64   `json:"uptime"`
	Version       string    `json:"version"`
	Timestamp     time.Time `json:"timestamp"`
}

// MetricsResponse represents system metrics
type MetricsResponse struct {
	FilesUploaded      int64   `json:"files_uploaded"`
	TotalBytesUploaded int64   `json:"total_bytes_uploaded"`
	ActiveConnections  int     `json:"active_connections"`
	MemoryUsage        float64 `json:"memory_usage"`
	CPUUsage           float64 `json:"cpu_usage"`
	Uptime             float64 `json:"uptime"`
	RequestCount       int64   `json:"request_count"`
	ErrorCount         int64   `json:"error_count"`
}

// APIKeyAuth represents API key authentication
type APIKeyAuth struct {
	APIKey string `header:"X-API-Key" binding:"required"`
}

// PasswordAuth represents password-based authentication for downloads
type PasswordAuth struct {
	Password string `json:"password" binding:"required,min=8"`
}
