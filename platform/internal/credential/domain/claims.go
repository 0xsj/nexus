package domain

import (
	"fmt"
	"sort"
)

// Claims is an immutable wrapper around credential claim data.
// Claims represent the verified data points within a credential.
type Claims struct {
	data map[string]any
}

// NewClaims creates a new Claims value object from a map.
// Returns an error if the data is nil or empty.
func NewClaims(data map[string]any) (Claims, error) {
	if data == nil || len(data) == 0 {
		return Claims{}, fmt.Errorf("claims data cannot be nil or empty")
	}

	// Defensive copy
	copied := make(map[string]any, len(data))
	for k, v := range data {
		copied[k] = v
	}

	return Claims{data: copied}, nil
}

// Get returns the value for a given claim key.
func (c Claims) Get(key string) (any, bool) {
	v, ok := c.data[key]
	return v, ok
}

// Keys returns a sorted list of claim keys.
func (c Claims) Keys() []string {
	keys := make([]string, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ToMap returns a defensive copy of the underlying claim data.
func (c Claims) ToMap() map[string]any {
	if c.data == nil {
		return nil
	}
	copied := make(map[string]any, len(c.data))
	for k, v := range c.data {
		copied[k] = v
	}
	return copied
}

// Len returns the number of claims.
func (c Claims) Len() int {
	return len(c.data)
}

// IsZero returns true if the Claims has no data.
func (c Claims) IsZero() bool {
	return len(c.data) == 0
}
