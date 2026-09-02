// Package handlers provides P2P protocol handlers
package handlers

import (
	"encoding/json"

	"github.com/sirupsen/logrus"
)

// HandleTopicMessage handles PubSub topic messages
func HandleTopicMessage(peerID string, message []byte) error {
	logrus.WithField("peer", peerID).Debug("Received topic message")

	// Try to parse as JSON first
	var msg map[string]interface{}
	if err := json.Unmarshal(message, &msg); err != nil {
		// If not JSON, treat as plain text
		logrus.WithFields(logrus.Fields{
			"peer": peerID,
			"size": len(message),
			"text": string(message),
		}).Debug("Received plain text topic message")
		return nil
	}

	// Handle structured messages
	msgType, ok := msg["type"].(string)
	if !ok {
		logrus.Debug("Topic message without type field")
		return nil
	}

	switch msgType {
	case "announcement":
		logrus.WithFields(logrus.Fields{
			"peer": peerID,
			"data": msg,
		}).Info("Received announcement via topic")

	case "file_available":
		cid, _ := msg["cid"].(string)
		logrus.WithFields(logrus.Fields{
			"peer": peerID,
			"cid":  cid,
		}).Info("File availability announced via topic")

	default:
		logrus.WithFields(logrus.Fields{
			"peer": peerID,
			"type": msgType,
		}).Debug("Received topic message")
	}

	return nil
}
