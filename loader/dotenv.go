package loader

import (
	"bytes"
	"errors"
	"strconv"
	"strings"

	"github.com/mxmauro/configreader/model"
)

// -----------------------------------------------------------------------------

func isDotEnv(data []byte) bool {
	eolIdx := bytes.IndexByte(data, '\n')
	if eolIdx < 0 {
		eolIdx = len(data)
	}
	_, _, err := parseDotEnvLine(data[:eolIdx])
	return err == nil
}

func parseDotEnv(data []byte) (model.Values, error) {
	ret := make(model.Values)

	for len(data) > 0 {
		// Look for the end of line
		eolIdx := bytes.IndexByte(data, '\n')
		if eolIdx < 0 {
			eolIdx = len(data)
		}

		varName, varValue, err := parseDotEnvLine(data[:eolIdx])
		if err != nil {
			return nil, err
		}

		// Add to values map
		if len(varName) > 0 {
			ret[varName] = varValue
		}

		// Continue in next line
		if eolIdx < len(data) && data[eolIdx] == '\n' {
			eolIdx += 1
		}
		data = data[eolIdx:]
	}

	// Done
	return ret, nil
}

func parseDotEnvLine(data []byte) (varName string, varValue string, err error) {
	// Skip spaces
	data = skipSpaces(data)

	// A comment or empty line
	if len(data) == 0 || data[0] == '#' {
		return "", "", nil
	}

	// Export prefix?
	if len(data) >= 7 && bytes.Equal(data[:6], []byte("export")) && (data[6] == ' ' || data[6] == '\t') {
		// Skip it
		data = skipSpaces(data[7:])
	}

	// Get variable name
	sepIdx := bytes.IndexByte(data, '=')
	sepIdx2 := bytes.IndexByte(data, ':')
	if sepIdx < 0 || (sepIdx2 >= 0 && sepIdx2 < sepIdx) {
		sepIdx = sepIdx2
	}
	if sepIdx <= 0 || sepIdx >= len(data)-1 {
		err = errors.New("invalid environment variable name")
		return
	}
	varName = strings.Trim(string(data[:sepIdx]), " \t")
	if len(varName) == 01 {
		err = errors.New("invalid environment variable name")
		return
	}

	// Skip spaces after separator
	data = skipSpaces(data[sepIdx+1:])

	// Get value
	buf := new(bytes.Buffer)
	currQuote := byte(0)
	idx := 0
	dataLen := len(data)
ValueLoop:
	for idx < dataLen {
		switch data[idx] {
		case '"':
			fallthrough
		case '\'':
			if currQuote == 0 {
				currQuote = data[0]
			} else if currQuote == data[0] {
				currQuote = 0
			} else {
				_ = buf.WriteByte(data[idx])
			}
			idx += 1

		case '\\':
			if currQuote == '"' && idx+1 < dataLen {
				idx += 2
				switch data[idx-1] {
				case 'x':
					fallthrough
				case 'u':
					size := 2
					if data[idx-1] == 'u' {
						size = 4
					}
					if idx+size <= dataLen {
						if num, err2 := strconv.ParseInt(string(data[idx:idx+size]), 16, 32); err2 == nil && num != 0 {
							_, _ = buf.WriteString(string(rune(num)))
							idx += size
						} else {
							idx = dataLen
						}
					} else {
						idx = dataLen
					}

				case 'o':
					size := 0
					for idx+size < dataLen && data[idx+size] >= '0' && data[idx+size] <= '7' {
						size += 1
					}
					if size > 0 {
						if num, err2 := strconv.ParseInt(string(data[idx:idx+size]), 8, 32); err2 == nil && num != 0 {
							_, _ = buf.WriteString(string(rune(num)))
						}
					}
					idx += size

				default:
					code := bytes.IndexByte([]byte("abtnvfr\"'\\"), data[idx-1])
					switch {
					case code >= 0 && code <= 6:
						_ = buf.WriteByte(byte(7 + code))
					case code == 7:
						_, _ = buf.Write([]byte{34})
					case code == 8:
						_, _ = buf.Write([]byte{39})
					case code == 9:
						_, _ = buf.Write([]byte{92})
					}
				}
			} else {
				_ = buf.WriteByte(data[idx])
				idx += 1
			}

		case ' ':
			fallthrough
		case '\t':
			if currQuote == 0 {
				break ValueLoop
			}
			fallthrough

		default:
			if data[idx] != 0 {
				_ = buf.WriteByte(data[idx])
			}
			idx += 1
		}
	}
	varValue = string(buf.Bytes())

	// Skip spaces
	data = skipSpaces(data[idx:])

	// Sanity check
	if len(data) != 0 && data[0] != '#' {
		err = errors.New("invalid environment variable value")
	}

	// Done
	return
}

func skipSpaces(data []byte) []byte {
	for idx := 0; idx < len(data); idx++ {
		if data[idx] != ' ' && data[idx] != '\t' {
			return data[idx:]
		}
	}
	return data[:0]
}
