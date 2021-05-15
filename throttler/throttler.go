// Package throttler defines a stream throttler.
// Its function is to limit the rate at which to pass output from one file descriptor to another.
package throttler

import "time"

// A Throttler is a stream throttler.
type Throttler interface {
	// Start starts up the throttler.
	Start() error

	// Stop shuts down the throttler.
	Stop() error

	// DoneRead indicates that there is no more data to be read into the throttler.
	DoneRead() error

	// Wait blocks until the throttler can write more data.
	Wait() error

	// Write writes the next chunk of data.
	Write([]byte) (int, error)
}

// Write writes a chunk of data to a throttler after waiting for a period of time,
// the throttler may perform additional waiting of its own.
func Write(t Throttler, b []byte, d time.Duration) error {
	if err := t.Wait(); err != nil {
		return err
	}
	time.Sleep(d)
	_, err := t.Write(b)
	return err
}
