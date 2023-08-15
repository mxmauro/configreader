package helpers

import (
	"errors"
	"os"
	"reflect"
	"strings"
)

// -----------------------------------------------------------------------------

const (
	maxEnvReplacementDeph = 4
)

// -----------------------------------------------------------------------------

var (
	ErrIsNil = errors.New("nil pointer")
)

// -----------------------------------------------------------------------------

type EnumEnvAllowedValues struct {
	Name  string
	Value int
}

// -----------------------------------------------------------------------------

// GetEnv tries to load the content from an environment variable and, if not found, returns an error
func GetEnv(name string) (string, error) {
	value, ok := os.LookupEnv(name)
	if !ok {
		return "", errors.New("environment variable \"" + name + "\" not found")
	}
	return value, nil
}

// LoadAndReplaceEnvs expands environment variable references in the given text
func LoadAndReplaceEnvs(text string) (string, error) {
	return loadAndReplaceEnvsRecursive(text, 1)
}

func LoadAndReplaceEnvsByte(text []byte) ([]byte, error) {
	replaced, err := LoadAndReplaceEnvs(string(text))
	if err != nil {
		return nil, err
	}
	return []byte(replaced), nil
}

// GetBoolEnv returns a boolean value based on the provided input
func GetBoolEnv(value interface{}) (bool, error) {
	if value == nil {
		return false, ErrIsNil
	}
	valueOf := reflect.ValueOf(value)
	if valueOf.Kind() == reflect.Ptr {
		if valueOf.IsNil() {
			return false, ErrIsNil
		}
	}

	b := false

	switch v := value.(type) {
	case int:
		b = v != 0
	case uint:
		b = v != 0

	case int64:
		b = v != 0
	case uint64:
		b = v != 0

	case int32:
		b = v != 0
	case uint32:
		b = v != 0

	case bool:
		b = v

	case string:
		var err error

		v, err = LoadAndReplaceEnvs(v)
		if err != nil {
			return false, err
		}
		switch strings.ToLower(v) {
		case "1", "t", "true", "on", "yes":
			b = true
		}

	default:
		return false, errors.New("unsupported type")
	}

	// Done
	return b, nil
}

// GetEnumEnv returns an integer value based on the provided input
func GetEnumEnv(value interface{}, allowed []EnumEnvAllowedValues) (int, error) {
	if value == nil {
		return 0, ErrIsNil
	}
	valueOf := reflect.ValueOf(value)
	if valueOf.Kind() == reflect.Ptr {
		if valueOf.IsNil() {
			return 0, ErrIsNil
		}
		value = valueOf.Elem().Interface()
	}

	var i int

	switch v := value.(type) {
	case int:
		i = v
	case uint:
		i = int(v)

	case int64:
		i = int(v)
	case uint64:
		i = int(v)

	case int32:
		i = int(v)
	case uint32:
		i = int(v)

	case string:
		var err error

		v, err = LoadAndReplaceEnvs(v)
		if err != nil {
			return 0, err
		}
		v = strings.ToLower(v)

		for _, elem := range allowed {
			if v == elem.Name {
				return elem.Value, nil
			}
		}

		return 0, errors.New("invalid value")

	default:
		return 0, errors.New("unsupported type")
	}

	for _, elem := range allowed {
		if i == elem.Value {
			return i, nil
		}
	}

	// Done
	return 0, errors.New("invalid value")
}

func loadAndReplaceEnvsRecursive(text string, depth int) (string, error) {
	var replacement string
	var err error

	if depth > maxEnvReplacementDeph {
		return "", errors.New("too many nested environment variable expansion")
	}
	ofs := 0
	for {
		start := strings.IndexByte(text[ofs:], '%')
		if start < 0 {
			break
		}
		start += ofs

		end := strings.IndexByte(text[start+1:], '%')
		if end < 0 {
			return "", errors.New("environment variable closing tag not found")
		}
		end += start + 1
		if end == start+1 {
			// Double %, remove one
			replacement = "%"
		} else {
			// Lookup environment variable
			replacement, err = GetEnv(text[start+1 : end])
			if err == nil {
				replacement, err = loadAndReplaceEnvsRecursive(replacement, depth+1)
			}
			if err != nil {
				return "", err
			}
		}

		// Replace
		text = text[:start] + replacement + text[end+1:]
		ofs = start + len(replacement)
	}

	// Done
	return text, nil
}
