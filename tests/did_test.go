package tests

import (
	"testing"
	"time"

	"ipfs-encrypted-storage/src/did"
)

func TestDIDCreation(t *testing.T) {
	manager := did.NewDIDManager()

	// Create a new DID
	doc, privateKey, err := manager.CreateDID("test")
	if err != nil {
		t.Fatalf("Failed to create DID: %v", err)
	}

	// Verify DID structure
	if doc.ID == "" {
		t.Error("DID ID should not be empty")
	}
	if doc.Controller != doc.ID {
		t.Error("Controller should match DID ID")
	}
	if len(doc.VerificationMethod) == 0 {
		t.Error("Should have at least one verification method")
	}
	if len(doc.Authentication) == 0 {
		t.Error("Should have authentication methods")
	}

	// Verify key pair
	if privateKey == nil {
		t.Error("Private key should not be nil")
	}

	// Test DID format
	if !manager.IsValidDID(doc.ID) {
		t.Errorf("Generated DID is invalid: %s", doc.ID)
	}
}

func TestDIDResolution(t *testing.T) {
	manager := did.NewDIDManager()

	// Create a DID
	doc, _, err := manager.CreateDID("test")
	if err != nil {
		t.Fatalf("Failed to create DID: %v", err)
	}

	// Resolve the DID
	resolved, err := manager.ResolveDID(doc.ID)
	if err != nil {
		t.Fatalf("Failed to resolve DID: %v", err)
	}

	// Verify resolution
	if resolved.ID != doc.ID {
		t.Errorf("Resolved DID ID mismatch: expected %s, got %s", doc.ID, resolved.ID)
	}
}

func TestDIDValidation(t *testing.T) {
	manager := did.NewDIDManager()

	validDIDs := []string{
		"did:test:123456789abcdef",
		"did:example:123",
		"did:ethr:0x123456789abcdef",
	}

	for _, didStr := range validDIDs {
		if !manager.IsValidDID(didStr) {
			t.Errorf("DID should be valid: %s", didStr)
		}
	}

	invalidDIDs := []string{
		"",
		"not-a-did",
		"did:",
		"did:test",
		":test:123",
	}

	for _, didStr := range invalidDIDs {
		if manager.IsValidDID(didStr) {
			t.Errorf("DID should be invalid: %s", didStr)
		}
	}
}

func TestDIDUpdate(t *testing.T) {
	manager := did.NewDIDManager()

	// Create a DID
	doc, _, err := manager.CreateDID("test")
	if err != nil {
		t.Fatalf("Failed to create DID: %v", err)
	}

	originalUpdateTime := doc.Updated

	// Update DID
	updates := &did.DID{
		Service: []did.Service{
			{
				ID:              doc.ID + "#service-1",
				Type:            "EncryptedStorageService",
				ServiceEndpoint: "https://example.com/storage",
			},
		},
	}

	time.Sleep(1 * time.Millisecond) // Ensure timestamp difference

	err = manager.UpdateDID(doc.ID, updates)
	if err != nil {
		t.Fatalf("Failed to update DID: %v", err)
	}

	// Verify update
	updated, err := manager.ResolveDID(doc.ID)
	if err != nil {
		t.Fatalf("Failed to resolve updated DID: %v", err)
	}

	if len(updated.Service) != 1 {
		t.Errorf("Expected 1 service, got %d", len(updated.Service))
	}
	if updated.Updated.Equal(originalUpdateTime) {
		t.Error("Update time should have changed")
	}
}

func TestVerificationMethod(t *testing.T) {
	manager := did.NewDIDManager()

	// Create a DID
	doc, _, err := manager.CreateDID("test")
	if err != nil {
		t.Fatalf("Failed to create DID: %v", err)
	}

	// Add verification method
	newVM := did.VerificationMethod{
		ID:                 doc.ID + "#keys-2",
		Type:               "Ed25519VerificationKey2020",
		Controller:         doc.ID,
		PublicKeyMultibase: "z6Mkexamplekey",
	}

	err = manager.AddVerificationMethod(doc.ID, newVM)
	if err != nil {
		t.Fatalf("Failed to add verification method: %v", err)
	}

	// Verify addition
	updated, err := manager.ResolveDID(doc.ID)
	if err != nil {
		t.Fatalf("Failed to resolve DID: %v", err)
	}

	if len(updated.VerificationMethod) != 2 {
		t.Errorf("Expected 2 verification methods, got %d", len(updated.VerificationMethod))
	}

	found := false
	for _, vm := range updated.VerificationMethod {
		if vm.ID == newVM.ID {
			found = true
			break
		}
	}
	if !found {
		t.Error("New verification method not found")
	}
}

func TestServiceEndpoint(t *testing.T) {
	manager := did.NewDIDManager()

	// Create a DID
	doc, _, err := manager.CreateDID("test")
	if err != nil {
		t.Fatalf("Failed to create DID: %v", err)
	}

	// Add service
	service := did.Service{
		ID:   doc.ID + "#storage-service",
		Type: "EncryptedStorageService",
		ServiceEndpoint: map[string]interface{}{
			"url": "https://storage.example.com",
			"api": "/api/v1",
		},
	}

	err = manager.AddService(doc.ID, service)
	if err != nil {
		t.Fatalf("Failed to add service: %v", err)
	}

	// Verify addition
	updated, err := manager.ResolveDID(doc.ID)
	if err != nil {
		t.Fatalf("Failed to resolve DID: %v", err)
	}

	if len(updated.Service) != 1 {
		t.Errorf("Expected 1 service, got %d", len(updated.Service))
	}

	if updated.Service[0].ID != service.ID {
		t.Errorf("Service ID mismatch: expected %s, got %s", service.ID, updated.Service[0].ID)
	}
}

func TestDIDCredential(t *testing.T) {
	manager := did.NewDIDManager()

	// Create issuer and subject DIDs
	issuerDoc, issuerKey, err := manager.CreateDID("test")
	if err != nil {
		t.Fatalf("Failed to create issuer DID: %v", err)
	}

	subjectDoc, _, err := manager.CreateDID("test")
	if err != nil {
		t.Fatalf("Failed to create subject DID: %v", err)
	}

	// Issue credential
	claims := map[string]interface{}{
		"name":         "John Doe",
		"role":         "Developer",
		"access_level": "admin",
	}

	credential, err := manager.IssueCredential(issuerDoc.ID, subjectDoc.ID, claims, issuerKey)
	if err != nil {
		t.Fatalf("Failed to issue credential: %v", err)
	}

	// Verify credential structure
	if credential.ID == "" {
		t.Error("Credential should have an ID")
	}
	if credential.Issuer != issuerDoc.ID {
		t.Errorf("Credential issuer mismatch: expected %s, got %s", issuerDoc.ID, credential.Issuer)
	}
	if credential.CredentialSubject["id"] != subjectDoc.ID {
		t.Errorf("Credential subject ID mismatch: expected %s, got %s", subjectDoc.ID, credential.CredentialSubject["id"])
	}
	if credential.CredentialSubject["name"] != "John Doe" {
		t.Error("Credential claim 'name' not found or incorrect")
	}

	// Verify credential
	if !manager.VerifyCredential(credential) {
		t.Error("Credential verification failed")
	}
}

func TestDIDRegistry(t *testing.T) {
	registry := did.NewInMemoryDIDRegistry()
	manager := did.NewDIDManager()

	// Create a DID
	doc, _, err := manager.CreateDID("test")
	if err != nil {
		t.Fatalf("Failed to create DID: %v", err)
	}

	// Register DID
	err = registry.Register(doc)
	if err != nil {
		t.Fatalf("Failed to register DID: %v", err)
	}

	// Resolve DID
	resolved, err := registry.Resolve(doc.ID)
	if err != nil {
		t.Fatalf("Failed to resolve DID: %v", err)
	}

	if resolved.ID != doc.ID {
		t.Errorf("Resolved DID ID mismatch: expected %s, got %s", doc.ID, resolved.ID)
	}

	// Update DID
	doc.Controller = "did:test:updated-controller"
	err = registry.Update(doc.ID, doc)
	if err != nil {
		t.Fatalf("Failed to update DID: %v", err)
	}

	// Verify update
	updated, err := registry.Resolve(doc.ID)
	if err != nil {
		t.Fatalf("Failed to resolve updated DID: %v", err)
	}

	if updated.Controller != "did:test:updated-controller" {
		t.Error("DID update not reflected in registry")
	}

	// Deactivate DID
	err = registry.Deactivate(doc.ID)
	if err != nil {
		t.Fatalf("Failed to deactivate DID: %v", err)
	}

	// Verify deactivation
	_, err = registry.Resolve(doc.ID)
	if err != did.ErrDIDNotFound {
		t.Error("DID should not be found after deactivation")
	}
}

func BenchmarkDIDCreation(b *testing.B) {
	manager := did.NewDIDManager()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _, _ = manager.CreateDID("test")
	}
}

func BenchmarkDIDResolution(b *testing.B) {
	manager := did.NewDIDManager()

	// Pre-create DID
	doc, _, err := manager.CreateDID("test")
	if err != nil {
		b.Fatalf("Failed to create test DID: %v", err)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = manager.ResolveDID(doc.ID)
	}
}

func BenchmarkCredentialIssuance(b *testing.B) {
	manager := did.NewDIDManager()

	// Pre-create issuer and subject
	issuerDoc, issuerKey, err := manager.CreateDID("test")
	if err != nil {
		b.Fatalf("Failed to create issuer: %v", err)
	}

	subjectDoc, _, err := manager.CreateDID("test")
	if err != nil {
		b.Fatalf("Failed to create subject: %v", err)
	}

	claims := map[string]interface{}{
		"name": "Benchmark User",
		"role": "Tester",
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = manager.IssueCredential(issuerDoc.ID, subjectDoc.ID, claims, issuerKey)
	}
}
