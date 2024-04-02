package testhelpers

import (
	"errors"
	"os"
	"testing"

	"github.com/mxmauro/configreader"
)

// -----------------------------------------------------------------------------

func ScopedEnvVars(vars map[string]string) func() {
	origVars := make(map[string]string)
	for k, v := range vars {
		origV, ok := os.LookupEnv(k)
		if ok {
			origVars[k] = "*" + origV
		} else {
			origVars[k] = ""
		}
		_ = os.Setenv(k, v)
	}

	return func() {
		for k, v := range origVars {
			if len(v) > 0 {
				_ = os.Setenv(k, v[1:])
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}
}

func DumpValidationErrors(t *testing.T, err error) {
	var vErr *configreader.ValidationError

	if errors.As(err, &vErr) {
		t.Logf("validation errors:")
		for _, f := range vErr.Failures {
			t.Logf("  %s", f.Error())
		}
	}
}
