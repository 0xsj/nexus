package id

import (
	"github.com/0xsj/nexus/platform/pkg/types"
)

// Generator generates unique identifiers.
type Generator interface {
	Generate() types.ID
}

// DefaultGenerator generates UUIDv7 IDs.
type DefaultGenerator struct{}

// NewGenerator creates a new default ID generator.
func NewGenerator() *DefaultGenerator {
	return &DefaultGenerator{}
}

// Generate generates a new ID.
func (g *DefaultGenerator) Generate() types.ID {
	return types.NewID()
}

// GeneratorFunc is a function adapter for Generator.
type GeneratorFunc func() types.ID

// Generate implements Generator.
func (f GeneratorFunc) Generate() types.ID {
	return f()
}
