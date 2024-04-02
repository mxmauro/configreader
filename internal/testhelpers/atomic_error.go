package testhelpers

import (
	"sync"
)

// -----------------------------------------------------------------------------

type AtomicError struct {
	mtx sync.Mutex
	err error
}

// -----------------------------------------------------------------------------

func NewAtomicError() *AtomicError {
	return &AtomicError{
		mtx: sync.Mutex{},
	}
}

func (ae *AtomicError) Set(err error) {
	ae.mtx.Lock()
	defer ae.mtx.Unlock()

	if ae.err == nil && err != nil {
		ae.err = err
	}
}

func (ae *AtomicError) Err() error {
	ae.mtx.Lock()
	defer ae.mtx.Unlock()

	return ae.err
}
