package loader_test

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/mxmauro/configreader"
	"github.com/mxmauro/configreader/internal/testhelpers"
	"github.com/mxmauro/configreader/loader"
)

// -----------------------------------------------------------------------------

func TestFileLoader(t *testing.T) {
	// Create a new temporary file
	file, err := os.CreateTemp("", "cr")
	if err != nil {
		t.Fatalf("unable to create temporary file [err=%v]", err)
	}
	defer func() {
		filename := file.Name()

		_ = file.Close()
		_ = os.Remove(filename)
	}()

	// Save good settings on it
	for k, v := range testhelpers.ToStringStringMap(testhelpers.GoodSettingsMap) {
		_, err = file.WriteString(fmt.Sprintf("%s=%s\n", k, testhelpers.QuoteValue(v)))
		if err != nil {
			t.Fatalf("unable to save good settings in a file [err=%v]", err)
		}
	}

	// Load configuration from file
	var settings *testhelpers.TestSettings
	settings, err = configreader.New[testhelpers.TestSettings]().
		WithLoader(loader.NewFile().WithFilename(file.Name())).
		Load(context.Background())
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Check if settings are the expected
	if !reflect.DeepEqual(settings, &testhelpers.GoodSettings) {
		t.Fatalf("settings mismatch")
	}
}
