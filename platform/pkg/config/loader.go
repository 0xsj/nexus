package config

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Load loads configuration into a struct using default options.
func Load[T any]() (*T, error) {
	return LoadWithOptions[T](DefaultOptions())
}

// LoadWithOptions loads configuration into a struct with custom options.
func LoadWithOptions[T any](opts Options) (*T, error) {
	const op = "config.Load"

	var cfg T

	// Load .env file if specified
	if opts.EnvFile != "" {
		if err := LoadEnvFileIfExists(opts.EnvFile); err != nil {
			return nil, err
		}
	}

	// Load configuration from sources
	for _, source := range opts.Sources {
		switch source {
		case SourceDefault:
			if err := loadDefaults(&cfg, opts); err != nil {
				return nil, err
			}
		case SourceEnv:
			if err := loadFromEnv(&cfg, opts); err != nil {
				return nil, err
			}
		case SourceFile:
			if opts.ConfigFile != "" {
				if err := loadFromFile(&cfg, opts.ConfigFile); err != nil {
					return nil, err
				}
			}
		}
	}

	// Validate required fields
	if len(opts.Required) > 0 {
		if err := validateRequired(&cfg, opts.Required); err != nil {
			return nil, err
		}
	}

	// Validate if the config implements Validator
	if v, ok := any(&cfg).(Validator); ok {
		if err := v.Validate(); err != nil {
			return nil, ErrConfigInvalid(op, err.Error())
		}
	}

	return &cfg, nil
}

// Validator is implemented by config structs that need custom validation.
type Validator interface {
	Validate() error
}

// ============================================================================
// Default Loading
// ============================================================================

// loadDefaults sets default values from struct tags.
func loadDefaults(cfg any, opts Options) error {
	return walkStruct(cfg, opts, func(field reflect.Value, tag structTag) error {
		if tag.defaultValue != "" && isZero(field) {
			return setFieldValue(field, tag.defaultValue)
		}
		return nil
	})
}

// ============================================================================
// Environment Loading
// ============================================================================

// loadFromEnv loads configuration from environment variables.
func loadFromEnv(cfg any, opts Options) error {
	return walkStruct(cfg, opts, func(field reflect.Value, tag structTag) error {
		envKey := tag.envKey
		if envKey == "" {
			return nil
		}

		// Add prefix if specified
		if opts.EnvPrefix != "" {
			envKey = opts.EnvPrefix + "_" + envKey
		}

		value, exists := os.LookupEnv(envKey)
		if !exists {
			return nil
		}

		// Expand environment variables if enabled
		if opts.ExpandEnv {
			value = os.ExpandEnv(value)
		}

		// Check for empty values
		if value == "" && !opts.AllowEmpty {
			return nil
		}

		return setFieldValue(field, value)
	})
}

// ============================================================================
// File Loading (placeholder for JSON/YAML support)
// ============================================================================

// loadFromFile loads configuration from a file.
func loadFromFile(cfg any, path string) error {
	const op = "config.loadFromFile"

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ErrConfigNotFound(op, path)
	}

	// Determine file type and load accordingly
	// For now, this is a placeholder - can be extended for JSON/YAML
	return ErrConfigInvalid(op, "file loading not yet implemented")
}

// ============================================================================
// Validation
// ============================================================================

// validateRequired checks that required fields are set.
func validateRequired(cfg any, required []string) error {
	const op = "config.validateRequired"

	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	for _, fieldName := range required {
		field := v.FieldByName(fieldName)
		if !field.IsValid() {
			return ErrConfigRequired(op, fieldName)
		}
		if isZero(field) {
			return ErrConfigRequired(op, fieldName)
		}
	}

	return nil
}

// ============================================================================
// Struct Walking
// ============================================================================

// structTag holds parsed struct tag information.
type structTag struct {
	envKey       string
	defaultValue string
	required     bool
}

// walkFunc is called for each field during struct walking.
type walkFunc func(field reflect.Value, tag structTag) error

// walkStruct walks a struct and calls fn for each field.
func walkStruct(cfg any, opts Options, fn walkFunc) error {
	v := reflect.ValueOf(cfg)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	return walkStructValue(v, opts, "", fn)
}

// walkStructValue recursively walks struct fields.
func walkStructValue(v reflect.Value, opts Options, prefix string, fn walkFunc) error {
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		// Parse struct tags
		tag := parseStructTag(fieldType, opts.Tag)

		// Build prefixed key for nested structs
		if prefix != "" && tag.envKey != "" {
			tag.envKey = prefix + "_" + tag.envKey
		} else if prefix != "" && tag.envKey == "" {
			tag.envKey = prefix + "_" + ToEnvKey(fieldType.Name)
		}

		// Handle nested structs
		if field.Kind() == reflect.Struct && fieldType.Type != reflect.TypeOf(time.Time{}) {
			nestedPrefix := tag.envKey
			if nestedPrefix == "" {
				nestedPrefix = prefix
				if nestedPrefix != "" {
					nestedPrefix += "_"
				}
				nestedPrefix += ToEnvKey(fieldType.Name)
			}
			if err := walkStructValue(field, opts, nestedPrefix, fn); err != nil {
				return err
			}
			continue
		}

		// Handle pointers to structs
		if field.Kind() == reflect.Ptr && field.Type().Elem().Kind() == reflect.Struct {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			if err := walkStructValue(field.Elem(), opts, tag.envKey, fn); err != nil {
				return err
			}
			continue
		}

		// Call the function for this field
		if err := fn(field, tag); err != nil {
			return err
		}
	}

	return nil
}

// parseStructTag parses struct tags for a field.
func parseStructTag(field reflect.StructField, tagName string) structTag {
	tag := structTag{}

	// Parse env tag
	if envTag := field.Tag.Get(tagName); envTag != "" {
		parts := strings.Split(envTag, ",")
		tag.envKey = parts[0]

		for _, part := range parts[1:] {
			if part == "required" {
				tag.required = true
			}
		}
	}

	// Parse default tag
	if defaultTag := field.Tag.Get("default"); defaultTag != "" {
		tag.defaultValue = defaultTag
	}

	// If no env tag, derive from field name
	if tag.envKey == "" {
		tag.envKey = ToEnvKey(field.Name)
	}

	return tag
}

// ============================================================================
// Value Setting
// ============================================================================

// setFieldValue sets a field value from a string.
func setFieldValue(field reflect.Value, value string) error {
	const op = "config.setFieldValue"

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Handle time.Duration specially
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			d, err := time.ParseDuration(value)
			if err != nil {
				return ErrConfigParseError(op, field.Type().Name(), err)
			}
			field.SetInt(int64(d))
			return nil
		}

		i, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return ErrConfigParseError(op, field.Type().Name(), err)
		}
		field.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return ErrConfigParseError(op, field.Type().Name(), err)
		}
		field.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return ErrConfigParseError(op, field.Type().Name(), err)
		}
		field.SetFloat(f)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return ErrConfigParseError(op, field.Type().Name(), err)
		}
		field.SetBool(b)

	case reflect.Slice:
		if field.Type().Elem().Kind() == reflect.String {
			parts := strings.Split(value, ",")
			slice := make([]string, 0, len(parts))
			for _, part := range parts {
				trimmed := strings.TrimSpace(part)
				if trimmed != "" {
					slice = append(slice, trimmed)
				}
			}
			field.Set(reflect.ValueOf(slice))
		}

	case reflect.Ptr:
		// Handle pointer types
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		return setFieldValue(field.Elem(), value)

	default:
		// Try to handle time.Time
		if field.Type() == reflect.TypeOf(time.Time{}) {
			t, err := parseTime(value)
			if err != nil {
				return ErrConfigParseError(op, field.Type().Name(), err)
			}
			field.Set(reflect.ValueOf(t))
			return nil
		}

		return ErrConfigTypeError(op, field.Type().Name(), "supported type", field.Kind().String())
	}

	return nil
}

// isZero returns true if a value is the zero value for its type.
func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Slice, reflect.Map:
		return v.IsNil() || v.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Struct:
		if v.Type() == reflect.TypeOf(time.Time{}) {
			return v.Interface().(time.Time).IsZero()
		}
		return false
	default:
		return false
	}
}

// parseTime attempts to parse a time string in various formats.
func parseTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, &parseError{msg: "unable to parse time: " + s}
}
