package configreader_test

import (
	"context"
	"errors"
	"testing"

	"github.com/mxmauro/configreader"
	"github.com/mxmauro/configreader/loader"
)

//------------------------------------------------------------------------------

var ExtendedValidatorValueError = errors.New("value must be greater than 0")

type ExtendedValidatorTest struct {
	Value int `json:"value"`
}

//------------------------------------------------------------------------------

func TestValidExtendedValidator(t *testing.T) {
	// Load configuration
	err := testExtendedValidatorCommon(`{ "value": 1 }`)
	if err != nil {
		t.Fatalf(err.Error())
	}
}

func TestInvalidExtendedValidator(t *testing.T) {
	// Load configuration
	err := testExtendedValidatorCommon(`{ "value": 0 }`)
	if err == nil {
		t.Fatalf("unexpected success")
	} else if !errors.Is(err, ExtendedValidatorValueError) {
		t.Fatalf("unexpected error [err=%v]", err)
	}
}

func testExtendedValidatorCommon(data string) error {
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
