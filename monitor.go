package configreader

import (
	"bytes"
	"errors"
	"sync"
	"time"

	"github.com/mxmauro/channelcontext"
)

// -----------------------------------------------------------------------------

// Monitor implements a module that periodically checks if the configuration setting changed.
type Monitor[T any] struct {
	mtx        sync.Mutex
	attachedCr *ConfigReader[T]

	pollInterval time.Duration
	callback     SettingsChangedCallback[T]

	wg     sync.WaitGroup
	stopCh chan struct{}

	settingsHash [64]byte
}

// -----------------------------------------------------------------------------

// NewMonitor creates a new configuration settings change monitor
func NewMonitor[T any](pollInterval time.Duration, callback SettingsChangedCallback[T]) *Monitor[T] {
	return &Monitor[T]{
		pollInterval: pollInterval,
		callback:     callback,
		wg:           sync.WaitGroup{},
	}
}

// Destroy stops a running configuration settings monitor and destroys it
func (m *Monitor[T]) Destroy() {
	if m.stopCh != nil {
		close(m.stopCh)
	}

	m.wg.Wait()

	m.stopCh = nil
}

func (m *Monitor[T]) start(cr *ConfigReader[T], settingsHash [64]byte) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()

	// Check if the configuration reader is already attached to this monitor
	if m.attachedCr == cr {
		copy(m.settingsHash[:], settingsHash[:]) // Just update the encoded settings hash
		return nil
	}

	// Check if the monitor is attached to another configuration reader
	if m.attachedCr != nil {
		return errors.New("monitor already attached to another configuration reader")
	}

	// Attach and create a copy of the encoded settings hash
	m.attachedCr = cr
	copy(m.settingsHash[:], settingsHash[:])

	// Start polling
	m.stopCh = make(chan struct{})
	m.wg.Add(1)
	go m.worker()

	// Done
	return nil
}

func (m *Monitor[T]) worker() {
	defer m.wg.Done()

MainLoop:
	for {
		select {
		case <-m.stopCh:
			break MainLoop

		case <-time.After(m.pollInterval):
			if m.doReload() {
				break MainLoop
			}
		}
	}

	// Detach configuration reader
	m.mtx.Lock()
	m.attachedCr = nil
	m.mtx.Unlock()
}

func (m *Monitor[T]) doReload() bool {
	stopCtx, cancelStopCtx := channelcontext.New[struct{}](m.stopCh)
	defer cancelStopCtx()

	// Load the whole data
	settings, settingsHash, err := m.attachedCr.load(stopCtx)
	if err != nil {
		select {
		case <-stopCtx.Done():
			return true
		default:
			m.callback(nil, err)
			return false
		}
	}

	changed := false

	m.mtx.Lock()

	// If encoded settings are the same, do nothing
	if !bytes.Equal(settingsHash[:], m.settingsHash[:]) {
		copy(m.settingsHash[:], settingsHash[:])
		changed = true
	}

	m.mtx.Unlock()

	// Call the callback if settings changed
	if changed {
		m.callback(settings, nil)
	}

	// Do nothing
	return false
}
