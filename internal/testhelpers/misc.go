package testhelpers

import (
	"errors"
	"os"
	"testing"

	"github.com/mxmauro/configreader"
)

//------------------------------------------------------------------------------

func ScopedEnvVar(varName string) func() {
	origValue := os.Getenv(varName)
	return func() {
		_ = os.Setenv(varName, origValue)
	}
}

func DumpValidationErrors(t *testing.T, err error) {
	var vErr *configreader.ValidationError

	if errors.As(err, &vErr) {
		t.Logf("validation errors:")
		for _, f := range vErr.Failures {
			t.Logf("  %v @ %v", f.Message, f.Location)
		}
	}
}
