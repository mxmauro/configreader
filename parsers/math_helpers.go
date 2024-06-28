package parsers

import (
	"math"
)

// -----------------------------------------------------------------------------

func addInt64WithOverflow(a, b int64) (int64, error) {
	if (b > 0 && a > math.MaxInt64-b) || (b < 0 && a < math.MinInt64-b) {
		return 0, errOverflow
	}
	return a + b, nil
}

func addUint64WithOverflow(a, b uint64) (uint64, error) {
	if a > math.MaxUint64-b {
		return 0, errOverflow
	}
	return a + b, nil
}

func mulUint64WithOverflow(a, b uint64) (uint64, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	if a > math.MaxUint64/b {
		return 0, errOverflow
	}
	return a * b, nil
}

func mulFloat64WithOverflow(a, b float64) (float64, error) {
	result := a * b
	if math.IsInf(result, 0) {
		return result, errOverflow
	}
	return result, nil
}
