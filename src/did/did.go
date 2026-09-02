// Package did provides decentralized identity (DID) functionality
// Implements W3C DID specification patterns for sovereign identity
package did

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// DID-related errors
var (
	ErrInvalidDID       = fmt.Errorf("invalid DID format")
	ErrDIDNotFound      = fmt.Errorf("DID not found")
	ErrInvalidKey       = fmt.Errorf("invalid key for DID operation")
	ErrResolutionFailed = fmt.Errorf("DID resolution failed")
)

// DID represents a Decentralized Identifier
type DID struct {
	ID                   string               `json:"id"`
	Controller           string               `json:"controller,omitempty"`
	AlsoKnownAs          []string             `json:"alsoKnownAs,omitempty"`
	VerificationMethod   []VerificationMethod `json:"verificationMethod,omitempty"`
	Authentication       []string             `json:"authentication,omitempty"`
	AssertionMethod      []string             `json:"assertionMethod,omitempty"`
	KeyAgreement         []string             `json:"keyAgreement,omitempty"`
	CapabilityInvocation []string             `json:"capabilityInvocation,omitempty"`
	CapabilityDelegation []string             `json:"capabilityDelegation,omitempty"`
	Service              []Service            `json:"service,omitempty"`
	Created              time.Time            `json:"created,omitempty"`
	Updated              time.Time            `json:"updated,omitempty"`
}

// VerificationMethod represents a verification method in a DID document
type VerificationMethod struct {
	ID                 string                 `json:"id"`
	Type               string                 `json:"type"`
	Controller         string                 `json:"controller"`
	PublicKeyMultibase string                 `json:"publicKeyMultibase,omitempty"`
	PublicKeyJwk       map[string]interface{} `json:"publicKeyJwk,omitempty"`
}

// Service represents a service endpoint in a DID document
type Service struct {
	ID              string      `json:"id"`
	Type            string      `json:"type"`
	ServiceEndpoint interface{} `json:"serviceEndpoint"`
}

// DIDDocument represents a complete DID document
type DIDDocument struct {
	*DID
	Context []string `json:"@context,omitempty"`
}

// DIDResolver resolves DIDs to DID documents
type DIDResolver interface {
	Resolve(did string) (*DIDDocument, error)
	Dereference(didURL string) (interface{}, error)
}

// DIDManager manages DID operations
type DIDManager struct {
	documents map[string]*DIDDocument
	resolvers map[string]DIDResolver
}

// NewDIDManager creates a new DID manager
func NewDIDManager() *DIDManager {
	return &DIDManager{
		documents: make(map[string]*DIDDocument),
		resolvers: make(map[string]DIDResolver),
	}
}

// CreateDID creates a new DID with Ed25519 key pair
func (dm *DIDManager) CreateDID(method string) (*DIDDocument, ed25519.PrivateKey, error) {
	// Generate Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Create DID ID
	didID := fmt.Sprintf("did:%s:%x", method, publicKey[:16]) // Simplified DID generation

	// Create verification method
	vm := VerificationMethod{
		ID:         fmt.Sprintf("%s#keys-1", didID),
		Type:       "Ed25519VerificationKey2020",
		Controller: didID,
		// In practice, would encode public key in multibase format
		PublicKeyMultibase: fmt.Sprintf("z%x", publicKey),
	}

	// Create DID document
	doc := &DIDDocument{
		DID: &DID{
			ID:                 didID,
			Controller:         didID,
			VerificationMethod: []VerificationMethod{vm},
			Authentication:     []string{fmt.Sprintf("%s#keys-1", didID)},
			AssertionMethod:    []string{fmt.Sprintf("%s#keys-1", didID)},
			Created:            time.Now(),
			Updated:            time.Now(),
		},
		Context: []string{"https://www.w3.org/ns/did/v1"},
	}

	// Store the document
	dm.documents[didID] = doc

	return doc, privateKey, nil
}

// ResolveDID resolves a DID to its document
func (dm *DIDManager) ResolveDID(did string) (*DIDDocument, error) {
	if !dm.IsValidDID(did) {
		return nil, ErrInvalidDID
	}

	doc, exists := dm.documents[did]
	if !exists {
		return nil, ErrDIDNotFound
	}

	return doc, nil
}

// IsValidDID checks if a string is a valid DID
func (dm *DIDManager) IsValidDID(did string) bool {
	parts := strings.Split(did, ":")
	if len(parts) < 3 || parts[0] != "did" {
		return false
	}

	// Basic validation - in practice would be more comprehensive
	method := parts[1]
	if method == "" {
		return false
	}

	return true
}

// UpdateDID updates a DID document
func (dm *DIDManager) UpdateDID(did string, updates *DID) error {
	doc, exists := dm.documents[did]
	if !exists {
		return ErrDIDNotFound
	}

	// Apply updates (simplified)
	if updates.Controller != "" {
		doc.Controller = updates.Controller
	}
	if updates.Service != nil {
		doc.Service = updates.Service
	}
	doc.Updated = time.Now()

	return nil
}

// AddVerificationMethod adds a verification method to a DID
func (dm *DIDManager) AddVerificationMethod(did string, vm VerificationMethod) error {
	doc, exists := dm.documents[did]
	if !exists {
		return ErrDIDNotFound
	}

	doc.VerificationMethod = append(doc.VerificationMethod, vm)
	doc.Updated = time.Now()

	return nil
}

// AddService adds a service endpoint to a DID
func (dm *DIDManager) AddService(did string, service Service) error {
	doc, exists := dm.documents[did]
	if !exists {
		return ErrDIDNotFound
	}

	doc.Service = append(doc.Service, service)
	doc.Updated = time.Now()

	return nil
}

// DIDCredential represents a verifiable credential
type DIDCredential struct {
	Context           []string               `json:"@context"`
	ID                string                 `json:"id"`
	Type              []string               `json:"type"`
	Issuer            string                 `json:"issuer"`
	IssuanceDate      time.Time              `json:"issuanceDate"`
	ExpirationDate    *time.Time             `json:"expirationDate,omitempty"`
	CredentialSubject map[string]interface{} `json:"credentialSubject"`
	Proof             CredentialProof        `json:"proof,omitempty"`
}

// CredentialProof represents a proof for a credential
type CredentialProof struct {
	Type               string    `json:"type"`
	Created            time.Time `json:"created"`
	VerificationMethod string    `json:"verificationMethod"`
	ProofPurpose       string    `json:"proofPurpose"`
	ProofValue         string    `json:"proofValue"`
}

// IssueCredential issues a credential for a DID
func (dm *DIDManager) IssueCredential(issuerDID string, subjectDID string, claims map[string]interface{}, privateKey ed25519.PrivateKey) (*DIDCredential, error) {
	// Check if issuer DID exists
	_, err := dm.ResolveDID(issuerDID)
	if err != nil {
		return nil, fmt.Errorf("issuer DID not found: %w", err)
	}

	credential := &DIDCredential{
		Context: []string{
			"https://www.w3.org/2018/credentials/v1",
			"https://www.w3.org/ns/did/v1",
		},
		ID:           fmt.Sprintf("urn:uuid:%d", time.Now().UnixNano()),
		Type:         []string{"VerifiableCredential"},
		Issuer:       issuerDID,
		IssuanceDate: time.Now(),
		CredentialSubject: map[string]interface{}{
			"id": subjectDID,
		},
	}

	// Add claims to subject
	for k, v := range claims {
		credential.CredentialSubject[k] = v
	}

	// Create proof (simplified)
	credentialData, _ := json.Marshal(credential)
	signature := ed25519.Sign(privateKey, credentialData)

	credential.Proof = CredentialProof{
		Type:               "Ed25519Signature2020",
		Created:            time.Now(),
		VerificationMethod: fmt.Sprintf("%s#keys-1", issuerDID),
		ProofPurpose:       "assertionMethod",
		ProofValue:         fmt.Sprintf("z%s", signature),
	}

	return credential, nil
}

// VerifyCredential verifies a DID credential
func (dm *DIDManager) VerifyCredential(credential *DIDCredential) bool {
	// Check if issuer DID exists
	issuerDoc, err := dm.ResolveDID(credential.Issuer)
	if err != nil {
		return false
	}

	// Check expiration
	if credential.ExpirationDate != nil && time.Now().After(*credential.ExpirationDate) {
		return false
	}

	// Verify proof (simplified)
	// In practice, would verify the cryptographic signature
	return len(issuerDoc.VerificationMethod) > 0
}

// DIDRegistry represents a DID registry for persistence
type DIDRegistry interface {
	Register(did *DIDDocument) error
	Update(did string, document *DIDDocument) error
	Resolve(did string) (*DIDDocument, error)
	Deactivate(did string) error
}

// InMemoryDIDRegistry implements an in-memory DID registry
type InMemoryDIDRegistry struct {
	documents map[string]*DIDDocument
}

// NewInMemoryDIDRegistry creates a new in-memory DID registry
func NewInMemoryDIDRegistry() *InMemoryDIDRegistry {
	return &InMemoryDIDRegistry{
		documents: make(map[string]*DIDDocument),
	}
}

// Register registers a DID document
func (r *InMemoryDIDRegistry) Register(did *DIDDocument) error {
	r.documents[did.ID] = did
	return nil
}

// Update updates a DID document
func (r *InMemoryDIDRegistry) Update(did string, document *DIDDocument) error {
	if _, exists := r.documents[did]; !exists {
		return ErrDIDNotFound
	}
	r.documents[did] = document
	return nil
}

// Resolve resolves a DID to its document
func (r *InMemoryDIDRegistry) Resolve(did string) (*DIDDocument, error) {
	doc, exists := r.documents[did]
	if !exists {
		return nil, ErrDIDNotFound
	}
	return doc, nil
}

// Deactivate deactivates a DID
func (r *InMemoryDIDRegistry) Deactivate(did string) error {
	delete(r.documents, did)
	return nil
}
