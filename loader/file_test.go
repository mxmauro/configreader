package loader_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/mxmauro/configreader"
	"github.com/mxmauro/configreader/internal/testhelpers"
	"github.com/mxmauro/configreader/loader"
)

//------------------------------------------------------------------------------

func TestFileLoader(t *testing.T) {
	// Create a new temporary file
	file, err := os.CreateTemp("", "cr")
	if err != nil {
		t.Fatalf("unable to create temporary file [err=%v]", err)
	}
	defer func() {
		_ = os.Remove(file.Name())
	}()

	// Save good settings on it
	_, err = file.Write([]byte(testhelpers.GoodSettingsJSON))
	if err != nil {
		t.Fatalf("unable to save good settings json [err=%v]", err)
	}

	// Load configuration from file
	var settings *testhelpers.TestSettings
	settings, err = configreader.New[testhelpers.TestSettings]().
		WithLoader(loader.NewFile().WithFilename(file.Name())).
		WithSchema(testhelpers.SchemaJSON).
		Load(context.Background())
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Check if settings are the expected
	if !reflect.DeepEqual(settings, &testhelpers.GoodSettings) {
		t.Fatalf("settings mismatch")
	}
}
