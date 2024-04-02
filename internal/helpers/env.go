package helpers

import (
	"errors"
)

// -----------------------------------------------------------------------------

type EnumEnvAllowedValues struct {
	Name  string
	Value int
}

// -----------------------------------------------------------------------------

// ExpandEnvVars expands environment variables
func ExpandEnvVars(s string) (string, error) {
	v, replaced, err := Expand(s, nil)
	if err != nil {
		return "", err
	}
	if !replaced {
		return s, nil
	}
	return v.(string), nil
}

// GetBoolEnv returns a boolean value based on the provided input
func GetBoolEnv(v interface{}) (bool, error) {
	replacement, replaced, err := Expand(v, nil)
	if err != nil {
		return false, err
	}
	if replaced {
		v = replacement
	}

	value, isNil, ok := ToBool(v)
	if !ok {
		return false, errors.New("unsupported type")
	}
	if isNil {
		return false, ErrIsNil
	}

	// Done
	return value, nil
}

// GetEnumEnv returns an integer value based on the provided input
func GetEnumEnv(v interface{}, allowed []EnumEnvAllowedValues) (int, error) {
	var valueInt int64
	var overflow bool

	replacement, replaced, err := Expand(v, nil)
	if err != nil {
		return 0, err
	}
	if replaced {
		v = replacement
	}

	valueStr, isNil, ok := ToString(v)
	if ok {
		if isNil {
			return 0, ErrIsNil
		}
		for _, elem := range allowed {
			if valueStr == elem.Name {
				return elem.Value, nil
			}
		}
		return 0, errors.New("invalid value")
	}

	valueInt, isNil, overflow, ok = ToInt(v, 64)
	if ok {
		if isNil {
			return 0, ErrIsNil
		}
		if !overflow {
			for _, elem := range allowed {
				if valueInt == int64(elem.Value) {
					return elem.Value, nil
				}
			}
		}
		return 0, errors.New("invalid value")
	}

	return 0, errors.New("unsupported type")
}
