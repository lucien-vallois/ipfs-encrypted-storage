// Package zkp provides demonstrative zero-knowledge proof functionality.
// The fixed small parameters used by the convenience functions are not secure
// for production.
package zkp

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math/big"
	"slices"
	"sort"
)

// ZKP-related errors
var (
	ErrInvalidProof             = errors.New("invalid zero-knowledge proof")
	ErrProofGeneration          = errors.New("failed to generate proof")
	ErrProofVerification        = errors.New("failed to verify proof")
	ErrRangeProofNotImplemented = errors.New("range proofs are not implemented")
)

// ProofType represents different types of zero-knowledge proofs
type ProofType int

const (
	// ProofOfKnowledge - Prove knowledge of a secret without revealing it
	ProofOfKnowledge ProofType = iota
	// ProofOfOwnership - Prove ownership of a resource
	ProofOfOwnership
	// ProofOfAccess - Prove access rights to encrypted data
	ProofOfAccess
	// ProofOfEquality - Prove that two commitments hide the same value
	ProofOfEquality
)

// ZKProof represents a zero-knowledge proof
type ZKProof struct {
	Type       ProofType     `json:"type"`
	Challenge  *big.Int      `json:"challenge"`
	Response   *big.Int      `json:"response"`
	PublicData []byte        `json:"public_data"`
	Metadata   ProofMetadata `json:"metadata"`
}

// ProofMetadata contains metadata about the proof
type ProofMetadata struct {
	Algorithm string `json:"algorithm"`
	Version   string `json:"version"`
	Timestamp int64  `json:"timestamp"`
}

// SchnorrProof implements Schnorr zero-knowledge proofs
type SchnorrProof struct {
	G *big.Int // Generator
	H *big.Int // Public key (g^x mod p)
	P *big.Int // Prime modulus
	Q *big.Int // Order of G
}

// NewSchnorrProof creates a new Schnorr proof system
func NewSchnorrProof(prime *big.Int, generator *big.Int) *SchnorrProof {
	schnorr := &SchnorrProof{}
	if prime != nil {
		schnorr.P = new(big.Int).Set(prime)
		schnorr.Q = new(big.Int).Sub(new(big.Int).Set(prime), big.NewInt(1))
	}
	if generator != nil {
		schnorr.G = new(big.Int).Set(generator)
	}
	return schnorr
}

// GenerateProof generates a non-interactive proof using an empty application context.
func (s *SchnorrProof) GenerateProof(secret *big.Int, publicKey *big.Int) (*ZKProof, error) {
	return s.GenerateProofWithContext(secret, publicKey, nil)
}

// GenerateProofWithContext binds context to the Fiat-Shamir challenge.
func (s *SchnorrProof) GenerateProofWithContext(secret *big.Int, publicKey *big.Int, context []byte) (*ZKProof, error) {
	if !s.validParameters() || secret == nil || secret.Sign() <= 0 || !s.validGroupElement(publicKey) {
		return nil, ErrProofGeneration
	}

	secretScalar := new(big.Int).Mod(new(big.Int).Set(secret), s.Q)
	if secretScalar.Sign() == 0 {
		return nil, ErrProofGeneration
	}
	expectedPublicKey := new(big.Int).Exp(s.G, secretScalar, s.P)
	if expectedPublicKey.Cmp(publicKey) != 0 {
		return nil, ErrProofGeneration
	}

	var r, commitment *big.Int
	for attempts := 0; attempts < 128; attempts++ {
		var err error
		r, err = randomNonZeroScalar(s.Q)
		if err != nil {
			return nil, ErrProofGeneration
		}
		commitment = new(big.Int).Exp(s.G, r, s.P)
		if s.validGroupElement(commitment) {
			break
		}
		commitment = nil
	}
	if commitment == nil {
		return nil, ErrProofGeneration
	}
	challenge := s.deriveChallenge(publicKey, commitment, context)

	// Compute response: (r + x * challenge) mod q
	xc := new(big.Int).Mul(secretScalar, challenge)
	xc.Mod(xc, s.Q)
	response := new(big.Int).Add(r, xc)
	response.Mod(response, s.Q)

	return &ZKProof{
		Type:       ProofOfKnowledge,
		Challenge:  challenge,
		Response:   response,
		PublicData: commitment.Bytes(),
		Metadata: ProofMetadata{
			Algorithm: "Schnorr",
			Version:   "2.0",
			Timestamp: 0, // Would be set to current time
		},
	}, nil
}

// VerifyProof verifies a non-interactive proof using an empty application context.
func (s *SchnorrProof) VerifyProof(proof *ZKProof, publicKey *big.Int) bool {
	return s.VerifyProofWithContext(proof, publicKey, nil)
}

// VerifyProofWithContext re-derives the Fiat-Shamir challenge for context.
func (s *SchnorrProof) VerifyProofWithContext(proof *ZKProof, publicKey *big.Int, context []byte) bool {
	if !s.validParameters() || proof == nil || proof.Type != ProofOfKnowledge ||
		proof.Challenge == nil || proof.Challenge.Sign() <= 0 || proof.Challenge.Cmp(s.Q) >= 0 ||
		proof.Response == nil || proof.Response.Sign() < 0 || proof.Response.Cmp(s.Q) >= 0 ||
		len(proof.PublicData) == 0 || len(proof.PublicData) > (s.P.BitLen()+7)/8 ||
		proof.PublicData[0] == 0 || !s.validGroupElement(publicKey) {
		return false
	}

	commitment := new(big.Int).SetBytes(proof.PublicData)
	if !s.validGroupElement(commitment) || proof.Challenge.Cmp(s.deriveChallenge(publicKey, commitment, context)) != 0 {
		return false
	}

	gResponse := new(big.Int).Exp(s.G, proof.Response, s.P)
	yChallenge := new(big.Int).Exp(publicKey, proof.Challenge, s.P)
	expected := new(big.Int).Mul(commitment, yChallenge)
	expected.Mod(expected, s.P)

	return gResponse.Cmp(expected) == 0
}

func (s *SchnorrProof) validParameters() bool {
	one := big.NewInt(1)
	if s == nil || s.P == nil || s.Q == nil || s.G == nil ||
		s.P.Cmp(big.NewInt(2)) <= 0 || !s.P.ProbablyPrime(32) ||
		s.Q.Cmp(one) <= 0 || s.Q.Cmp(s.P) >= 0 ||
		s.G.Cmp(one) <= 0 || s.G.Cmp(s.P) >= 0 {
		return false
	}
	return new(big.Int).Exp(s.G, s.Q, s.P).Cmp(one) == 0
}

func (s *SchnorrProof) validGroupElement(value *big.Int) bool {
	if value == nil || value.Cmp(big.NewInt(1)) <= 0 || value.Cmp(s.P) >= 0 {
		return false
	}
	return new(big.Int).Exp(value, s.Q, s.P).Cmp(big.NewInt(1)) == 0
}

func (s *SchnorrProof) deriveChallenge(publicKey, commitment *big.Int, context []byte) *big.Int {
	transcript := encodeFields(
		[]byte("ipfs-encrypted-storage/schnorr-fiat-shamir/v1"),
		[]byte("p"), s.P.Bytes(),
		[]byte("q"), s.Q.Bytes(),
		[]byte("g"), s.G.Bytes(),
		[]byte("y"), publicKey.Bytes(),
		[]byte("t"), commitment.Bytes(),
		[]byte("context"), context,
	)
	digest := sha256.Sum256(transcript)
	challenge := new(big.Int).SetBytes(digest[:])
	challenge.Mod(challenge, new(big.Int).Sub(s.Q, big.NewInt(1)))
	return challenge.Add(challenge, big.NewInt(1))
}

func randomNonZeroScalar(q *big.Int) (*big.Int, error) {
	value, err := rand.Int(rand.Reader, new(big.Int).Sub(q, big.NewInt(1)))
	if err != nil {
		return nil, err
	}
	return value.Add(value, big.NewInt(1)), nil
}

func encodeFields(fields ...[]byte) []byte {
	var size [8]byte
	encoded := make([]byte, 0)
	for _, field := range fields {
		binary.BigEndian.PutUint64(size[:], uint64(len(field)))
		encoded = append(encoded, size[:]...)
		encoded = append(encoded, field...)
	}
	return encoded
}

// newDemoSchnorrProof uses intentionally tiny parameters for examples and tests.
// It provides no production security.
func newDemoSchnorrProof() *SchnorrProof {
	return NewSchnorrProof(big.NewInt(23), big.NewInt(5))
}

// AccessControlProof is demonstrative metadata bound to a Schnorr transcript.
type AccessControlProof struct {
	ResourceID  string   `json:"resource_id"`
	UserID      string   `json:"user_id"`
	Permissions []string `json:"permissions"`
	Proof       *ZKProof `json:"proof"`
}

// GenerateAccessProof generates a proof for access to a resource
func GenerateAccessProof(resourceID, userID string, permissions []string, secret *big.Int) (*AccessControlProof, error) {
	schnorr := newDemoSchnorrProof()
	if resourceID == "" || userID == "" || secret == nil || secret.Sign() <= 0 {
		return nil, ErrProofGeneration
	}
	normalizedPermissions, ok := normalizePermissions(permissions)
	if !ok {
		return nil, ErrProofGeneration
	}
	publicKey := new(big.Int).Exp(schnorr.G, secret, schnorr.P)

	proof, err := schnorr.GenerateProofWithContext(secret, publicKey, accessProofContext(resourceID, userID, normalizedPermissions))
	if err != nil {
		return nil, err
	}

	return &AccessControlProof{
		ResourceID:  resourceID,
		UserID:      userID,
		Permissions: normalizedPermissions,
		Proof:       proof,
	}, nil
}

// VerifyAccessProof verifies only the proof's self-described metadata binding.
// Deprecated: use VerifyAccessProofFor with policy data from a trusted source.
func VerifyAccessProof(accessProof *AccessControlProof, publicKey *big.Int) bool {
	if accessProof == nil || accessProof.Proof == nil {
		return false
	}
	return VerifyAccessProofFor(accessProof, publicKey, accessProof.ResourceID, accessProof.UserID, accessProof.Permissions)
}

// VerifyAccessProofFor verifies a proof against access data supplied by the caller.
func VerifyAccessProofFor(accessProof *AccessControlProof, publicKey *big.Int, resourceID, userID string, permissions []string) bool {
	if accessProof == nil || accessProof.Proof == nil ||
		accessProof.ResourceID != resourceID || accessProof.UserID != userID {
		return false
	}
	actualPermissions, actualOK := normalizePermissions(accessProof.Permissions)
	expectedPermissions, expectedOK := normalizePermissions(permissions)
	if !actualOK || !expectedOK || !slices.Equal(actualPermissions, expectedPermissions) {
		return false
	}
	schnorr := newDemoSchnorrProof()
	context := accessProofContext(resourceID, userID, expectedPermissions)
	return schnorr.VerifyProofWithContext(accessProof.Proof, publicKey, context)
}

func normalizePermissions(permissions []string) ([]string, bool) {
	if len(permissions) == 0 {
		return nil, false
	}
	normalized := append([]string(nil), permissions...)
	sort.Strings(normalized)
	if normalized[0] == "" {
		return nil, false
	}
	for i := 1; i < len(normalized); i++ {
		if normalized[i] == normalized[i-1] {
			return nil, false
		}
	}
	return normalized, true
}

func accessProofContext(resourceID, userID string, permissions []string) []byte {
	var permissionCount [8]byte
	binary.BigEndian.PutUint64(permissionCount[:], uint64(len(permissions)))
	fields := make([][]byte, 0, 4+len(permissions))
	fields = append(fields,
		[]byte("ipfs-encrypted-storage/access-proof/v1"),
		[]byte(resourceID),
		[]byte(userID),
		permissionCount[:],
	)
	for _, permission := range permissions {
		fields = append(fields, []byte(permission))
	}
	return encodeFields(fields...)
}

// RangeProof is retained for API compatibility. No range-proof construction is
// implemented by this demonstrative package.
type RangeProof struct {
	Commitment *big.Int    `json:"commitment"`
	Range      [2]*big.Int `json:"range"` // [min, max]
	Proof      *ZKProof    `json:"proof"`
}

// GenerateRangeProof fails closed because a real range-proof construction is not implemented.
func GenerateRangeProof(value *big.Int, min, max *big.Int, secret *big.Int) (*RangeProof, error) {
	return nil, ErrRangeProofNotImplemented
}

// VerifyRangeProof fails closed because RangeProof has no sound implementation.
func VerifyRangeProof(rangeProof *RangeProof, publicKey *big.Int) bool {
	return false
}

// ZKPManager manages zero-knowledge proof operations
type ZKPManager struct {
	proofs map[string]*ZKProof
}

// NewZKPManager creates a new ZKP manager
func NewZKPManager() *ZKPManager {
	return &ZKPManager{
		proofs: make(map[string]*ZKProof),
	}
}

// StoreProof stores a proof for later verification
func (z *ZKPManager) StoreProof(id string, proof *ZKProof) {
	if z == nil {
		return
	}
	if z.proofs == nil {
		z.proofs = make(map[string]*ZKProof)
	}
	z.proofs[id] = proof
}

// GetProof retrieves a stored proof
func (z *ZKPManager) GetProof(id string) (*ZKProof, bool) {
	if z == nil {
		return nil, false
	}
	proof, exists := z.proofs[id]
	return proof, exists
}

// VerifyStoredProof verifies a stored proof
func (z *ZKPManager) VerifyStoredProof(id string, publicKey *big.Int) bool {
	if z == nil {
		return false
	}
	proof, exists := z.proofs[id]
	if !exists {
		return false
	}

	schnorr := newDemoSchnorrProof()
	return schnorr.VerifyProof(proof, publicKey)
}
