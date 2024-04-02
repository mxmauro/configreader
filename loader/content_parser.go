package loader

import (
	"encoding/json"
	"errors"

	"github.com/mxmauro/configreader/internal/helpers"
	"github.com/mxmauro/configreader/model"
)

// -----------------------------------------------------------------------------

const (
	parseDataHintIsDotEnv = 0x0001
)

// -----------------------------------------------------------------------------

func parseData(data []byte, hint int) (model.Values, error) {
	if data == nil {
		return make(model.Values), nil
	}

	// Normalize content
	data = helpers.NormalizeEOL(data)

	if (hint&parseDataHintIsDotEnv) == 0 || isDotEnv(data) {
		// Assume a .env file
		return parseDotEnv(data)
	}

	data = helpers.RemoveComments(data)

	if isJSON(data) {
		// Assume a value JSON file
		var ret model.Values

		err := json.Unmarshal(data, &ret)
		if err != nil {
			return nil, err
		}

		// Done
		return ret, nil
	}

	// Unable to identify file type
	return nil, errors.New("unable to identify file type")
}

func isJSON(data []byte) bool {
	// Try to guess a valid JSON
	for idx := 0; idx < len(data); idx++ {
		if data[idx] == '[' || data[idx] == '{' {
			return true
		}
		if data[idx] != ' ' && data[idx] != '\t' && data[idx] != '\r' && data[idx] != '\n' {
			break
		}
	}
	return false
}
