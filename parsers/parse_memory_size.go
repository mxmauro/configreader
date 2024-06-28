package parsers

import (
	"errors"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/v3/mem"
)

// -----------------------------------------------------------------------------

const (
	memUnitKB uint64 = 1000
	memUnitMB        = 1000 * memUnitKB
	memUnitGB        = 1000 * memUnitMB
	memUnitTB        = 1000 * memUnitGB
	memUnitPB        = 1000 * memUnitTB

	memUnitKiB uint64 = 1024
	memUnitMiB        = 1024 * memUnitKiB
	memUnitGiB        = 1024 * memUnitMiB
	memUnitTiB        = 1024 * memUnitGiB
	memUnitPiB        = 1024 * memUnitTiB
)

// -----------------------------------------------------------------------------

var textUnitMap = map[string]uint64{
	"B":   1,
	"KB":  memUnitKB,
	"MB":  memUnitMB,
	"GB":  memUnitGB,
	"TB":  memUnitTB,
	"PB":  memUnitPB,
	"KiB": memUnitKiB,
	"MiB": memUnitMiB,
	"GiB": memUnitGiB,
	"TiB": memUnitTiB,
	"PiB": memUnitPiB,
}

// -----------------------------------------------------------------------------

// ParseMemorySize parses the human-readable size string into the amount it represents.
func ParseMemorySize(s string) (uint64, error) {
	if strings.HasSuffix(s, "%") {
		return parseMemoryPct(s[:len(s)-1])
	}
	return parseMemoryFixedValue(s)
}

// -----------------------------------------------------------------------------

func parseMemoryFixedValue(s string) (uint64, error) {
	var idx int
	var multiplier uint64 = 1

	l := len(s)

	// Parse value
	dotPresent := false
	for idx = 0; idx < l; idx += 1 {
		if s[idx] == '.' {
			if dotPresent {
				return 0, errInvalidValue
			}
			dotPresent = true
		} else if s[idx] < '0' || s[idx] > '9' {
			break
		}
	}

	// Split integer and fraction parts
	intPart := s[:idx]
	if len(intPart) == 0 || intPart == "." {
		return 0, errInvalidValue
	}
	fracPart := ""
	dotIdx := strings.Index(intPart, ".")
	if dotIdx > 0 {
		fracPart = strings.TrimRight(intPart[(dotIdx+1):], "0")
		intPart = intPart[:dotIdx]
	}
	intPart = strings.TrimLeft(intPart, "0")

	// All zero?
	if len(intPart) == 0 && len(fracPart) == 0 {
		return 0, nil
	}

	// Skip spaces
	for idx < l && s[idx] == ' ' {
		idx += 1
	}

	// Parse unit multiplier, if any
	if idx < l {
		found := false
		unit := s[idx:]
		for k, v := range textUnitMap {
			if k == unit {
				multiplier = v
				found = true
				break
			}
		}
		if !found {
			return 0, errInvalidUnit
		}
	}

	// Convert the integer part
	value := uint64(0)
	if len(intPart) > 0 {
		var err error

		value, err = strconv.ParseUint(intPart, 10, 64)
		if err != nil {
			if errors.Is(err, strconv.ErrRange) {
				return 0, errOverflow
			}
			return 0, errInvalidValue
		}

		// And multiply by the unit
		value, err = mulUint64WithOverflow(value, multiplier)
		if err != nil {
			return 0, err
		}
	}

	// Convert the fractional part
	if len(fracPart) > 0 {
		// Make safe for float, 10 digits is enough
		if len(fracPart) > 10 {
			fracPart = fracPart[:10]
		}

		fracValue, err := strconv.ParseFloat("0."+fracPart, 64)
		if err != nil {
			if errors.Is(err, strconv.ErrRange) {
				return 0, errOverflow
			}
			return 0, errInvalidValue
		}
		value, err = addUint64WithOverflow(value, uint64(fracValue*float64(multiplier)))
		if err != nil {
			return 0, err
		}
	}

	// Done
	return value, nil
}

func parseMemoryPct(s string) (uint64, error) {
	var v *mem.VirtualMemoryStat
	var result uint64

	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil || n < 1 || n > 100 {
		return 0, errors.New("invalid percentage value")
	}
	v, err = mem.VirtualMemory()
	if err != nil {
		return 0, errors.New("unable to get virtual memory size")
	}

	// Try multiplication first because more precise
	if result, err = mulUint64WithOverflow(v.Total, n); err == nil {
		return result / 100, nil
	}
	return (v.Total / 100) * n, nil
}
