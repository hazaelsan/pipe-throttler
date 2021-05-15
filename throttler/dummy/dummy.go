// Package dummy implements a dummy throttler.
// It depends on the runner doing any actual throttling.
package dummy

import "io"

// New instantiates a dummy throttler.
func New(w io.WriteCloser) *Dummy {
	return &Dummy{w}
}

// A Dummy is a dummy throttler.
type Dummy struct {
	w io.WriteCloser
}

// Start starts up the throttler.
func (*Dummy) Start() error {
	return nil
}

// Stop shuts down the throttler.
func (d *Dummy) Stop() error {
	return d.w.Close()
}

// DoneRead indicates that there is no more data to be read into the throttler.
func (*Dummy) DoneRead() error {
	return nil
}

// Wait is a no-op for this throttler.
func (*Dummy) Wait() error {
	return nil
}

// Write writes the next chunk of data to the underlying Writer.
func (d *Dummy) Write(b []byte) (int, error) {
	return d.w.Write(b)
}
