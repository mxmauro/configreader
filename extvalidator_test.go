package configreader_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mxmauro/configreader"
	"github.com/mxmauro/configreader/loader"
	"github.com/mxmauro/configreader/model"
)

// -----------------------------------------------------------------------------

var ExtendedValidatorValueError = errors.New("value must be greater than 0")

type ExtendedValidatorTest struct {
	Value int `config:"TEST_VALUE"`
}

// -----------------------------------------------------------------------------

func TestValidExtendedValidator(t *testing.T) {
	// Load configuration
	err := testExtendedValidatorCommon(model.Values{
		"TEST_VALUE": 1,
	})
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestInvalidExtendedValidator(t *testing.T) {
	// Load configuration
	err := testExtendedValidatorCommon(model.Values{
		"TEST_VALUE": 0,
	})
	if err == nil {
		t.Fatalf("unexpected success")
	} else if !errors.Is(err, ExtendedValidatorValueError) {
		t.Fatalf("unexpected error [err=%v]", err)
	}
}

// -----------------------------------------------------------------------------

func testExtendedValidatorCommon(data model.Values) error {
	_, err := configreader.New[ExtendedValidatorTest]().
		WithLoader(loader.NewMemory().WithData(data)).
		WithExtendedValidator(func(settings *ExtendedValidatorTest) error {
			if settings.Value < 1 {
				return ExtendedValidatorValueError
			}
			return nil
		}).
		Load(context.Background())

	return err
}
