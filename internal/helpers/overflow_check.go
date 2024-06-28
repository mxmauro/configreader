package helpers

import (
	"math"
	"reflect"
)

// -----------------------------------------------------------------------------

func OverflowCheckInt64(t reflect.Type, value int64) bool {
	switch t.Kind() {
	case reflect.Int:
		if t.Size() == 4 {
			if value < math.MinInt32 || value > math.MaxInt32 {
				return false
			}
		}

	case reflect.Int8:
		if value < math.MinInt8 || value > math.MaxInt8 {
			return false
		}

	case reflect.Int16:
		if value < math.MinInt16 || value > math.MaxInt16 {
			return false
		}

	case reflect.Int32:
		if value < math.MinInt32 || value > math.MaxInt32 {
			return false
		}

	case reflect.Int64:

	case reflect.Uint:
		fallthrough
	case reflect.Uintptr:
		if value < 0 {
			return false
		}
		if t.Size() == 4 {
			if value > math.MaxUint32 {
				return false
			}
		}

	case reflect.Uint8:
		if value < 0 || value > math.MaxUint8 {
			return false
		}

	case reflect.Uint16:
		if value < 0 || value > math.MaxUint16 {
			return false
		}

	case reflect.Uint32:
		if value < 0 || value > math.MaxUint32 {
			return false
		}

	case reflect.Uint64:
		if value < 0 {
			return false
		}
	}
	return true
}

func OverflowCheckUint64(t reflect.Type, value uint64) bool {
	switch t.Kind() {
	case reflect.Int:
		if t.Size() == 4 {
			if value > math.MaxInt32 {
				return false
			}
		}

	case reflect.Int8:
		if value > math.MaxInt8 {
			return false
		}

	case reflect.Int16:
		if value > math.MaxInt16 {
			return false
		}

	case reflect.Int32:
		if value > math.MaxInt32 {
			return false
		}

	case reflect.Int64:
		if value > math.MaxInt64 {
			return false
		}

	case reflect.Uint:
		fallthrough
	case reflect.Uintptr:
		if t.Size() == 4 {
			if value > math.MaxUint32 {
				return false
			}
		}

	case reflect.Uint8:
		if value > math.MaxUint8 {
			return false
		}

	case reflect.Uint16:
		if value > math.MaxUint16 {
			return false
		}

	case reflect.Uint32:
		if value > math.MaxUint32 {
			return false
		}

	case reflect.Uint64:
	}
	return true
}
