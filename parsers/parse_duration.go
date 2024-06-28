package parsers

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// -----------------------------------------------------------------------------

// ParseDuration parses a human-readable duration string into the amount it represents.
// CLARIFICATION: time.ParseDuration does not support days, so this function augments its behavior.
func ParseDuration(s string) (time.Duration, error) {
	if len(s) == 0 {
		return 0, errInvalidDuration
	}

	accum := int64(0)

	// Check if days were specified
	dayPos := strings.Index(s, "d")
	if dayPos >= 0 {
		var daysFragmentF float64

		// Get the whole day part
		dayStart := dayPos - 1
		for dayStart >= 0 {
			if (s[dayStart] < '0' || s[dayStart] > '9') && s[dayStart] != '.' && s[dayStart] != '+' && s[dayStart] != '-' {
				break
			}
			dayStart -= 1
		}
		dayStart += 1
		dayPos += 1

		// Strip from the original string
		dayPart := s[dayStart:dayPos]
		s = s[:dayStart] + s[dayPos:]

		// Convert days to a number
		if len(dayPart) < 1 {
			return 0, errInvalidDuration
		}
		days, err := strconv.ParseFloat(dayPart[:len(dayPart)-1], 64)
		if err != nil {
			if errors.Is(err, strconv.ErrRange) {
				return 0, errOverflow
			}
			return 0, errInvalidDuration
		}

		daysFragmentF, err = mulFloat64WithOverflow(days, float64(24*time.Hour))
		if err != nil {
			return 0, err
		}
		accum = int64(daysFragmentF)
	}

	// Parse the rest without days
	if len(s) > 0 {
		d, err := time.ParseDuration(s)
		if err != nil {
			return 0, errInvalidDuration
		}

		accum, err = addInt64WithOverflow(int64(d), accum)
		if err != nil {
			return 0, err
		}
	}

	// Done
	return time.Duration(accum), nil
}
