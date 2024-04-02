package loader_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/mxmauro/configreader"
	"github.com/mxmauro/configreader/internal/testhelpers"
)

// -----------------------------------------------------------------------------

func TestEnvironmentVariable(t *testing.T) {
	// Save test environment variable and restore on exit
	defer testhelpers.ScopedEnvVars(testhelpers.ToStringStringMap(testhelpers.GoodSettingsMap))()

	// Load configuration from data stream
	settings, err := configreader.New[testhelpers.TestSettings]().
		Load(context.Background())
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Check if settings are the expected
	if !reflect.DeepEqual(settings, &testhelpers.GoodSettings) {
		t.Fatalf("settings mismatch")
	}
}
