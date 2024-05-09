package configreader

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/gob"
	"slices"

	"github.com/mxmauro/configreader/internal/helpers"
	"github.com/mxmauro/configreader/loader"
	"github.com/mxmauro/configreader/model"
)

// -----------------------------------------------------------------------------

// ConfigReader contains configurable loader options.
type ConfigReader[T any] struct {
	loader                []model.Loader
	extendedValidator     ExtendedValidator[T]
	disableEnvVarOverride []string

	monitor *Monitor[T]

	err error
}

// ExtendedValidator is a function to call in order to do configuration validation not covered by this library.
type ExtendedValidator[T any] func(settings *T) error

// SettingsChangedCallback is a function to call when the re-loader detects a change in the configuration settings.
type SettingsChangedCallback[T any] func(settings *T, loadErr error)

// New creates a new configuration reader
func New[T any]() *ConfigReader[T] {
	cfg := ConfigReader[T]{}

	// Check if T is a struct
	err := cfg.checkT()
	if err != nil {
		panic(err)
	}

	// Done
	return &cfg
}

// WithLoader sets the content loader
func (cr *ConfigReader[T]) WithLoader(l ...model.Loader) *ConfigReader[T] {
	if cr.err == nil {
		if cr.loader == nil {
			cr.loader = make([]model.Loader, 0)
		}
		cr.loader = append(cr.loader, l...)
	}
	return cr
}

// WithExtendedValidator sets an optional settings validator callback
func (cr *ConfigReader[T]) WithExtendedValidator(validator ExtendedValidator[T]) *ConfigReader[T] {
	if cr.err == nil {
		cr.extendedValidator = validator
	}
	return cr
}

// WithDisableEnvVarOverride stops the loader from replacing environment variables that can be found inside
func (cr *ConfigReader[T]) WithDisableEnvVarOverride(vars ...string) *ConfigReader[T] {
	if cr.err == nil {
		if cr.disableEnvVarOverride == nil {
			cr.disableEnvVarOverride = make([]string, 0)
		}
		cr.disableEnvVarOverride = append(cr.disableEnvVarOverride, vars...)
	}
	return cr
}

// WithMonitor sets a monitor that will inform about configuration settings changes
//
// NOTE: First send a value to stop monitoring and then wait until a value is received on the same channel
func (cr *ConfigReader[T]) WithMonitor(m *Monitor[T]) *ConfigReader[T] {
	if cr.err == nil {
		cr.monitor = m
	}
	return cr
}

// Load settings from the specified source
func (cr *ConfigReader[T]) Load(ctx context.Context) (*T, error) {
	// If an error was set by a With... function, return it
	if cr.err != nil {
		return nil, cr.err
	}

	// If no context was specified, use a default
	if ctx == nil {
		ctx = context.Background()
	}

	// Load the whole data
	settings, settingsHash, err := cr.load(ctx)
	if err != nil {
		return nil, err
	}

	// Start re-loader goroutine if provided
	if cr.monitor != nil {
		err = cr.monitor.start(cr, settingsHash)
		if err != nil {
			return nil, err
		}
	}

	// Done
	return settings, nil
}

func (cr *ConfigReader[T]) load(ctx context.Context) (*T, [64]byte, error) {
	var hash [64]byte

	// Load the whole data
	values, err := cr.loadValues(ctx)
	if err != nil {
		return nil, hash, err
	}

	// Decode settings
	settings := new(T)
	err = cr.fillFields(settings, values)
	if err != nil {
		return nil, hash, err
	}

	// Validate settings
	err = cr.validate(settings)
	if err != nil {
		return nil, hash, err
	}

	// Calculate hash
	hash, err = cr.calcHash(settings)
	if err != nil {
		return nil, hash, err
	}

	// Done
	return settings, hash, nil
}

func (cr *ConfigReader[T]) loadValues(ctx context.Context) (model.Values, error) {
	// Load the whole data
	mergedValues := make(model.Values)
	for _, l := range cr.loader {
		values, err := l.Load(ctx)
		if err != nil {
			return nil, err
		}
		for k, v := range values {
			mergedValues[k] = v
		}
	}

	// Merge environment variables
	for k, v := range loader.GetEnvVars() {
		if !slices.Contains(cr.disableEnvVarOverride, k) {
			mergedValues[k] = v
		}
	}

	// Expand nested expressions
	for k, v := range mergedValues {
		replacement, replaced, err := helpers.Expand(v, mergedValues)
		if err != nil {
			return nil, err
		}
		if replaced {
			mergedValues[k] = replacement
		}
	}

	// Done
	return mergedValues, nil
}

func (cr *ConfigReader[T]) calcHash(setting *T) ([64]byte, error) {
	buf := bytes.Buffer{}
	encoder := gob.NewEncoder(&buf)

	// Encode
	err := encoder.Encode(setting)
	if err != nil {
		return [64]byte{}, err
	}

	// Calculate hash
	return sha512.Sum512(buf.Bytes()), nil
}
