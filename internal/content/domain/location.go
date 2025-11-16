package domain

import "github.com/0xsj/result"

// Location represents a geographic location.
type Location struct {
	Latitude  float64
	Longitude float64
	Name      string
	PlaceID   *string
}

// NewLocation creates a new Location.
func NewLocation(latitude, longitude float64, name string, placeID *string) result.Result[*Location] {
	// Validate latitude (-90 to 90)
	if latitude < -90 || latitude > 90 {
		return result.Err[*Location](ErrInvalidLocation())
	}

	// Validate longitude (-180 to 180)
	if longitude < -180 || longitude > 180 {
		return result.Err[*Location](ErrInvalidLocation())
	}

	location := &Location{
		Latitude:  latitude,
		Longitude: longitude,
		Name:      name,
		PlaceID:   placeID,
	}

	return result.Ok(location)
}

// IsValid checks if the location has valid coordinates.
func (l *Location) IsValid() bool {
	return l.Latitude >= -90 && l.Latitude <= 90 &&
		l.Longitude >= -180 && l.Longitude <= 180
}
