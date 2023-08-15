package configreader

import (
	"bytes"
	"context"
	"time"
)

// -----------------------------------------------------------------------------

func (cr *ConfigReader[T]) startReloadPoller(encodedSettings []byte) {
	// Create a copy of the encoded settings
	cr.reloader.encodedSettings = make([]byte, len(encodedSettings))
	copy(cr.reloader.encodedSettings, encodedSettings)

	// Start polling
	cr.reloader.stopCh = make(chan struct{})
	go cr.reloadPoll()
}

func (cr *ConfigReader[T]) stopReloadPoller() {
	if cr.reloader.stopCh != nil {
		cr.reloader.stopCh <- struct{}{}
		<-cr.reloader.doneCh
		close(cr.reloader.stopCh)
		close(cr.reloader.doneCh)
	}
}

func (cr *ConfigReader[T]) reloadPoll() {
	for {
		select {
		case <-cr.reloader.stopCh:
			cr.reloader.doneCh <- struct{}{}
			return

		case <-time.After(cr.reloader.timeout):
			shutdown := false
			reloadCompletedCh := make(chan struct{})
			ctx, ctxCancel := context.WithCancel(context.Background())

			// Start a goroutine that will cancel the context if we receive the shutdown signal
			go func() {
				select {
				case <-cr.reloader.stopCh:
					// Cancel the context on receipt of the shutdown signal
					ctxCancel()
					shutdown = true

				case <-ctx.Done():
				}
			}()

			// Execute reload on background
			go func() {
				cr.reload(ctx)
				reloadCompletedCh <- struct{}{}
			}()

			// Wait until reload completes (or canceled)
			<-reloadCompletedCh
			ctxCancel()
			close(reloadCompletedCh)

			// If we were told to stop, then stop
			if shutdown {
				cr.reloader.doneCh <- struct{}{}
				return
			}
		}
	}
}

func (cr *ConfigReader[T]) reload(ctx context.Context) {
	// Load the whole data
	settings, encodedSettings, err := cr.load(ctx)
	if err != nil {
		cr.reloader.callback(nil, err)
		return
	}

	// If encoded settings are the same, do nothing
	if bytes.Equal(encodedSettings, cr.reloader.encodedSettings) {
		return
	}

	// Preserve new encoded settings for future comparison
	cr.reloader.encodedSettings = encodedSettings

	// Call the callback
	cr.reloader.callback(settings, nil)
}
