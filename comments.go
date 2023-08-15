package configreader

// -----------------------------------------------------------------------------

func removeComments(data []byte) {
	state := 0

	dataLength := len(data)

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
						state = 2         // Start of single line comment
						data[index] = ' ' //Remove comment start
						index += 1
						data[index] = ' ' //Remove comment start

					case '*':
						state = 3         // Start of multiple line comment
						data[index] = ' ' //Remove comment start
						index += 1
						data[index] = ' ' //Remove comment start
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
				state = 0 // End of single line comment
			} else {
				data[index] = ' ' // Remove comment
			}

		case 3: // Multi line comment
			if ch == '*' && index+1 < dataLength && data[index+1] == '/' {
				state = 0         // End of multi line comment
				data[index] = ' ' // Remove comment end
				index += 1
				data[index] = ' ' // Remove comment end
			} else {
				data[index] = ' ' // Remove comment
			}
		}
	}
}
