package loader_test

import (
	"bytes"
	"context"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/mxmauro/configreader"
	"github.com/mxmauro/configreader/internal/testhelpers"
	"github.com/mxmauro/configreader/loader"
)

//------------------------------------------------------------------------------

func TestEnvironmentVariable(t *testing.T) {
	// Save test environment variable and restore on exit
	defer testhelpers.ScopedEnvVar("GO_CONFIGREADER_TEST")()

	// Save the data stream into test environment variable
	_ = os.Setenv("GO_CONFIGREADER_TEST", testhelpers.GoodSettingsJSON)

	// Load configuration from data stream
	settings, err := configreader.New[testhelpers.TestSettings]().
		WithLoader(loader.NewMemoryFromEnvironmentVariable("GO_CONFIGREADER_TEST")).
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

func TestComplexVariableExpansion(t *testing.T) {
	// Save test environment variables and restore on exit
	defer testhelpers.ScopedEnvVar("GO_CONFIGREADER_MONGODB_USER")()
	defer testhelpers.ScopedEnvVar("GO_CONFIGREADER_MONGODB_PASSWORD")()
	defer testhelpers.ScopedEnvVar("GO_CONFIGREADER_MONGODB_DATABASE")()
	defer testhelpers.ScopedEnvVar("GO_CONFIGREADER_MONGODB_HOST")()
	defer testhelpers.ScopedEnvVar("GO_CONFIGREADER_MONGODB_URL")()

	// Find a known setting and replace with a data source reference
	toReplace := "mongodb://user:pass@127.0.0.1:27017/sample_database?replSet=rs0"
	pos := strings.Index(testhelpers.GoodSettingsJSON, toReplace)
	if pos < 0 {
		t.Fatalf("unexpected string find failure")
	}

	// Create the modified version of the settings json
	modifiedSettingsJSON := bytes.Join([][]byte{
		([]byte(testhelpers.GoodSettingsJSON))[0:pos],
		[]byte("%GO_CONFIGREADER_MONGODB_URL%"),
		([]byte(testhelpers.GoodSettingsJSON))[pos+len(toReplace):],
	}, nil)

	// Setup our complex url which includes a source and environment variables (some embedded).
	_ = os.Setenv("GO_CONFIGREADER_MONGODB_URL", "mongodb://%GO_CONFIGREADER_MONGODB_USER%:%GO_CONFIGREADER_MONGODB_PASSWORD%@%GO_CONFIGREADER_MONGODB_HOST%/%GO_CONFIGREADER_MONGODB_DATABASE%?replSet=rs0")

	// Save the credentials in environment variables
	_ = os.Setenv("GO_CONFIGREADER_MONGODB_USER", "user")
	_ = os.Setenv("GO_CONFIGREADER_MONGODB_PASSWORD", "pass")

	// Also, the host and database name
	_ = os.Setenv("GO_CONFIGREADER_MONGODB_HOST", "127.0.0.1:27017")
	_ = os.Setenv("GO_CONFIGREADER_MONGODB_DATABASE", "sample_database")

	// Load configuration from data stream source
	settings, err := configreader.New[testhelpers.TestSettings]().
		WithLoader(loader.NewMemory().WithData(string(modifiedSettingsJSON))).
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
