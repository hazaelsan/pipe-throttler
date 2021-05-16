// Package expect implements a throttler that responds to another command's stdout/stderr.
package expect

import (
	"bufio"
	"errors"
	"io"
	"os"
	"os/exec"
	"time"
)

var (
	// ErrClosed is returned when the reader has already been closed.
	ErrClosed = errors.New("reader closed")

	// ErrNoCommand is returned when there's no command to execute.
	ErrNoCommand = errors.New("no command to execute")

	// ErrTimeout is returned when the read timeout is exceeded.
	ErrTimeout = errors.New("read timeout exceeded")
)

// Options is a set of options to instantiate an Expect throttler.
type Options struct {
	// Command is the wrapped command to execute.
	Command []string

	// MatchStderr indicates whether the wrapped command's stderr should be checked.
	// If unset then the wrapped command's stdout will be used.
	MatchStderr bool

	// SplitFunc is the function to use to split output from the wrapped command.
	SplitFunc bufio.SplitFunc

	// Timeout indicates how long to wait for the wrapped command to output matching text.
	// If <=0 then the throttler will wait indefinitely.
	Timeout time.Duration
}

// New instantiates an Expect throttler.
// Takes a regular expression to match and a command line to run.
func New(opts Options) (*Expect, error) {
	if len(opts.Command) == 0 {
		return nil, ErrNoCommand
	}
	return &Expect{
		opts:   opts,
		cmd:    exec.Command(opts.Command[0], opts.Command[1:]...),
		stdout: os.Stdout,
		stderr: os.Stderr,
		found:  make(chan struct{}, 1),
		errc:   make(chan error, 1),
	}, nil
}

// Expect is an output-based throttler.
type Expect struct {
	opts   Options
	cmd    *exec.Cmd
	r      io.ReadCloser
	w      io.WriteCloser
	s      *bufio.Scanner
	tee    io.Writer
	stdout io.Writer
	stderr io.Writer
	closed bool
	done   chan struct{}
	found  chan struct{}
	errc   chan error
}

func (e *Expect) setupCmd() error {
	var err error
	if e.w, err = e.cmd.StdinPipe(); err != nil {
		return err
	}
	if e.opts.MatchStderr {
		if e.r, err = e.cmd.StderrPipe(); err != nil {
			return err
		}
		e.cmd.Stdout = e.stdout
		e.tee = e.stderr
	} else {
		if e.r, err = e.cmd.StdoutPipe(); err != nil {
			return err
		}
		e.cmd.Stderr = e.stderr
		e.tee = e.stdout
	}
	return nil
}

// reader reads from the wrapped command's stdout/stderr,
// writes the buffer to *this* command's corresponding stdout/stderr and notifies `found`,
// notifies `errc` on error.
func (e *Expect) reader() {
	for e.s.Scan() {
		b := e.s.Bytes()
		if _, err := e.tee.Write(b); err != nil {
			e.errc <- err
			return
		}
		e.found <- struct{}{}
	}
	e.errc <- e.s.Err()
}

// Start starts up the throttler.
func (e *Expect) Start() error {
	if err := e.setupCmd(); err != nil {
		return err
	}
	e.s = bufio.NewScanner(e.r)
	e.s.Split(e.opts.SplitFunc)
	go e.reader()
	return e.cmd.Start()
}

// Stop shuts down the throttler.
func (e *Expect) Stop() error {
	return e.cmd.Wait()
}

// DoneRead indicates that there is no more data to be read into the throttler.
func (e *Expect) DoneRead() error {
	return e.w.Close()
}

// Wait blocks until the wrapped program's output matches the expected string
// or the timeout is exceeded.
func (e *Expect) Wait() error {
	if e.closed {
		return ErrClosed
	}
	if e.opts.Timeout <= 0 {
		select {
		case <-e.found:
			return nil
		case err := <-e.errc:
			e.closed = true
			return err
		}
	}
	select {
	case <-e.found:
		return nil
	case err := <-e.errc:
		e.closed = true
		return err
	case <-time.After(e.opts.Timeout):
		return ErrTimeout
	}
}

// Write writes the next chunk of data to the wrapped program's stdin.
func (e *Expect) Write(b []byte) (int, error) {
	return e.w.Write(b)
}
