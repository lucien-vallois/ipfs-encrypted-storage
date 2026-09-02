package tests

import (
	"errors"
	"math/big"
	"testing"

	"ipfs-encrypted-storage/src/zkp"
)

func TestSchnorrProof(t *testing.T) {
	// Create Schnorr proof system with small parameters for testing
	p := big.NewInt(23) // Small prime
	g := big.NewInt(5)  // Generator

	schnorr := zkp.NewSchnorrProof(p, g)

	// Test secret and public key
	secret := big.NewInt(7)                     // x = 7
	publicKey := new(big.Int).Exp(g, secret, p) // y = g^x mod p

	// Generate proof
	proof, err := schnorr.GenerateProof(secret, publicKey)
	if err != nil {
		t.Fatalf("Failed to generate Schnorr proof: %v", err)
	}

	// Verify proof
	if !schnorr.VerifyProof(proof, publicKey) {
		t.Error("Schnorr proof verification failed")
	}

	// Test with wrong public key
	wrongPublicKey := new(big.Int).Exp(g, big.NewInt(8), p)
	if schnorr.VerifyProof(proof, wrongPublicKey) {
		t.Error("Proof should fail with wrong public key")
	}
}

func TestSchnorrProofRejectsEquationOnlyForgery(t *testing.T) {
	p := big.NewInt(23)
	g := big.NewInt(5)
	publicKey := big.NewInt(17)
	forged := &zkp.ZKProof{
		Type:       zkp.ProofOfKnowledge,
		Challenge:  big.NewInt(1),
		Response:   big.NewInt(1),
		PublicData: big.NewInt(3).Bytes(),
	}

	left := new(big.Int).Exp(g, forged.Response, p)
	right := new(big.Int).Exp(publicKey, forged.Challenge, p)
	right.Mul(right, big.NewInt(3)).Mod(right, p)
	if left.Cmp(right) != 0 {
		t.Fatal("invalid forgery fixture: Schnorr equation must hold")
	}
	if zkp.NewSchnorrProof(p, g).VerifyProof(forged, publicKey) {
		t.Fatal("equation-only forgery passed without a valid Fiat-Shamir challenge")
	}
}

func TestSchnorrProofBindsContext(t *testing.T) {
	p := big.NewInt(23)
	g := big.NewInt(5)
	secret := big.NewInt(7)
	publicKey := new(big.Int).Exp(g, secret, p)
	schnorr := zkp.NewSchnorrProof(p, g)

	proof, err := schnorr.GenerateProofWithContext(secret, publicKey, []byte("context-a"))
	if err != nil {
		t.Fatal(err)
	}
	if !schnorr.VerifyProofWithContext(proof, publicKey, []byte("context-a")) {
		t.Fatal("proof failed with its original context")
	}
	if schnorr.VerifyProofWithContext(proof, publicKey, []byte("context-b")) {
		t.Fatal("proof verified with a different context")
	}

	withoutContext, err := schnorr.GenerateProofWithContext(secret, publicKey, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !schnorr.VerifyProofWithContext(withoutContext, publicKey, []byte{}) {
		t.Fatal("nil and empty contexts should have the same canonical encoding")
	}
}

func TestSchnorrProofHandlesLowOrderGenerator(t *testing.T) {
	schnorr := zkp.NewSchnorrProof(big.NewInt(23), big.NewInt(22))
	secret := big.NewInt(1)
	publicKey := big.NewInt(22)
	for i := 0; i < 25; i++ {
		proof, err := schnorr.GenerateProof(secret, publicKey)
		if err != nil {
			t.Fatalf("generation failed on iteration %d: %v", i, err)
		}
		if !schnorr.VerifyProof(proof, publicKey) {
			t.Fatalf("generated proof did not verify on iteration %d", i)
		}
	}
}

func TestAccessControlProof(t *testing.T) {
	resourceID := "resource-123"
	userID := "user-456"
	permissions := []string{"write", "read"}
	secret := big.NewInt(7)

	// Generate access proof
	accessProof, err := zkp.GenerateAccessProof(resourceID, userID, permissions, secret)
	if err != nil {
		t.Fatalf("Failed to generate access proof: %v", err)
	}

	// Verify access proof
	p := big.NewInt(23)
	g := big.NewInt(5)
	publicKey := new(big.Int).Exp(g, secret, p)

	if !zkp.VerifyAccessProof(accessProof, publicKey) {
		t.Error("Access proof verification failed")
	}

	// Test metadata
	if accessProof.ResourceID != resourceID {
		t.Errorf("Expected resource ID %s, got %s", resourceID, accessProof.ResourceID)
	}
	if accessProof.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, accessProof.UserID)
	}
	if len(accessProof.Permissions) != len(permissions) {
		t.Errorf("Expected %d permissions, got %d", len(permissions), len(accessProof.Permissions))
	}
	if accessProof.Permissions[0] != "read" || accessProof.Permissions[1] != "write" {
		t.Fatalf("permissions were not normalized: %v", accessProof.Permissions)
	}
	if !zkp.VerifyAccessProofFor(accessProof, publicKey, resourceID, userID, permissions) {
		t.Fatal("access proof failed against matching external expectations")
	}
	if _, err := zkp.GenerateAccessProof(resourceID, userID, []string{"read", "read"}, secret); err == nil {
		t.Fatal("duplicate permissions must be rejected")
	}

	tamperCases := []struct {
		name   string
		mutate func(*zkp.AccessControlProof)
	}{
		{"resource ID", func(proof *zkp.AccessControlProof) { proof.ResourceID = "resource-999" }},
		{"user ID", func(proof *zkp.AccessControlProof) { proof.UserID = "user-999" }},
		{"permissions", func(proof *zkp.AccessControlProof) { proof.Permissions[0] = "admin" }},
	}
	for _, test := range tamperCases {
		t.Run("rejects tampered "+test.name, func(t *testing.T) {
			tampered := *accessProof
			tampered.Permissions = append([]string(nil), accessProof.Permissions...)
			test.mutate(&tampered)
			if zkp.VerifyAccessProofFor(&tampered, publicKey, resourceID, userID, permissions) {
				t.Fatalf("proof verified after tampering with %s", test.name)
			}
		})
	}
}

func TestZKPRejectsMalformedInputsWithoutPanicking(t *testing.T) {
	p := big.NewInt(23)
	g := big.NewInt(5)
	secret := big.NewInt(7)
	publicKey := new(big.Int).Exp(g, secret, p)
	schnorr := zkp.NewSchnorrProof(p, g)
	proof, err := schnorr.GenerateProof(secret, publicKey)
	if err != nil {
		t.Fatalf("failed to create valid fixture: %v", err)
	}

	if schnorr.VerifyProof(nil, publicKey) || schnorr.VerifyProof(proof, nil) {
		t.Fatal("nil proof or public key must be rejected")
	}
	if _, err := schnorr.GenerateProof(nil, publicKey); err == nil {
		t.Fatal("nil secret must be rejected")
	}
	if _, err := zkp.NewSchnorrProof(nil, nil).GenerateProof(secret, publicKey); err == nil {
		t.Fatal("nil parameters must be rejected")
	}

	mutations := []struct {
		name   string
		mutate func(*zkp.ZKProof)
	}{
		{"wrong proof type", func(proof *zkp.ZKProof) { proof.Type = zkp.ProofOfAccess }},
		{"changed challenge", func(proof *zkp.ZKProof) {
			proof.Challenge.Add(proof.Challenge, big.NewInt(1))
			if proof.Challenge.Cmp(big.NewInt(22)) >= 0 {
				proof.Challenge.SetInt64(1)
			}
		}},
		{"nil challenge", func(proof *zkp.ZKProof) { proof.Challenge = nil }},
		{"zero challenge", func(proof *zkp.ZKProof) { proof.Challenge = big.NewInt(0) }},
		{"challenge at order", func(proof *zkp.ZKProof) { proof.Challenge = big.NewInt(22) }},
		{"oversized challenge", func(proof *zkp.ZKProof) { proof.Challenge = new(big.Int).Lsh(big.NewInt(1), 256) }},
		{"nil response", func(proof *zkp.ZKProof) { proof.Response = nil }},
		{"changed response", func(proof *zkp.ZKProof) {
			proof.Response.Add(proof.Response, big.NewInt(1)).Mod(proof.Response, big.NewInt(22))
		}},
		{"negative response", func(proof *zkp.ZKProof) { proof.Response = big.NewInt(-1) }},
		{"response at order", func(proof *zkp.ZKProof) { proof.Response = big.NewInt(22) }},
		{"empty commitment", func(proof *zkp.ZKProof) { proof.PublicData = nil }},
		{"changed commitment", func(proof *zkp.ZKProof) {
			replacement := big.NewInt(2)
			if new(big.Int).SetBytes(proof.PublicData).Cmp(replacement) == 0 {
				replacement.SetInt64(3)
			}
			proof.PublicData = replacement.Bytes()
		}},
		{"oversized commitment", func(proof *zkp.ZKProof) { proof.PublicData = []byte{1, 0} }},
		{"non-canonical commitment", func(proof *zkp.ZKProof) { proof.PublicData = append([]byte{0}, proof.PublicData...) }},
	}
	for _, test := range mutations {
		t.Run(test.name, func(t *testing.T) {
			malformed := *proof
			malformed.Challenge = new(big.Int).Set(proof.Challenge)
			malformed.Response = new(big.Int).Set(proof.Response)
			malformed.PublicData = append([]byte(nil), proof.PublicData...)
			test.mutate(&malformed)
			if schnorr.VerifyProof(&malformed, publicKey) {
				t.Fatal("malformed proof was accepted")
			}
		})
	}

	if zkp.VerifyAccessProof(nil, publicKey) || zkp.VerifyRangeProof(nil, publicKey) {
		t.Fatal("nil wrapper proofs must be rejected")
	}
	if _, err := zkp.GenerateAccessProof("resource", "user", nil, nil); err == nil {
		t.Fatal("access proof with nil secret must be rejected")
	}
	if _, err := zkp.GenerateRangeProof(nil, big.NewInt(0), big.NewInt(1), secret); err == nil {
		t.Fatal("range proof with nil value must be rejected")
	}
}

func TestRangeProofFailsClosed(t *testing.T) {
	value := big.NewInt(50)
	min := big.NewInt(0)
	max := big.NewInt(100)
	secret := big.NewInt(7)

	rangeProof, err := zkp.GenerateRangeProof(value, min, max, secret)
	if rangeProof != nil || !errors.Is(err, zkp.ErrRangeProofNotImplemented) {
		t.Fatalf("range proof must fail closed with not-implemented error, got proof=%v err=%v", rangeProof, err)
	}

	p := big.NewInt(23)
	g := big.NewInt(5)
	publicKey := new(big.Int).Exp(g, secret, p)
	if zkp.VerifyRangeProof(&zkp.RangeProof{}, publicKey) {
		t.Fatal("unimplemented range proof must never verify")
	}
}

func TestZKPManager(t *testing.T) {
	manager := zkp.NewZKPManager()

	// Create a proof
	p := big.NewInt(23)
	g := big.NewInt(5)
	schnorr := zkp.NewSchnorrProof(p, g)
	secret := big.NewInt(7)
	publicKey := new(big.Int).Exp(g, secret, p)

	proof, err := schnorr.GenerateProof(secret, publicKey)
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	// Store proof
	proofID := "test-proof-123"
	manager.StoreProof(proofID, proof)

	// Retrieve and verify proof
	if !manager.VerifyStoredProof(proofID, publicKey) {
		t.Error("Stored proof verification failed")
	}

	// Test non-existent proof
	if manager.VerifyStoredProof("non-existent", publicKey) {
		t.Error("Should fail verification for non-existent proof")
	}
}

func TestProofMetadata(t *testing.T) {
	p := big.NewInt(23)
	g := big.NewInt(5)
	schnorr := zkp.NewSchnorrProof(p, g)
	secret := big.NewInt(7)
	publicKey := new(big.Int).Exp(g, secret, p)

	proof, err := schnorr.GenerateProof(secret, publicKey)
	if err != nil {
		t.Fatalf("Failed to generate proof: %v", err)
	}

	// Check metadata
	if proof.Metadata.Algorithm != "Schnorr" {
		t.Errorf("Expected algorithm 'Schnorr', got '%s'", proof.Metadata.Algorithm)
	}
	if proof.Metadata.Version != "2.0" {
		t.Errorf("Expected version '2.0', got '%s'", proof.Metadata.Version)
	}
}

func BenchmarkSchnorrProofGeneration(b *testing.B) {
	p := big.NewInt(23)
	g := big.NewInt(5)
	schnorr := zkp.NewSchnorrProof(p, g)
	secret := big.NewInt(7)
	publicKey := new(big.Int).Exp(g, secret, p)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = schnorr.GenerateProof(secret, publicKey)
	}
}

func BenchmarkSchnorrProofVerification(b *testing.B) {
	p := big.NewInt(23)
	g := big.NewInt(5)
	schnorr := zkp.NewSchnorrProof(p, g)
	secret := big.NewInt(7)
	publicKey := new(big.Int).Exp(g, secret, p)

	proof, _ := schnorr.GenerateProof(secret, publicKey)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = schnorr.VerifyProof(proof, publicKey)
	}
}

func BenchmarkAccessProofGeneration(b *testing.B) {
	resourceID := "resource-123"
	userID := "user-456"
	permissions := []string{"read", "write"}
	secret := big.NewInt(7)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = zkp.GenerateAccessProof(resourceID, userID, permissions, secret)
	}
}
