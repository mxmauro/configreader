package helpers

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/mxmauro/configreader/model"
)

// -----------------------------------------------------------------------------

const (
	maxEnvReplacementDepth = 4
)

// -----------------------------------------------------------------------------

var (
	ErrIsNil = errors.New("nil pointer")
)

// -----------------------------------------------------------------------------

// Expand expands variables on any interface object
func Expand(v interface{}, values model.Values) (interface{}, bool, error) {
	return expandRecursive(v, values, 1, 1)
}

// -----------------------------------------------------------------------------

func expandRecursive(v interface{}, values model.Values, childDepth int, expansionDepth int) (interface{}, bool, error) {
	if childDepth > maxEnvReplacementDepth {
		return "", false, errors.New("too many nested variable")
	}
	if expansionDepth > maxEnvReplacementDepth {
		return "", false, errors.New("too many nested environment variable expansion")
	}

	// Check if a nil interface or pointer to nil
	if v == nil {
		return v, false, nil
	}

	rV := reflect.ValueOf(v)
	switch rV.Kind() {
	case reflect.Pointer:
		pointedElem := rV.Elem()
		if pointedElem.IsValid() && pointedElem.CanInterface() && pointedElem.CanSet() {
			replacement, replaced, err := expandRecursive(pointedElem.Interface(), values, childDepth+1, expansionDepth)
			if err != nil {
				return "", false, err
			}
			if replaced {
				pointedElem.SetString(replacement.(string))
			}
		}

	case reflect.Array:
		fallthrough
	case reflect.Slice:
		for idx := 0; idx < rV.Len(); idx++ {
			elem := rV.Index(idx)
			if elem.IsValid() && elem.CanInterface() && elem.CanSet() {
				replacement, replaced, err := expandRecursive(elem.Interface(), values, childDepth+1, expansionDepth)
				if err != nil {
					return "", false, err
				}
				if replaced {
					elem.SetString(replacement.(string))
				}
			}
		}

	case reflect.String:
		// Analyze and expand string
		s := rV.String()

		sb := strings.Builder{}
		ofs := 0
		for {
			newVal := ""

			start := strings.Index(s[ofs:], "${")
			if start < 0 {
				if ofs == 0 {
					// No replacements were made
					return s, false, nil
				}

				// Else add remaining before ending the loop and return
				_, _ = sb.WriteString(s[ofs:])
				return sb.String(), true, nil
			}
			start += ofs

			end := strings.IndexByte(s[start+2:], '}')
			if end < 0 {
				return "", false, errors.New("variable closing tag not found")
			}
			end += start + 2

			// Add data before tag
			_, _ = sb.WriteString(s[ofs:start])

			//
			if end > start+2 {
				var replacement interface{}
				var isNil bool
				var ok bool
				var replaced bool
				var err error

				varName := s[start+2 : end]

				// Lookup replacement
				if values != nil {
					replacement, ok = values[varName]
				} else {
					replacement, ok = os.LookupEnv(strings.ToUpper(varName))
				}
				if !ok {
					return "", false, fmt.Errorf("variable \"%s\" not found", varName)
				}

				// Convert replacement to string if needed
				newVal, isNil, ok = ToString(replacement)
				if !ok {
					return "", false, fmt.Errorf("unable to convert non-string value of variable \"%s\"", varName)
				}
				if isNil {
					return "", false, ErrIsNil
				}

				// Recurse expansion
				replacement, replaced, err = expandRecursive(newVal, values, childDepth, expansionDepth+1)
				if err != nil {
					return "", false, err
				}
				if replaced {
					newVal = replacement.(string)
				}
			}
			// Else empty ${}, remove it

			// Add replacement
			_, _ = sb.WriteString(newVal)

			// Advance offset
			ofs = end + 1
		}

	default:
	}

	// Done
	return "", false, nil
}
