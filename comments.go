package configreader

// -----------------------------------------------------------------------------

func removeComments(data []byte) []byte {
	var commentStartIndex int

	dataLength := len(data)

	state := 0
	for index := 0; index < dataLength; index++ {
		ch := data[index]

		switch state {
		case 0: // Outside quotes and comments
			if ch == '"' {
				state = 1 // Start of quoted value
			} else if ch == '/' {
				if index+1 < dataLength {
					switch data[index+1] {
					case '/':
						state = 2 // Start of single line comment
						commentStartIndex = index
						index += 1

					case '*':
						state = 3 // Start of multiple line comment
						commentStartIndex = index
						index += 1
					}
				}
			}

		case 1: // Inside quoted value
			if ch == '\\' {
				if index+1 < dataLength {
					index += 1 // Escaped character
				}
			} else if ch == '"' {
				state = 0 // End of quoted value
			}

		case 2: // Single line comment
			if ch == '\n' {
				// End of single line comment
				state = 0
				copy(data[commentStartIndex:], data[index+1:])
				dataLength -= (index + 1) - commentStartIndex
			}

		case 3: // Multi line comment
			if ch == '*' && index+1 < dataLength && data[index+1] == '/' {
				// End of multi line comment
				state = 0
				index += 1
				copy(data[commentStartIndex:], data[index+1:])
				dataLength -= (index + 1) - commentStartIndex
			}
		}
	}

	// Done
	return data
}
