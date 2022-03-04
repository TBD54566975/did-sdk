package cryptosuite

import (
	"crypto"
	"reflect"

	"github.com/goccy/go-json"
	"github.com/pkg/errors"

	. "github.com/TBD54566975/did-sdk/util"

	"github.com/gobuffalo/packr/v2"
)

type (
	Proof         interface{}
	SignatureType string
	ProofPurpose  string
)

const (
	W3CSecurityContext                    = "https://w3id.org/security/v1"
	JWS2020LinkedDataContext string       = "https://w3id.org/security/suites/jws-2020/v1"
	AssertionMethod          ProofPurpose = "assertionMethod"
)

var (
	contextBox = packr.New("Known JSON-LD Contexts", "./context")
)

// CryptoSuite encapsulates the behavior of a proof type as per the W3C specification
// on data integrity https://w3c-ccg.github.io/data-integrity-spec/#creating-new-proof-types
type CryptoSuite interface {
	CryptoSuiteInfo

	// Sign https://w3c-ccg.github.io/data-integrity-spec/#proof-algorithm
	// this method mutates the provided provable object, adding a `proof` block`
	Sign(s Signer, p Provable) error
	// Verify https://w3c-ccg.github.io/data-integrity-spec/#proof-verification-algorithm
	Verify(v Verifier, p Provable) error
}

type CryptoSuiteInfo interface {
	ID() string
	Type() LDKeyType
	CanonicalizationAlgorithm() string
	MessageDigestAlgorithm() crypto.Hash
	SignatureAlgorithm() SignatureType
	RequiredContexts() []string
}

// CryptoSuiteProofType is an interface that defines functionality needed to sign and verify data
// It encapsulates the functionality defined by the data integrity proof type specification
// https://w3c-ccg.github.io/data-integrity-spec/#creating-new-proof-types
type CryptoSuiteProofType interface {
	Marshal(data interface{}) ([]byte, error)
	Canonicalize(marshaled []byte) (*string, error)
	// CreateVerifyHash https://w3c-ccg.github.io/data-integrity-spec/#create-verify-hash-algorithm
	CreateVerifyHash(provable Provable, proof Proof, proofOptions *ProofOptions) ([]byte, error)
	Digest(tbd []byte) ([]byte, error)
}

type Provable interface {
	GetProof() *Proof
	SetProof(p *Proof)
}

type Signer interface {
	KeyID() string
	KeyType() string
	SignatureType() SignatureType
	SigningAlgorithm() string
	Sign(tbs []byte) ([]byte, error)
}

type Verifier interface {
	KeyID() string
	KeyType() string
	Verify(message, signature []byte) error
}

type ProofOptions struct {
	// JSON-LD contexts to add to the proof
	Contexts []string
}

// GetContextsFromProvable searches from a Linked Data `@context` property in the document and returns the value
// associated with the context, if it exists.
func GetContextsFromProvable(p Provable) ([]string, error) {
	provableBytes, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	var genericProvable map[string]interface{}
	if err := json.Unmarshal(provableBytes, &genericProvable); err != nil {
		return nil, err
	}
	contexts, ok := genericProvable["@context"]
	if !ok {
		return nil, nil
	}
	strContexts, err := InterfaceToStrings(contexts)
	if err != nil {
		return nil, err
	}
	return strContexts, nil
}

func ProvableToType(p Provable, t interface{}) error {
	tType := reflect.TypeOf(t)
	tKind := tType.Kind()
	if !(tKind == reflect.Ptr || tKind == reflect.Slice) {
		return errors.New("t is not of kind ptr or slice")
	}

	jsonBytes, err := json.Marshal(p)
	if err != nil {
		return errors.Wrap(err, "could not convert provable to json")
	}

	return json.Unmarshal(jsonBytes, &t)
}

func getKnownContext(fileName string) (string, error) {
	return contextBox.FindString(fileName)
}
