package loader

import (
	"os"
	"strings"

	"github.com/mxmauro/configreader/model"
)

// -----------------------------------------------------------------------------

func GetEnvVars() model.Values {
	ret := make(model.Values)

	for _, envVar := range os.Environ() {
		sepIdx := strings.Index(envVar, "=")
		if sepIdx <= 0 {
			continue
		}

		varValue := ""
		if sepIdx < len(envVar)-1 {
			varValue = envVar[sepIdx+1:]
		}

		// Add to values map
		ret[envVar[:sepIdx]] = varValue

	}

	// Done
	return ret
}
