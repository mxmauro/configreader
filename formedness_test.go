package configreader_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/mxmauro/configreader"
	"github.com/mxmauro/configreader/internal/testhelpers"
	"github.com/mxmauro/configreader/loader"
)

// -----------------------------------------------------------------------------

func TestWellFormedWithSchema(t *testing.T) {
	// Load configuration
	settings, err := configreader.New[testhelpers.TestSettings]().
		WithLoader(loader.NewMemory().WithData(testhelpers.GoodSettingsMap)).
		Load(context.Background())
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !reflect.DeepEqual(settings, &testhelpers.GoodSettings) {
		t.Fatalf("settings mismatch")
	}
}

func TestMalformedWithSchema(t *testing.T) {
	// Load configuration
	_, err := configreader.New[testhelpers.TestSettings]().
		WithLoader(loader.NewMemory().WithData(testhelpers.BadSettingsMap)).
		Load(context.Background())
	if err == nil {
		t.Fatalf("unexpected success")
	}

	testhelpers.DumpValidationErrors(t, err)
}
