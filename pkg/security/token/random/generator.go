package random

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"

	"github.com/0xsj/result"
)

// Generator generates cryptographically secure random tokens.
type Generator interface {
	// GenerateBytes generates random bytes
	GenerateBytes(length int) result.Result[[]byte]

	// GenerateBase64 generates base64-encoded token (URL-safe)
	GenerateBase64(length int) result.Result[string]

	// GenerateHex generates hex-encoded token
	GenerateHex(length int) result.Result[string]
}

type generator struct{}

// NewGenerator creates a new random token generator.
func NewGenerator() Generator {
	return &generator{}
}

// GenerateBytes generates cryptographically secure random bytes.
func (g *generator) GenerateBytes(length int) result.Result[[]byte] {
	if length <= 0 {
		return result.Err[[]byte](ErrInvalidLength{Length: length})
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return result.Err[[]byte](ErrGenerationFailed{Err: err})
	}

	return result.Ok(bytes)
}

// GenerateBase64 generates a base64-encoded random token (URL-safe).
func (g *generator) GenerateBase64(length int) result.Result[string] {
	return result.AndThenMap(
		g.GenerateBytes(length),
		func(bytes []byte) result.Result[string] {
			token := base64.URLEncoding.EncodeToString(bytes)
			return result.Ok(token)
		},
	)
}

// GenerateHex generates a hex-encoded random token.
func (g *generator) GenerateHex(length int) result.Result[string] {
	return result.AndThenMap(
		g.GenerateBytes(length),
		func(bytes []byte) result.Result[string] {
			token := hex.EncodeToString(bytes)
			return result.Ok(token)
		},
	)
}
