package runctl

import "sync"

// flag symbols a busy process with the possibility
// to be called from multiple goroutines multiple times.
type Flag struct {
	bmutex sync.Mutex
	b      bool
}

// InProgress aquires the flag if not set. Safe for parallel use.
// The caller is responsible to call done in case no error is returned.
func (f *Flag) InProgress() (done func(), err error) {
	f.bmutex.Lock()
	if f.b {
		f.bmutex.Unlock()
		return nil, ErrInProgress
	}

	f.b = true
	f.bmutex.Unlock()
	return func() {
		f.bmutex.Lock()
		f.b = false
		f.bmutex.Unlock()
	}, nil
}
