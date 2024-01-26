package exporter

import "sync"

type ErrorHandler struct {
	lock      sync.Mutex
	lastError error
}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

func (e *ErrorHandler) Handle(err error) {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.lastError = err
}

func (e *ErrorHandler) Error() error {
	e.lock.Lock()
	defer e.lock.Unlock()

	return e.lastError
}

func (e *ErrorHandler) Clear() {
	e.lock.Lock()
	defer e.lock.Unlock()

	e.lastError = nil
}
