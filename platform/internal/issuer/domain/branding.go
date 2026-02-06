package domain

// IssuerBranding is an immutable value object representing an issuer's visual branding.
type IssuerBranding struct {
	logoURL           string
	primaryColor      string
	secondaryColor    string
	certificateDesign string
}

// NewIssuerBranding creates a new IssuerBranding value object.
func NewIssuerBranding(logoURL, primaryColor, secondaryColor, certificateDesign string) IssuerBranding {
	return IssuerBranding{
		logoURL:           logoURL,
		primaryColor:      primaryColor,
		secondaryColor:    secondaryColor,
		certificateDesign: certificateDesign,
	}
}

// LogoURL returns the logo URL.
func (b IssuerBranding) LogoURL() string {
	return b.logoURL
}

// PrimaryColor returns the primary color.
func (b IssuerBranding) PrimaryColor() string {
	return b.primaryColor
}

// SecondaryColor returns the secondary color.
func (b IssuerBranding) SecondaryColor() string {
	return b.secondaryColor
}

// CertificateDesign returns the certificate design template.
func (b IssuerBranding) CertificateDesign() string {
	return b.certificateDesign
}

// IsZero returns true if the branding has no data.
func (b IssuerBranding) IsZero() bool {
	return b.logoURL == "" && b.primaryColor == "" && b.secondaryColor == "" && b.certificateDesign == ""
}

// ToMap returns the branding as a map for JSON serialization.
func (b IssuerBranding) ToMap() map[string]any {
	return map[string]any{
		"logo_url":           b.logoURL,
		"primary_color":      b.primaryColor,
		"secondary_color":    b.secondaryColor,
		"certificate_design": b.certificateDesign,
	}
}

// BrandingFromMap reconstructs an IssuerBranding from a map.
func BrandingFromMap(m map[string]any) IssuerBranding {
	if m == nil {
		return IssuerBranding{}
	}

	getString := func(key string) string {
		if v, ok := m[key]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
		return ""
	}

	return IssuerBranding{
		logoURL:           getString("logo_url"),
		primaryColor:      getString("primary_color"),
		secondaryColor:    getString("secondary_color"),
		certificateDesign: getString("certificate_design"),
	}
}
