// pkg/config/loader.go
package config

import (
	"fmt"
	"reflect"
	"time"
)

// Load populates a config struct from environment variables using reflection.
// The cfg parameter must be a pointer to a struct.
// Returns the same pointer on success.
func Load[T any](cfg T) (T, error) {
	return LoadWithPrefix(cfg, "")
}

// LoadWithPrefix is like Load but prepends a prefix to all environment variable names.
func LoadWithPrefix[T any](cfg T, prefix string) (T, error) {
	var zero T

	v := reflect.ValueOf(cfg)

	// Must be a pointer to a struct
	if v.Kind() != reflect.Ptr {
		return zero, fmt.Errorf("config must be a pointer to a struct, got %T", cfg)
	}

	v = v.Elem()
	if v.Kind() != reflect.Struct {
		return zero, fmt.Errorf("config must be a pointer to a struct, got pointer to %s", v.Kind())
	}

	t := v.Type()

	// Iterate through struct fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Get struct tags
		envTag := fieldType.Tag.Get("env")
		defaultTag := fieldType.Tag.Get("default")
		requiredTag := fieldType.Tag.Get("required")

		// Skip fields without env tag
		if envTag == "" {
			continue
		}

		// Build full env var name with prefix
		envKey := prefix + envTag

		// Get value from environment
		envValue, exists := GetEnvRequired(envKey)

		// Handle required fields
		if requiredTag == "true" {
			if !exists || envValue == "" {
				return zero, fmt.Errorf("required environment variable %s is not set", envKey)
			}
		}

		// Use default if not set
		if !exists || envValue == "" {
			envValue = defaultTag
		}

		// Skip if still empty
		if envValue == "" {
			continue
		}

		// Parse and set value based on field type
		if err := setField(field, envValue, envKey); err != nil {
			return zero, err
		}
	}

	return cfg, nil
}

// setField sets a struct field value by parsing the string value according to the field type
func setField(field reflect.Value, value string, envKey string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Special handling for time.Duration
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := ParseDuration(value)
			if err != nil {
				return fmt.Errorf("failed to parse %s as duration: %w", envKey, err)
			}
			field.SetInt(int64(d))
		} else {
			i, err := ParseInt64(value)
			if err != nil {
				return fmt.Errorf("failed to parse %s as integer: %w", envKey, err)
			}
			field.SetInt(i)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := ParseInt64(value)
		if err != nil || i < 0 {
			return fmt.Errorf("failed to parse %s as unsigned integer: %w", envKey, err)
		}
		field.SetUint(uint64(i))

	case reflect.Float32, reflect.Float64:
		f, err := ParseFloat(value)
		if err != nil {
			return fmt.Errorf("failed to parse %s as float: %w", envKey, err)
		}
		field.SetFloat(f)

	case reflect.Bool:
		b, err := ParseBool(value)
		if err != nil {
			return fmt.Errorf("failed to parse %s as boolean: %w", envKey, err)
		}
		field.SetBool(b)

	case reflect.Slice:
		if err := setSliceField(field, value, envKey); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported field type %s for %s", field.Kind(), envKey)
	}

	return nil
}

// setSliceField handles slice type fields
func setSliceField(field reflect.Value, value string, envKey string) error {
	switch field.Type().Elem().Kind() {
	case reflect.String:
		slice := ParseStringSlice(value)
		field.Set(reflect.ValueOf(slice))

	case reflect.Int:
		slice, err := ParseIntSlice(value)
		if err != nil {
			return fmt.Errorf("failed to parse %s as integer slice: %w", envKey, err)
		}
		field.Set(reflect.ValueOf(slice))

	default:
		return fmt.Errorf("unsupported slice type %s for %s", field.Type().Elem().Kind(), envKey)
	}

	return nil
}

// MustLoad is like Load but panics on error.
func MustLoad[T any](cfg T) T {
	result, err := Load(cfg)
	if err != nil {
		panic(err)
	}
	return result
}

// MustLoadWithPrefix is like LoadWithPrefix but panics on error.
func MustLoadWithPrefix[T any](cfg T, prefix string) T {
	result, err := LoadWithPrefix(cfg, prefix)
	if err != nil {
		panic(err)
	}
	return result
}
