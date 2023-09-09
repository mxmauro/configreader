package configreader

import (
	"context"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"time"

	"github.com/mxmauro/configreader/internal/helpers"
	"github.com/mxmauro/configreader/loader"
)

// -----------------------------------------------------------------------------

// ConfigReader contains configurable loader options.
type ConfigReader[T any] struct {
	loader            loader.Loader
	schema            string
	extendedValidator ExtendedValidator[T]
	noReplaceEnvVars  bool

	reloader struct {
		timeout               time.Duration
		callback              SettingsChangedCallback[T]
		stopCh                chan struct{}
		doneCh                chan struct{}
		hashOfEncodedSettings [64]byte
	}

	err error
}

// ExtendedValidator is a function to call in order to do configuration validation not covered by this library.
type ExtendedValidator[T any] func(settings *T) error

// SettingsChangedCallback is a function to call when the reloader detects a change in the configuration settings.
type SettingsChangedCallback[T any] func(settings *T, loadErr error)

//------------------------------------------------------------------------------

// New creates a new configuration reader
func New[T any]() *ConfigReader[T] {
	return &ConfigReader[T]{}
}

// WithLoader sets the content loader
func (cr *ConfigReader[T]) WithLoader(l loader.Loader) *ConfigReader[T] {
	if cr.err == nil {
		cr.loader = l
	}
	return cr
}

// WithSchema sets an optional JSON schema validator
func (cr *ConfigReader[T]) WithSchema(schema string) *ConfigReader[T] {
	if cr.err == nil {
		cr.schema = schema
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

// WithNoReplaceEnvVars stops the loader from replacing environment variables that can be found inside
func (cr *ConfigReader[T]) WithNoReplaceEnvVars() *ConfigReader[T] {
	if cr.err == nil {
		cr.noReplaceEnvVars = true
	}
	return cr
}

// WithReload sets a polling interval and a callback to call if the configuration settings changes
func (cr *ConfigReader[T]) WithReload(pollInterval time.Duration, callback SettingsChangedCallback[T]) *ConfigReader[T] {
	if cr.err == nil {
		cr.reloader.timeout = pollInterval
		cr.reloader.callback = callback
	}
	return cr
}

// Load settings from the specified source
func (cr *ConfigReader[T]) Load(ctx context.Context) (*T, error) {
	var hashOfEncodedSettings [64]byte
	var settings *T
	var err error

	// If an error was set by a With... function, return it
	if cr.err != nil {
		return nil, cr.err
	}

	// If no context was specified, use a default
	if ctx == nil {
		ctx = context.Background()
	}

	// Check if a loader was specified
	if cr.loader == nil {
		return nil, newConfigLoadError(errors.New("loader not defined"))
	}

	// Load the whole data
	settings, hashOfEncodedSettings, err = cr.load(ctx)
	if err != nil {
		return nil, err
	}

	// Start re-loader goroutine if provided
	if cr.reloader.callback != nil && cr.reloader.timeout > 0 {
		cr.startReloadPoller(hashOfEncodedSettings)
	}

	// Done
	return settings, nil
}

// Destroy destroys the configuration reader. Used mainly to stop reload goroutine.
func (cr *ConfigReader[T]) Destroy() {
	cr.stopReloadPoller()
}

func (cr *ConfigReader[T]) load(ctx context.Context) (*T, [64]byte, error) {
	var settings *T

	// Load the whole data
	encodedSettings, err := cr.loader.Load(ctx)
	if err != nil {
		return nil, [64]byte{}, newConfigLoadError(err)
	}

	// Replace environment variables inside the resulting json
	if !cr.noReplaceEnvVars {
		encodedSettings, err = helpers.LoadAndReplaceEnvsByte(encodedSettings)
		if err != nil {
			return nil, [64]byte{}, newConfigLoadError(err)
		}
	}

	// Remove comments from json
	encodedSettings = removeComments(encodedSettings)

	// If resulting configuration is empty, throw error
	if len(encodedSettings) == 0 {
		return nil, [64]byte{}, newConfigLoadError(errors.New("empty data"))
	}

	// Do final validation and decoding
	err = cr.validate(encodedSettings)
	if err != nil {
		return nil, [64]byte{}, newConfigLoadError(err)
	}

	// Do final validation and decoding
	settings, err = cr.decode(encodedSettings)
	if err != nil {
		return nil, [64]byte{}, newConfigLoadError(err)
	}

	// Done
	return settings, sha512.Sum512(encodedSettings), nil
}

func (cr *ConfigReader[T]) decode(encodedSettings []byte) (*T, error) {
	var settings T

	// Parse configuration settings json object
	err := json.Unmarshal(encodedSettings, &settings)
	if err != nil {
		return nil, err
	}

	// Execute the extended validation if one was specified
	if cr.extendedValidator != nil {
		err = cr.extendedValidator(&settings)
		if err != nil {
			return nil, err
		}
	}

	// Done
	return &settings, nil
}
