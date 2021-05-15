// Package runner handles reading and writing to/from file descriptors.
package runner

import (
	"bufio"
	"io"
	"sync"
	"time"

	"github.com/hazaelsan/pipe-throttler/throttler"
)

// Options is a set of options to initialize a Runner.
type Options struct {
	// Reader is the input source for bytes to write.
	Reader io.Reader

	// Throttler is the output throttler to rate-limit writes.
	Throttler throttler.Throttler

	// SplitFunc is the function used to split output from the wrapped program.
	SplitFunc bufio.SplitFunc

	// WaitDuration is how long to wait after the Throttler has indicated
	// it's ready before writing the next chunk of data.
	WaitDuration time.Duration
}

// New initializes a Runner.
func New(opts Options) *Runner {
	r := &Runner{
		s:    bufio.NewScanner(opts.Reader),
		t:    opts.Throttler,
		wait: opts.WaitDuration,
	}
	r.s.Split(opts.SplitFunc)
	return r
}

// A Runner handles reading and writing to/from file descriptors.
type Runner struct {
	s    *bufio.Scanner
	t    throttler.Throttler
	wait time.Duration
	wg   sync.WaitGroup
}

// Run copies bytes from the source reader to the throttled destination.
func (r *Runner) Run() error {
	if err := r.t.Start(); err != nil {
		return err
	}
	c := make(chan []byte)
	errc := make(chan error)
	go r.reader(c, errc)
	go r.writer(c, errc)
	if err := <-errc; err != nil {
		r.t.Stop()
		return err
	}
	r.wg.Wait()
	return r.t.Stop()
}

func (r *Runner) reader(c chan<- []byte, errc chan<- error) {
	r.wg.Add(1)
	defer r.wg.Done()
	defer close(c)
	for r.s.Scan() {
		c <- r.s.Bytes()
	}
	if err := r.s.Err(); err != nil {
		errc <- err
	}
}

func (r *Runner) writer(c <-chan []byte, errc chan<- error) {
	r.wg.Add(1)
	defer r.wg.Done()
	defer r.t.DoneRead()
	for {
		b, ok := <-c
		if !ok {
			errc <- nil
			return
		}
		if err := throttler.Write(r.t, b, r.wait); err != nil {
			errc <- err
			return
		}
	}
}
