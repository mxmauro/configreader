package configreader

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/mxmauro/configreader/internal/helpers"
	"github.com/mxmauro/configreader/model"
	"github.com/mxmauro/configreader/parsers"
)

// -----------------------------------------------------------------------------

var (
	typeOfTimeDuration = reflect.TypeOf(time.Duration(0))
)

// -----------------------------------------------------------------------------

func (cr *ConfigReader[T]) checkT() error {
	var instT T

	// Use reflection to inspect the configuration struct
	rSettings := reflect.ValueOf(&instT).Elem()
	if rSettings.Kind() != reflect.Struct {
		return errors.New("template parameter must be a struct")
	}

	// Done
	return nil
}

func (cr *ConfigReader[T]) fillFields(settings *T, values model.Values) error {
	rSettings := reflect.ValueOf(settings).Elem()
	return cr.fillFieldsRecursive(rSettings, "", values)
}

func (cr *ConfigReader[T]) fillFieldsRecursive(v reflect.Value, parentName string, values model.Values) error {
	// Ignore non-valid items
	if !v.IsValid() {
		return nil
	}

	// Ignore non-struct and pointers (single, double, ...) to non-structs
	if !isStructOrPtrToStruct(v) {
		return nil
	}

	// Allocate pointers if needed
	v = ptrAlloc(v)
	vType := v.Type()

	// Populate each field
	for fIdx := 0; fIdx < v.NumField(); fIdx++ {
		field := v.Field(fIdx)
		structField := vType.Field(fIdx)

		// Analyze field tags
		tags := structField.Tag

		// Get our tag
		configTag := tags.Get("config")
		if len(configTag) == 0 {
			// This field has no configuration but handle it if a struct or a pointer to one
			effField := ptrAlloc(field)
			if effField.Kind() == reflect.Struct {
				// Create struct object
				effField.Set(reflect.Zero(effField.Type()))

				// Go deeper
				err := cr.fillFieldsRecursive(effField, parentName+structField.Name+".", values)
				if err != nil {
					return err
				}
			}

			// Jump to next field
			continue
		}

		// Signal error if field cannot be written
		if !field.CanSet() {
			return fmt.Errorf("field \"%s%s\" is not settable", parentName, structField.Name)
		}

		// Get default value to use if no value is present
		defaultValue, defaultValuePresent := tags.Lookup("default")

		// Get type override
		typeOverride := tags.Get("type")

		// Check if this field should be treated as JSON
		isJson, isJsonPresent := helpers.Str2Bool(tags.Get("isjson"))
		if !isJsonPresent {
			isJson, isJsonPresent = helpers.Str2Bool(tags.Get("is_json"))
			if !isJsonPresent {
				isJson, isJsonPresent = helpers.Str2Bool(tags.Get("is-json"))
				if !isJsonPresent {
					isJson = false
				}
			}
		}

		// Get value to store
		vToSet, vToSetIsPresent := values[configTag]
		if !vToSetIsPresent {
			if defaultValuePresent {
				vToSet = defaultValue
			} else {
				vToSet = nil
			}
		}

		// Get effective field
		effField := ptrAlloc(field)
		effFieldIsPtr := field.Kind() == reflect.Pointer

		// Signal error if field cannot be written
		if !effField.CanSet() {
			return fmt.Errorf("field \"%s%s\" is not settable", parentName, structField.Name)
		}

		// Treat as JSON?
		if !isJson {
			// Special cases
			if typeOverride == "memory" {
				// Special case for memory type
				switch effField.Kind() {
				case reflect.Int, reflect.Int32, reflect.Int64:
					valueStr, isNil, ok := helpers.ToString(vToSet)
					if !ok {
						return newUnableToConvertFieldError(parentName, structField.Name)
					}
					if isNil {
						if !effFieldIsPtr {
							return newNoDefaultValueForFieldError(parentName, structField.Name)
						}
						effField.Set(reflect.Zero(effField.Type()))
					} else {
						m, err := parsers.ParseMemorySize(valueStr)
						if err != nil {
							return newUnableToConvertFieldError(parentName, structField.Name)
						}
						if helpers.OverflowCheckUint64(effField.Type(), m) {
							return newOverflowFieldError(parentName, structField.Name)
						}
						effField.SetInt(int64(m))
					}

				case reflect.Uint, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
					valueStr, isNil, ok := helpers.ToString(vToSet)
					if !ok {
						return newUnableToConvertFieldError(parentName, structField.Name)
					}
					if isNil {
						if !effFieldIsPtr {
							return newNoDefaultValueForFieldError(parentName, structField.Name)
						}
						effField.Set(reflect.Zero(effField.Type()))
					} else {
						m, err := parsers.ParseMemorySize(valueStr)
						if err != nil {
							return newUnableToConvertFieldError(parentName, structField.Name)
						}
						if helpers.OverflowCheckUint64(effField.Type(), m) {
							return newOverflowFieldError(parentName, structField.Name)
						}
						effField.SetUint(m)
					}

				default:
					return newUnableToConvertFieldError(parentName, structField.Name)
				}

			} else if effField.Type() == typeOfTimeDuration {

				//Special case for time duration fields
				switch effField.Kind() {
				case reflect.Int, reflect.Int32, reflect.Int64:
					valueStr, isNil, ok := helpers.ToString(vToSet)
					if !ok {
						return newUnableToConvertFieldError(parentName, structField.Name)
					}
					if isNil {
						if !effFieldIsPtr {
							return newNoDefaultValueForFieldError(parentName, structField.Name)
						}
						effField.Set(reflect.Zero(effField.Type()))
					} else {
						d, err := parsers.ParseDuration(valueStr)
						if err != nil {
							return newUnableToConvertFieldError(parentName, structField.Name)
						}
						effField.SetInt(int64(d))
					}

				case reflect.Uint, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
					valueStr, isNil, ok := helpers.ToString(vToSet)
					if !ok {
						return newUnableToConvertFieldError(parentName, structField.Name)
					}
					if isNil {
						if !effFieldIsPtr {
							return newNoDefaultValueForFieldError(parentName, structField.Name)
						}
						effField.Set(reflect.Zero(effField.Type()))
					} else {
						d, err := parsers.ParseDuration(valueStr)
						if err != nil || d < 0 {
							return newUnableToConvertFieldError(parentName, structField.Name)
						}
						effField.SetUint(uint64(d))
					}

				default:
					return newUnableToConvertFieldError(parentName, structField.Name)
				}

			} else {

				// For the rest of the field types
				switch effField.Kind() {
				case reflect.String:
					valueStr, isNil, ok := helpers.ToString(vToSet)
					if !ok {
						return newUnableToConvertFieldError(parentName, structField.Name)
					}
					if isNil {
						if !effFieldIsPtr {
							return newNoDefaultValueForFieldError(parentName, structField.Name)
						}
						effField.Set(reflect.Zero(effField.Type()))
					} else {
						effField.SetString(valueStr)
					}

				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					valueInt, isNil, overflow, ok := helpers.ToInt(vToSet, effField.Type().Bits())
					if !ok {
						return newUnableToConvertFieldError(parentName, structField.Name)
					}
					if isNil {
						if !effFieldIsPtr {
							return newNoDefaultValueForFieldError(parentName, structField.Name)
						}
						effField.Set(reflect.Zero(effField.Type()))
					} else if overflow {
						return newOverflowFieldError(parentName, structField.Name)
					} else {
						effField.SetInt(valueInt)
					}

				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
					valueUint, isNil, overflow, ok := helpers.ToUint(vToSet, effField.Type().Bits())
					if !ok {
						return newUnableToConvertFieldError(parentName, structField.Name)
					}
					if isNil {
						if !effFieldIsPtr {
							return newNoDefaultValueForFieldError(parentName, structField.Name)
						}
						effField.Set(reflect.Zero(effField.Type()))
					} else if overflow {
						return newOverflowFieldError(parentName, structField.Name)
					} else {
						effField.SetUint(valueUint)
					}

				case reflect.Float32, reflect.Float64:
					valueFloat, isNil, overflow, ok := helpers.ToFloat(vToSet)
					if !ok {
						return newUnableToConvertFieldError(parentName, structField.Name)
					}
					if isNil {
						if !effFieldIsPtr {
							return newNoDefaultValueForFieldError(parentName, structField.Name)
						}
						effField.Set(reflect.Zero(effField.Type()))
					} else if overflow {
						return newOverflowFieldError(parentName, structField.Name)
					} else {
						effField.SetFloat(valueFloat)
					}

				case reflect.Bool:
					valueBool, isNil, ok := helpers.ToBool(vToSet)
					if !ok {
						return newUnableToConvertFieldError(parentName, structField.Name)
					}
					if isNil {
						if !effFieldIsPtr {
							return newNoDefaultValueForFieldError(parentName, structField.Name)
						}
						effField.Set(reflect.Zero(effField.Type()))
					} else {
						effField.SetBool(valueBool)
					}

				case reflect.Struct:
					// Create struct object
					effField.Set(reflect.Zero(effField.Type()))

					// Go deeper
					err := cr.fillFieldsRecursive(effField, parentName+structField.Name+".", values)
					if err != nil {
						return err
					}

				case reflect.Slice:
					fallthrough
				case reflect.Array:
					if vToSet == nil {
						if !effFieldIsPtr {
							return newNoDefaultValueForFieldError(parentName, structField.Name)
						}
						effField.Set(reflect.Zero(effField.Type()))

						// Done
						break
					}

					// If the target a byte array/slice and the source a string, assume base64 encoded data
					if effField.Elem().Kind() == reflect.Uint8 && reflect.TypeOf(vToSet).Kind() == reflect.String {
						data, err := base64.StdEncoding.DecodeString(vToSet.(string))
						if err != nil {
							return newUnableToConvertFieldError(parentName, structField.Name)
						}

						// Check array size or create slice
						if effField.Kind() == reflect.Slice {
							l := len(data)
							effField.Set(reflect.MakeSlice(effField.Type(), l, l))
						} else {
							if len(data) != effField.Len() {
								return fmt.Errorf("base64 encoded data length mismatch for field \"%s%s\"", parentName, structField.Name)
							}
						}

						// Copy decoded data
						copy(effField.Bytes(), data)

						// Done
						break
					}

					// No other conversion supported for now
					return newUnableToConvertFieldError(parentName, structField.Name)

				default:
					return newUnableToConvertFieldError(parentName, structField.Name)
				}
			}

		} else {

			valueStr, isNil, ok := helpers.ToString(vToSet)
			if !ok {
				return newUnableToConvertFieldErrorJSON(parentName, structField.Name)
			}
			if isNil {
				if !effFieldIsPtr {
					return newUnableToConvertFieldErrorJSON(parentName, structField.Name)
				}
				effField.Set(reflect.Zero(effField.Type()))
			} else {
				iface := effField.Addr().Interface()

				// Unmarshal
				err := json.Unmarshal([]byte(valueStr), iface)
				if err != nil {
					return err
				}
			}

		}
	}

	// Done
	return nil
}

// -----------------------------------------------------------------------------

func ptrAlloc(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}

func isStructOrPtrToStruct(v reflect.Value) bool {
	vType := v.Type()
	for vType.Kind() == reflect.Pointer {
		vType = vType.Elem()
	}
	return vType.Kind() == reflect.Struct
}

func newUnableToConvertFieldError(parentName string, structFieldName string) error {
	return fmt.Errorf("unable to convert value for field \"%s%s\"", parentName, structFieldName)
}

func newNoDefaultValueForFieldError(parentName string, structFieldName string) error {
	return fmt.Errorf("no value specified nor default defined for field \"%s%s\"", parentName, structFieldName)
}

func newOverflowFieldError(parentName string, structFieldName string) error {
	return fmt.Errorf("overflow while assigning value to field \"%s%s\"", parentName, structFieldName)
}

func newUnableToConvertFieldErrorJSON(parentName string, structFieldName string) error {
	return fmt.Errorf("unable to convert value for json field \"%s%s\"", parentName, structFieldName)
}
