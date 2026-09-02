// Package handlers provides P2P protocol handlers
package handlers

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// StorageMessage represents a storage protocol message
type StorageMessage struct {
	Type      string `json:"type"` // "announce", "query", "response"
	CID       string `json:"cid,omitempty"`
	Metadata  string `json:"metadata,omitempty"`
	Timestamp int64  `json:"timestamp"`
}

// HandleStorageMessage handles storage protocol messages
func HandleStorageMessage(peerID string, message []byte) error {
	logrus.WithField("peer", peerID).Info("Received storage message")

	var msg StorageMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		logrus.WithError(err).Error("Failed to parse storage message")
		return err
	}

	switch msg.Type {
	case "announce":
		// Peer is announcing they have content
		logrus.WithFields(logrus.Fields{
			"peer": peerID,
			"cid":  msg.CID,
		}).Info("Peer announced content availability")
		// Could store this in a local cache for future lookups

	case "query":
		// Peer is querying if we have content
		logrus.WithFields(logrus.Fields{
			"peer": peerID,
			"cid":  msg.CID,
		}).Info("Peer queried for content")
		// Could check if we have the content and respond

	default:
		logrus.WithField("type", msg.Type).Warn("Unknown storage message type")
	}

	return nil
}
