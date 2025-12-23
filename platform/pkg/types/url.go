package types

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// URL is a validated URL value object.
// Ensures URLs are syntactically valid and normalized.
//
// Zero value is invalid — use NewURL() to create valid URLs.
type URL struct {
	value *url.URL
}

// NewURL parses and validates a URL string.
func NewURL(s string) (URL, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return URL{}, fmt.Errorf("url cannot be empty")
	}

	parsed, err := url.Parse(s)
	if err != nil {
		return URL{}, fmt.Errorf("invalid url format: %w", err)
	}

	// Require scheme
	if parsed.Scheme == "" {
		return URL{}, fmt.Errorf("url must have a scheme (http, https, etc.)")
	}

	// Require host for http/https
	if (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Host == "" {
		return URL{}, fmt.Errorf("url must have a host")
	}

	return URL{value: parsed}, nil
}

// MustNewURL parses a URL and panics if invalid.
// Only use for constants or tests.
func MustNewURL(s string) URL {
	u, err := NewURL(s)
	if err != nil {
		panic(fmt.Sprintf("invalid url: %v", err))
	}
	return u
}

// FromStdURL creates a URL from a standard library *url.URL.
func FromStdURL(u *url.URL) URL {
	if u == nil {
		return URL{}
	}
	// Clone to avoid external mutation
	clone := *u
	return URL{value: &clone}
}

// String returns the URL string representation.
func (u URL) String() string {
	if u.value == nil {
		return ""
	}
	return u.value.String()
}

// StdURL returns the underlying *url.URL.
// Returns nil for zero value.
func (u URL) StdURL() *url.URL {
	if u.value == nil {
		return nil
	}
	// Clone to prevent external mutation
	clone := *u.value
	return &clone
}

// IsEmpty returns true if the URL is empty.
func (u URL) IsEmpty() bool {
	return u.value == nil
}

// IsValid returns true if the URL is valid (non-empty).
func (u URL) IsValid() bool {
	return u.value != nil
}

// Scheme returns the URL scheme (http, https, etc.).
func (u URL) Scheme() string {
	if u.value == nil {
		return ""
	}
	return u.value.Scheme
}

// Host returns the host (including port if present).
func (u URL) Host() string {
	if u.value == nil {
		return ""
	}
	return u.value.Host
}

// Hostname returns the host without port.
func (u URL) Hostname() string {
	if u.value == nil {
		return ""
	}
	return u.value.Hostname()
}

// Port returns the port, or empty string if not specified.
func (u URL) Port() string {
	if u.value == nil {
		return ""
	}
	return u.value.Port()
}

// Path returns the URL path.
func (u URL) Path() string {
	if u.value == nil {
		return ""
	}
	return u.value.Path
}

// Query returns the query parameters.
func (u URL) Query() url.Values {
	if u.value == nil {
		return url.Values{}
	}
	return u.value.Query()
}

// Fragment returns the URL fragment (after #).
func (u URL) Fragment() string {
	if u.value == nil {
		return ""
	}
	return u.value.Fragment
}

// IsHTTPS returns true if the scheme is https.
func (u URL) IsHTTPS() bool {
	return u.Scheme() == "https"
}

// IsHTTP returns true if the scheme is http or https.
func (u URL) IsHTTP() bool {
	scheme := u.Scheme()
	return scheme == "http" || scheme == "https"
}

// WithPath returns a new URL with the given path.
func (u URL) WithPath(path string) URL {
	if u.value == nil {
		return URL{}
	}
	clone := *u.value
	clone.Path = path
	return URL{value: &clone}
}

// WithQuery returns a new URL with the given query parameters.
func (u URL) WithQuery(query url.Values) URL {
	if u.value == nil {
		return URL{}
	}
	clone := *u.value
	clone.RawQuery = query.Encode()
	return URL{value: &clone}
}

// JoinPath returns a new URL with path elements joined.
func (u URL) JoinPath(elem ...string) URL {
	if u.value == nil {
		return URL{}
	}
	joined := u.value.JoinPath(elem...)
	return URL{value: joined}
}

// Equals checks if two URLs are equal.
func (u URL) Equals(other URL) bool {
	return u.String() == other.String()
}

// ============================================================================
// JSON Marshaling
// ============================================================================

// MarshalJSON implements json.Marshaler.
func (u URL) MarshalJSON() ([]byte, error) {
	if u.IsEmpty() {
		return []byte("null"), nil
	}
	return json.Marshal(u.String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (u *URL) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	if s == "" || s == "null" {
		*u = URL{}
		return nil
	}

	parsed, err := NewURL(s)
	if err != nil {
		return err
	}

	*u = parsed
	return nil
}

// ============================================================================
// SQL Scanning
// ============================================================================

// Scan implements sql.Scanner.
func (u *URL) Scan(value interface{}) error {
	if value == nil {
		*u = URL{}
		return nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			*u = URL{}
			return nil
		}
		parsed, err := NewURL(v)
		if err != nil {
			return err
		}
		*u = parsed
		return nil

	case []byte:
		if len(v) == 0 {
			*u = URL{}
			return nil
		}
		parsed, err := NewURL(string(v))
		if err != nil {
			return err
		}
		*u = parsed
		return nil

	default:
		return fmt.Errorf("cannot scan %T into URL", value)
	}
}

// Value implements driver.Valuer.
func (u URL) Value() (driver.Value, error) {
	if u.IsEmpty() {
		return nil, nil
	}
	return u.String(), nil
}

// ============================================================================
// Text Marshaling
// ============================================================================

// MarshalText implements encoding.TextMarshaler.
func (u URL) MarshalText() ([]byte, error) {
	if u.IsEmpty() {
		return []byte{}, nil
	}
	return []byte(u.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (u *URL) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		*u = URL{}
		return nil
	}

	parsed, err := NewURL(string(text))
	if err != nil {
		return err
	}

	*u = parsed
	return nil
}

// ============================================================================
// Helpers
// ============================================================================

// GoString implements fmt.GoStringer for debugging.
func (u URL) GoString() string {
	if u.IsEmpty() {
		return "URL{empty}"
	}
	return fmt.Sprintf("URL{%s}", u.String())
}
