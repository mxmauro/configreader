package helpers

import (
	"bytes"
)

// -----------------------------------------------------------------------------

func NormalizeEOL(data []byte) []byte {
	data = bytes.Replace(data, []byte("\r\n"), []byte("\n"), -1)
	data = bytes.Replace(data, []byte("\r"), []byte("\n"), -1)
	return data
}

func RemoveComments(data []byte) []byte {
	buf := new(bytes.Buffer)

	dataLen := len(data)

	for idx := 0; idx < dataLen; idx++ {
		var startIdx int
		// Loop
		state := 0
		for startIdx = idx; idx < dataLen && state < 101; idx++ {
			switch state {
			case 0: // Outside quotes and comments
				switch data[idx] {
				case '"':
					// Start of double-quoted value
					state = 1
					idx += 1

				case '\'':
					// Start of single-quoted value
					state = 2
					idx += 1

				case '/':
					if idx+1 < dataLen {
						switch data[idx+1] {
						case '/':
							state = 101 // Start of single-line comment

						case '*':
							state = 102 // Start of multi-line comment
						}
					}

				case '#':
					// Start of single line comment
					state = 103 // Start of single-line comment
				}

			case 1: // Inside double-quoted value
				switch data[idx] {
				case '\\':
					if idx+1 < dataLen {
						idx += 1 // Skip escaped character
					}
				case '"':
					state = 0 // End of quoted value
				}

			case 2: // Inside single-quoted value
				if data[idx] == '\'' {
					state = 0 // End of quoted value
				}
			}
		}

		// Add the buffer until here
		_, _ = buf.Write(data[startIdx:idx])

		// Skip comment
		switch state {
		case 101: // Single-line comment
			idx += 1
			fallthrough
		case 103: // Single-line comment starting with #
			idx += 1

			for ; idx < dataLen; idx += 1 {
				if data[idx] == '\n' {
					// End of comment
					idx += 1
					break
				}
			}

		case 4: // Multi-line comment
			for ; idx < dataLen; idx += 1 {
				if data[idx] == '*' && idx+1 < dataLen && data[idx+1] == '/' {
					// End of comment
					idx += 2
					break
				}
			}
		}
	}

	// Done
	return buf.Bytes()
}
