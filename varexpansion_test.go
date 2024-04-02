package configreader_test

import (
	"context"
	"testing"

	"github.com/mxmauro/configreader"
	"github.com/mxmauro/configreader/loader"
)

// -----------------------------------------------------------------------------

type VariableExpansionTest struct {
	Url string `config:"TEST_URL"`
}

// -----------------------------------------------------------------------------

func TestVariableExpansion(t *testing.T) {
	// Load configuration from data stream source
	settings, err := configreader.New[VariableExpansionTest]().
		WithLoader(loader.NewMemory().WithData(map[string]interface{}{
			"TEST_URL":      "mongodb://${TEST_USERNAME}:${TEST_PASSWORD}@${TEST_HOST}:${TEST_PORT}/${TEST_DATABASE}?replSet=rs0",
			"TEST_HOST":     "127.0.0.1",
			"TEST_PORT":     27017,
			"TEST_USERNAME": "user",
			"TEST_PASSWORD": "pass",
			"TEST_DATABASE": "sample_database",
		})).
		Load(context.Background())
	if err != nil {
		t.Fatalf(err.Error())
	}

	// Check if settings are the expected
	if settings.Url != "mongodb://user:pass@127.0.0.1:27017/sample_database?replSet=rs0" {
		t.Fatalf("settings mismatch")
	}
}
