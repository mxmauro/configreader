package helpers

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// -----------------------------------------------------------------------------

func ToString(v interface{}) (value string, isNil bool, ok bool) {
	if v == nil {
		isNil = true
		ok = true
		return
	}
	rV := dereference(v)
	if rV.Kind() == reflect.Pointer && rV.IsNil() {
		isNil = true
		ok = true
		return
	}

	switch rV.Kind() {
	case reflect.String:
		value = rV.String()

	case reflect.Bool:
		if v.(bool) {
			value = "true"
		} else {
			value = "false"
		}

	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		value = fmt.Sprintf("%v", rV.Int())

	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uintptr:
		value = fmt.Sprintf("%v", rV.Uint())

	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		value = fmt.Sprintf("%f", rV.Float())

	case reflect.Struct:
		fallthrough
	case reflect.Slice:
		fallthrough
	case reflect.Array:
		encoded, err := json.Marshal(v)
		if err == nil {
			value = string(encoded)
		}

	default:
		return
	}

	// Done
	ok = true
	return
}

func ToBool(v interface{}) (value bool, isNil bool, ok bool) {
	if v == nil {
		isNil = true
		ok = true
		return
	}
	rV := dereference(v)
	if rV.Kind() == reflect.Pointer && rV.IsNil() {
		isNil = true
		ok = true
		return
	}

	switch rV.Kind() {
	case reflect.String:
		value, ok = Str2Bool(rV.String())
		if !ok {
			return
		}

	case reflect.Bool:
		value = rV.Bool()

	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		value = rV.Int() == 0

	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uintptr:
		value = rV.Uint() == 0

	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		f := rV.Float()
		value = f >= 0.0000001 || f <= -0.0000001

	default:
		return
	}

	// Done
	ok = true
	return
}

func ToInt(v interface{}, bitSize int) (value int64, isNil bool, overflow bool, ok bool) {
	if v == nil {
		isNil = true
		ok = true
		return
	}
	rV := dereference(v)
	if rV.Kind() == reflect.Pointer && rV.IsNil() {
		isNil = true
		ok = true
		return
	}

	switch rV.Kind() {
	case reflect.String:
		var err error

		value, err = strconv.ParseInt(rV.String(), 10, 64)
		if err != nil {
			return
		}

	case reflect.Bool:
		if v.(bool) {
			value = 1
		}

	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		value = rV.Int()
		if OverflowCheckInt64(rV.Type(), value) {
			overflow = true
			return
		}

	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uintptr:
		valU := rV.Uint()
		if OverflowCheckUint64(rV.Type(), valU) {
			overflow = true
			return
		}
		value = int64(valU)

	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		f := rV.Float()
		if f < float64(math.MinInt64) || f > float64(math.MaxInt64) {
			overflow = true
			return
		}
		value = int64(f)

	default:
	}

	// Check overflow
	if value < -1<<(bitSize-1) || value > (1<<(bitSize-1))-1 {
		overflow = true
		return
	}

	// Done
	ok = true
	return
}

func ToUint(v interface{}, bitSize int) (value uint64, isNil bool, overflow bool, ok bool) {
	if v == nil {
		isNil = true
		ok = true
		return
	}
	rV := dereference(v)
	if rV.Kind() == reflect.Pointer && rV.IsNil() {
		isNil = true
		ok = true
		return
	}

	switch rV.Kind() {
	case reflect.String:
		var err error

		value, err = strconv.ParseUint(rV.String(), 10, 64)
		if err != nil {
			return
		}

	case reflect.Bool:
		if rV.Bool() {
			value = 1
		}

	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		valI := rV.Int()
		if OverflowCheckInt64(rV.Type(), valI) {
			overflow = true
			return
		}
		value = uint64(valI)

	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uintptr:
		value = rV.Uint()
		if OverflowCheckUint64(rV.Type(), value) {
			overflow = true
			return
		}

	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		f := rV.Float()
		if f < -0.0000001 || f > float64(math.MaxUint64) {
			overflow = true
			return
		}
		if f >= 0.0000001 {
			value = uint64(f)
		}

	default:
		return
	}

	// Check overflow
	if value > (1<<(bitSize-1))-1 {
		overflow = true
		return
	}

	// Done
	ok = true
	return
}

func ToFloat(v interface{}) (value float64, isNil bool, overflow bool, ok bool) {
	if v == nil {
		isNil = true
		ok = true
		return
	}
	rV := dereference(v)
	if rV.Kind() == reflect.Pointer && rV.IsNil() {
		isNil = true
		ok = true
		return
	}

	switch rV.Kind() {
	case reflect.String:
		var err error

		value, err = strconv.ParseFloat(rV.String(), 64)
		if err != nil {
			return
		}

	case reflect.Bool:
		if v.(bool) {
			value = 1.0
		}

	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		value = float64(rV.Int())

	case reflect.Uint:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		fallthrough
	case reflect.Uintptr:
		value = float64(rV.Uint())

	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		value = rV.Float()

	default:
		return
	}

	// Done
	ok = true
	return
}

func Str2Bool(s string) (value bool, ok bool) {
	switch strings.Trim(strings.ToLower(s), " \t") {
	case "0":
		fallthrough
	case "no":
		fallthrough
	case "false":
		fallthrough
	case "f":
		fallthrough
	case "n":
		value = false
		ok = true
	case "1":
		fallthrough
	case "yes":
		fallthrough
	case "true":
		fallthrough
	case "t":
		fallthrough
	case "y":
		value = true
		ok = true
	}
	return
}

// -----------------------------------------------------------------------------

func dereference(v interface{}) reflect.Value {
	rV := reflect.ValueOf(v)
	for rV.Kind() == reflect.Pointer {
		if rV.IsNil() {
			break
		}
		rV = rV.Elem()
	}
	return rV
}
