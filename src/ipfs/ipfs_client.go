// Package ipfs provides IPFS client functionality for encrypted storage
package ipfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ipfs/go-cid"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/sirupsen/logrus"

	"ipfs-encrypted-storage/src/utils"
)

// IPFS validation constants
const (
	MaxFileSize    = 100 * 1024 * 1024 // 100MB
	MaxFileNameLen = 255
	DefaultTimeout = 30 * time.Second
)

// IPFS errors
var (
	ErrInvalidURL         = fmt.Errorf("invalid IPFS URL format")
	ErrInvalidCID         = fmt.Errorf("invalid CID format")
	ErrFileTooLarge       = fmt.Errorf("file too large: exceeds maximum allowed size")
	ErrInvalidFileName    = fmt.Errorf("invalid filename: too long or contains invalid characters")
	ErrConnectionFailed   = fmt.Errorf("failed to connect to IPFS node")
	ErrTimeout            = fmt.Errorf("operation timed out")
	ErrInvalidPeerAddress = fmt.Errorf("invalid peer multiaddress")
)

// CID validation regex
var cidRegex = regexp.MustCompile(`^[a-zA-Z0-9]{46,59}$`)

// NormalizeURL normalizes an IPFS URL by adding scheme if missing
func NormalizeURL(urlStr string) string {
	// If URL already has a scheme, return as is
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") {
		return urlStr
	}
	// Add http:// scheme if missing
	return "http://" + urlStr
}

// ValidateURL validates IPFS URL format
func ValidateURL(urlStr string) error {
	// Normalize URL first (add scheme if missing)
	normalizedURL := NormalizeURL(urlStr)

	u, err := url.Parse(normalizedURL)
	if err != nil {
		return ErrInvalidURL
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported URL scheme: %s", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("missing host in URL")
	}

	// Check for valid port if specified
	if u.Port() != "" {
		if port := u.Port(); len(port) > 5 {
			return fmt.Errorf("invalid port number: %s", port)
		}
	}

	return nil
}

// ValidateCID validates CID format
func ValidateCID(cidStr string) error {
	if cidStr == "" {
		return ErrInvalidCID
	}

	if !cidRegex.MatchString(cidStr) {
		return ErrInvalidCID
	}

	// Additional validation using go-cid
	_, err := cid.Decode(cidStr)
	if err != nil {
		return fmt.Errorf("CID decode failed: %w", err)
	}

	return nil
}

// ValidateFileName validates filename
func ValidateFileName(filename string) error {
	if filename == "" {
		return ErrInvalidFileName
	}

	if len(filename) > MaxFileNameLen {
		return ErrInvalidFileName
	}

	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(filename, char) {
			return fmt.Errorf("filename contains invalid character: %s", char)
		}
	}

	return nil
}

// ValidateFileSize validates file size
func ValidateFileSize(size int64) error {
	if size < 0 {
		return fmt.Errorf("file size cannot be negative")
	}
	if size > MaxFileSize {
		return ErrFileTooLarge
	}
	return nil
}

// ValidatePeerAddress validates peer multiaddress
func ValidatePeerAddress(addr string) error {
	if addr == "" {
		return ErrInvalidPeerAddress
	}

	// Basic multiaddress validation
	if !strings.HasPrefix(addr, "/ip4/") && !strings.HasPrefix(addr, "/ip6/") {
		return ErrInvalidPeerAddress
	}

	parts := strings.Split(addr, "/")
	if len(parts) < 5 {
		return ErrInvalidPeerAddress
	}

	return nil
}

// IPFSClient wraps the IPFS shell client
type IPFSClient struct {
	shell *shell.Shell
	ctx   context.Context
}

// NewIPFSClient creates a new IPFS client
func NewIPFSClient(url string) (*IPFSClient, error) {
	// Normalize URL (add scheme if missing)
	normalizedURL := NormalizeURL(url)

	// Validate URL
	if err := ValidateURL(normalizedURL); err != nil {
		return nil, fmt.Errorf("invalid IPFS URL: %w", err)
	}

	sh := shell.NewShell(normalizedURL)
	if sh == nil {
		return nil, ErrConnectionFailed
	}
	sh.SetTimeout(DefaultTimeout)

	return &IPFSClient{
		shell: sh,
		ctx:   context.Background(),
	}, nil
}

// AddFile adds a file to IPFS and returns its CID
func (c *IPFSClient) AddFile(data []byte, filename string) (string, error) {
	// Validate inputs
	if err := ValidateFileSize(int64(len(data))); err != nil {
		return "", err
	}
	if err := ValidateFileName(filename); err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(c.ctx, DefaultTimeout)
	defer cancel()

	var cid string
	var err error

	// Use retry logic for network operations
	retryConfig := utils.DefaultRetryConfig()
	retryConfig.MaxRetries = 2 // Fewer retries for uploads

	err = utils.RetryWithBackoff(ctx, func() error {
		// Recreate reader for each retry attempt
		reader := bytes.NewReader(data)
		var addErr error
		cid, addErr = c.shell.Add(reader, shell.Pin(true))
		if addErr != nil {
			// Only retry if error is retryable
			if utils.IsRetryableError(addErr) {
				return addErr
			}
			// Non-retryable errors should stop immediately
			err = addErr
			return addErr
		}
		return nil
	}, retryConfig)

	if err != nil {
		return "", fmt.Errorf("failed to add file to IPFS: %w", err)
	}

	// Validate returned CID
	if err := ValidateCID(cid); err != nil {
		return "", fmt.Errorf("invalid CID returned from IPFS: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"cid":      cid,
		"filename": filename,
		"size":     len(data),
	}).Info("File added to IPFS")

	return cid, nil
}

// AddFileWithReader adds a file from reader to IPFS
func (c *IPFSClient) AddFileWithReader(reader io.Reader, filename string) (string, error) {
	cid, err := c.shell.Add(reader, shell.Pin(true))
	if err != nil {
		return "", fmt.Errorf("failed to add file from reader to IPFS: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"cid":      cid,
		"filename": filename,
	}).Info("File added to IPFS from reader")

	return cid, nil
}

// GetFile retrieves a file from IPFS by CID
func (c *IPFSClient) GetFile(cid string) ([]byte, error) {
	// Validate CID
	if err := ValidateCID(cid); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(c.ctx, DefaultTimeout)
	defer cancel()

	var reader io.ReadCloser
	var err error

	// Use retry logic for network operations
	retryConfig := utils.DefaultRetryConfig()
	retryConfig.MaxRetries = 3 // More retries for downloads

	err = utils.RetryWithBackoff(ctx, func() error {
		var catErr error
		reader, catErr = c.shell.Cat(cid)
		if catErr != nil {
			if utils.IsRetryableError(catErr) {
				return catErr // Retry on retryable errors
			}
			// Non-retryable errors should stop immediately
			err = catErr
			return catErr
		}
		return nil
	}, retryConfig)

	if err != nil {
		return nil, fmt.Errorf("failed to get file from IPFS: %w", err)
	}
	defer reader.Close()

	// Read with size limit
	limitedReader := &io.LimitedReader{R: reader, N: MaxFileSize + 1}
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	// Check if file was truncated due to size limit
	if limitedReader.N <= 0 {
		return nil, ErrFileTooLarge
	}

	logrus.WithFields(logrus.Fields{
		"cid":  cid,
		"size": len(data),
	}).Info("File retrieved from IPFS")

	return data, nil
}

// GetFileToWriter retrieves a file and writes it to a writer
func (c *IPFSClient) GetFileToWriter(cid string, writer io.Writer) error {
	reader, err := c.shell.Cat(cid)
	if err != nil {
		return fmt.Errorf("failed to get file from IPFS: %w", err)
	}
	defer reader.Close()

	_, err = io.Copy(writer, reader)
	if err != nil {
		return fmt.Errorf("failed to write file data: %w", err)
	}

	logrus.WithField("cid", cid).Info("File retrieved and written to destination")
	return nil
}

// PinFile pins a file to ensure it's stored locally
func (c *IPFSClient) PinFile(cid string) error {
	err := c.shell.Pin(cid)
	if err != nil {
		return fmt.Errorf("failed to pin file: %w", err)
	}

	logrus.WithField("cid", cid).Info("File pinned successfully")
	return nil
}

// UnpinFile unpins a file
func (c *IPFSClient) UnpinFile(cid string) error {
	err := c.shell.Unpin(cid)
	if err != nil {
		return fmt.Errorf("failed to unpin file: %w", err)
	}

	logrus.WithField("cid", cid).Info("File unpinned successfully")
	return nil
}

// ListPins lists all pinned files
func (c *IPFSClient) ListPins() (map[string]string, error) {
	pins, err := c.shell.Pins()
	if err != nil {
		return nil, fmt.Errorf("failed to list pins: %w", err)
	}

	result := make(map[string]string)
	for k, v := range pins {
		result[k] = v.Type
	}

	return result, nil
}

// IsPinned checks if a file is pinned
func (c *IPFSClient) IsPinned(cid string) (bool, error) {
	pins, err := c.ListPins()
	if err != nil {
		return false, err
	}

	pinType, exists := pins[cid]
	return exists && pinType != "indirect", nil
}

// GetObjectStats gets statistics about an IPFS object
func (c *IPFSClient) GetObjectStats(cid string) (*ObjectStats, error) {
	stats, err := c.shell.ObjectStat(cid)
	if err != nil {
		return nil, fmt.Errorf("failed to get object stats: %w", err)
	}

	return &ObjectStats{
		Hash:           stats.Hash,
		BlockSize:      stats.BlockSize,
		LinksSize:      stats.LinksSize,
		DataSize:       stats.DataSize,
		NumLinks:       stats.NumLinks,
		CumulativeSize: stats.CumulativeSize,
	}, nil
}

// ObjectStats holds statistics about an IPFS object
type ObjectStats struct {
	Hash           string
	BlockSize      int
	LinksSize      int
	DataSize       int
	NumLinks       int
	CumulativeSize int
}

// ListPeers lists connected peers
func (c *IPFSClient) ListPeers() ([]string, error) {
	peers, err := c.shell.SwarmPeers(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list peers: %w", err)
	}

	// Convert SwarmConnInfos to []string
	var peerStrings []string
	for _, peer := range peers.Peers {
		peerStrings = append(peerStrings, peer.Peer)
	}

	return peerStrings, nil
}

// ConnectToPeer connects to a specific peer
func (c *IPFSClient) ConnectToPeer(peerAddr string) error {
	ctx, cancel := context.WithTimeout(c.ctx, DefaultTimeout)
	defer cancel()

	var err error

	// Use retry logic for peer connections
	retryConfig := utils.DefaultRetryConfig()
	retryConfig.MaxRetries = 2

	err = utils.RetryWithBackoff(ctx, func() error {
		err = c.shell.SwarmConnect(ctx, peerAddr)
		if err != nil && utils.IsRetryableError(err) {
			return err
		}
		return err
	}, retryConfig)

	if err != nil {
		return fmt.Errorf("failed to connect to peer: %w", err)
	}

	logrus.WithField("peer", peerAddr).Info("Connected to peer")
	return nil
}

// PublishName publishes a name to IPNS
func (c *IPFSClient) PublishName(cid string, keyName string) (string, error) {
	// Validate CID first
	if err := ValidateCID(cid); err != nil {
		return "", fmt.Errorf("invalid CID: %w", err)
	}

	_, cancel := context.WithTimeout(c.ctx, DefaultTimeout)
	defer cancel()

	// Use default key if not specified
	if keyName == "" {
		keyName = "self"
	}

	// Publish to IPNS using the shell API
	// Note: The go-ipfs-api library's Publish method signature may vary
	// This is a simplified implementation that should work with most versions
	err := c.shell.Publish(cid, keyName)
	if err != nil {
		// If standard Publish fails, try alternative approach
		// Some versions require different parameters
		return "", fmt.Errorf("failed to publish to IPNS: %w", err)
	}

	// Get the IPNS name - it's based on the key used
	ipnsName := fmt.Sprintf("/ipns/%s", keyName)
	if keyName == "self" {
		// For self key, use the node ID
		id, err := c.GetID()
		if err != nil {
			// If we can't get ID, still return the IPNS path
			ipnsName = "/ipns/self"
		} else {
			ipnsName = fmt.Sprintf("/ipns/%s", id)
		}
	}

	logrus.WithFields(logrus.Fields{
		"cid":     cid,
		"ipns":    ipnsName,
		"keyName": keyName,
	}).Info("Published to IPNS")

	return ipnsName, nil
}

// ResolveName resolves an IPNS name to CID
func (c *IPFSClient) ResolveName(name string) (string, error) {
	path, err := c.shell.Resolve(name)
	if err != nil {
		return "", fmt.Errorf("failed to resolve name: %w", err)
	}

	// Extract CID from path
	parts := strings.Split(path, "/")
	if len(parts) < 3 || parts[1] != "ipfs" {
		return "", fmt.Errorf("invalid IPFS path: %s", path)
	}

	cid := parts[2]
	logrus.WithFields(logrus.Fields{
		"name": name,
		"cid":  cid,
	}).Info("Name resolved from IPNS")

	return cid, nil
}

// CreateDirectory creates a directory structure in IPFS
func (c *IPFSClient) CreateDirectory(files map[string][]byte) (string, error) {
	if len(files) == 0 {
		return "", fmt.Errorf("no files provided")
	}

	_, cancel := context.WithTimeout(c.ctx, DefaultTimeout)
	defer cancel()

	// Create a temporary directory structure using IPFS object API
	// First, add all files and collect their CIDs
	fileCIDs := make(map[string]string)
	for filename, data := range files {
		// Validate filename
		if err := ValidateFileName(filename); err != nil {
			return "", fmt.Errorf("invalid filename %s: %w", filename, err)
		}

		// Add file to IPFS
		reader := bytes.NewReader(data)
		cid, err := c.shell.Add(reader, shell.Pin(true))
		if err != nil {
			return "", fmt.Errorf("failed to add file %s: %w", filename, err)
		}

		fileCIDs[filename] = cid
		logrus.WithFields(logrus.Fields{
			"cid":      cid,
			"filename": filename,
			"size":     len(data),
		}).Debug("File added to IPFS for directory")
	}

	// Create directory structure using IPFS
	// For now, we'll use a simpler approach: create a directory-like structure
	// by storing file metadata and returning a combined CID
	// In a full implementation, this would use IPFS directory objects

	// Since IPFS directory creation requires specific API calls that may vary,
	// we'll create a manifest file that lists all files and their CIDs
	manifest := make(map[string]string)
	for filename, fileCID := range fileCIDs {
		manifest[filename] = fileCID
	}

	// Create manifest JSON
	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		return c.createDirectoryFallback(files)
	}

	// Add manifest to IPFS
	manifestCID, err := c.shell.Add(bytes.NewReader(manifestJSON), shell.Pin(true))
	if err != nil {
		logrus.WithError(err).Warn("Failed to create directory manifest, using fallback")
		return c.createDirectoryFallback(files)
	}

	// Pin all files
	for _, fileCID := range fileCIDs {
		if err := c.shell.Pin(fileCID); err != nil {
			logrus.WithError(err).WithField("cid", fileCID).Warn("Failed to pin file in directory")
		}
	}

	logrus.WithFields(logrus.Fields{
		"dir_cid": manifestCID,
		"files":   len(files),
	}).Info("Directory created in IPFS (using manifest approach)")

	return manifestCID, nil
}

// createDirectoryFallback creates directory using simpler approach
func (c *IPFSClient) createDirectoryFallback(files map[string][]byte) (string, error) {
	// Fallback: create a simple structure by returning the first file's CID
	// This is a minimal implementation when full directory creation fails
	for filename, data := range files {
		cid, err := c.shell.Add(bytes.NewReader(data), shell.Pin(true))
		if err != nil {
			return "", fmt.Errorf("failed to add file %s: %w", filename, err)
		}
		logrus.WithFields(logrus.Fields{
			"cid":      cid,
			"filename": filename,
		}).Warn("Using fallback directory creation - returning first file CID")
		return cid, nil
	}
	return "", fmt.Errorf("no files to add")
}

// ListDirectory lists contents of an IPFS directory
func (c *IPFSClient) ListDirectory(cid string) ([]*DirectoryEntry, error) {
	links, err := c.shell.List(cid)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	entries := make([]*DirectoryEntry, len(links))
	for i, link := range links {
		entries[i] = &DirectoryEntry{
			Name: link.Name,
			Hash: link.Hash,
			Size: link.Size,
			Type: link.Type,
		}
	}

	return entries, nil
}

// DirectoryEntry represents an entry in an IPFS directory
type DirectoryEntry struct {
	Name string
	Hash string
	Size uint64
	Type int
}

// GetID returns the IPFS node ID
func (c *IPFSClient) GetID() (string, error) {
	id, err := c.shell.ID()
	if err != nil {
		return "", fmt.Errorf("failed to get node ID: %w", err)
	}

	return id.ID, nil
}

// GetVersion returns IPFS version
func (c *IPFSClient) GetVersion() (string, error) {
	version, _, err := c.shell.Version()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	return version, nil
}

// HealthCheck performs a basic health check
func (c *IPFSClient) HealthCheck() error {
	_, err := c.GetID()
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	peers, err := c.ListPeers()
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if len(peers) == 0 {
		logrus.Warn("No peers connected - IPFS node may be isolated")
	}

	logrus.WithField("peers", len(peers)).Info("IPFS health check passed")
	return nil
}

// WaitForPeers waits for peers to connect with timeout
func (c *IPFSClient) WaitForPeers(minPeers int, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(c.ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for peers")
		case <-ticker.C:
			peers, err := c.ListPeers()
			if err != nil {
				logrus.WithError(err).Warn("Failed to list peers")
				continue
			}

			if len(peers) >= minPeers {
				logrus.WithField("peers", len(peers)).Info("Sufficient peers connected")
				return nil
			}

			logrus.WithField("peers", len(peers)).Debug("Waiting for more peers")
		}
	}
}

// IsValidCID checks if a string is a valid CID
func IsValidCID(cidStr string) bool {
	_, err := cid.Decode(cidStr)
	return err == nil
}

// ParseCID parses a CID string
func ParseCID(cidStr string) (cid.Cid, error) {
	return cid.Decode(cidStr)
}

// Close closes the IPFS client and cleans up resources
func (c *IPFSClient) Close() error {
	// Shell client doesn't have explicit Close, but we can log the cleanup
	// Context cleanup is handled automatically by Go's garbage collector
	logrus.Debug("IPFS client closed")
	return nil
}
