package expect

import (
	"bufio"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hazaelsan/pipe-throttler/split"
)

const (
	wantStdout = "stdout 1\nstdout 2\nstdout 3"
	wantStderr = "stderr 1\nstderr 2\nstderr 3"
	shCmd      = "echo stdout 1; echo stderr 1 >/dev/stderr; echo stdout 2; echo stderr 2 >/dev/stderr; echo -n stdout 3; echo -n stderr 3 >/dev/stderr"
)

var (
	errClosed = errors.New("already closed")
	errWrite  = errors.New("write error")
)

type appendWriter struct {
	s      []string
	closed bool
	err    error
}

func (a *appendWriter) Write(b []byte) (int, error) {
	a.s = append(a.s, string(b))
	return len(b), a.err
}

func (a *appendWriter) Close() error {
	if a.closed {
		return errClosed
	}
	a.closed = true
	return nil
}

func goodOpts(d time.Duration) Options {
	return Options{
		Command:   []string{"sh", "-c", shCmd},
		Timeout:   d,
		SplitFunc: split.ByRE(regexp.MustCompile("\n")),
	}
}

func TestNew(t *testing.T) {
	testdata := []struct {
		name string
		opts Options
		err  error
	}{
		{
			name: "good",
			opts: goodOpts(time.Millisecond),
		},
		{
			name: "no command",
			opts: Options{},
			err:  ErrNoCommand,
		},
	}
	for _, tt := range testdata {
		if _, err := New(tt.opts); !errors.Is(err, tt.err) {
			t.Errorf("New(%v) error = %v, want %v", tt.name, err, tt.err)
		}
	}
}

func TestStart(t *testing.T) {
	testdata := []struct {
		name string
		f    func(*Expect)
		ok   bool
	}{
		{
			name: "good stdout",
			f:    func(*Expect) {},
			ok:   true,
		},
		{
			name: "good stderr",
			f: func(e *Expect) {
				e.opts.MatchStderr = true
			},
			ok: true,
		},
		{
			name: "bad stdin",
			f: func(e *Expect) {
				e.cmd.Stdin = strings.NewReader("foo")
			},
		},
		{
			name: "bad stdout",
			f: func(e *Expect) {
				e.cmd.Stdout = new(strings.Builder)
			},
		},
		{
			name: "bad stderr",
			f: func(e *Expect) {
				e.opts.MatchStderr = true
				e.cmd.Stderr = new(strings.Builder)
			},
		},
	}
	for _, tt := range testdata {
		e, err := New(goodOpts(time.Millisecond))
		if err != nil {
			t.Errorf("New(%v) error = %v", tt.name, err)
			continue
		}
		tt.f(e)
		if err := e.Start(); err != nil {
			if tt.ok {
				t.Errorf("Start(%v) error = %v", tt.name, err)
			}
			continue
		}
		if !tt.ok {
			t.Errorf("Start(%v) error = nil", tt.name)
		}
	}
}

func TestWait(t *testing.T) {
	var errNegative = errors.New("some error")
	testdata := []struct {
		name string
		d    time.Duration
		f    func(e *Expect)
		err  error
	}{
		{
			name: "closed",
			f:    func(e *Expect) { e.closed = true },
			err:  ErrClosed,
		},
		{
			name: "notimeout-found",
			f:    func(e *Expect) { e.found <- struct{}{} },
		},
		{
			name: "notimeout-error",
			f:    func(e *Expect) { e.errc <- errNegative },
			err:  errNegative,
		},
		{
			name: "found",
			d:    time.Second,
			f:    func(e *Expect) { e.found <- struct{}{} },
		},
		{
			name: "error",
			d:    time.Second,
			f:    func(e *Expect) { e.errc <- errNegative },
			err:  errNegative,
		},
		{
			name: "timeout",
			d:    10 * time.Millisecond,
			f:    func(e *Expect) {},
			err:  ErrTimeout,
		},
	}
	for _, tt := range testdata {
		e, err := New(goodOpts(tt.d))
		if err != nil {
			t.Errorf("New(%v) error = %v", tt.name, err)
			continue
		}
		go tt.f(e)
		time.Sleep(time.Millisecond)
		if err := e.Wait(); !errors.Is(err, tt.err) {
			t.Errorf("Wait(%v) error = %v, want %v", tt.name, err, tt.err)
		}
	}
}

func TestReader(t *testing.T) {
	input := "foo\nbar baz\nquux"
	want := []string{
		"foo\n",
		"bar baz\n",
		"quux",
	}
	e, err := New(goodOpts(time.Millisecond))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	e.s = bufio.NewScanner(strings.NewReader(input))
	e.s.Split(e.opts.SplitFunc)
	e.tee = new(appendWriter)
	go e.reader()
	for i := 0; i < len(want); i++ {
		select {
		case <-e.found:
			got := e.tee.(*appendWriter).s[i]
			if got != want[i] {
				t.Errorf("Write(%v) = %q, want %q", i, got, want[i])
			}
		case err := <-e.errc:
			if err != nil {
				t.Errorf("Write(%v) error = %v", i, err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout")
		}
	}
	select {
	case <-e.found:
		t.Error("found = true")
	case err := <-e.errc:
		if err != nil {
			t.Errorf("Write() error = %v", err)
		}
	}
}

func TestReader_error(t *testing.T) {
	want := "foo\nbar baz\nquux"
	e, err := New(goodOpts(time.Millisecond))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	e.s = bufio.NewScanner(strings.NewReader(want))
	e.s.Split(e.opts.SplitFunc)
	e.tee = new(appendWriter)
	e.tee.(*appendWriter).err = errWrite
	go e.reader()
	select {
	case <-e.found:
		t.Error("found = true")
	case err := <-e.errc:
		if !errors.Is(err, errWrite) {
			t.Errorf("Write() error = %v, want %v", err, errWrite)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout")
	}
}

func TestWrite(t *testing.T) {
	e, err := New(goodOpts(10 * time.Millisecond))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	e.w = new(appendWriter)
	e.w.(*appendWriter).err = errWrite
	if _, err := e.Write([]byte("foo")); !errors.Is(err, errWrite) {
		t.Errorf("Write() error = %v, want %v", err, errWrite)
	}
}

func TestExpect(t *testing.T) {
	wantStdout := "stdout 1\nstdout 2\nstdout 3"
	wantStderr := "stderr 1\nstderr 2\nstderr 3"
	e, err := New(goodOpts(10 * time.Millisecond))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	e.stdout = new(strings.Builder)
	e.stderr = new(strings.Builder)
	if err := e.Start(); err != nil {
		t.Errorf("Start() error = %v", err)
	}
	for {
		if err := e.Wait(); err != nil {
			t.Errorf("Wait() error = %v", err)
			continue
		}
		break
	}
	if err := e.DoneRead(); err != nil {
		t.Errorf("DoneRead() error = %v", err)
	}
	if err := e.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
	if got := e.stdout.(*strings.Builder).String(); got != wantStdout {
		t.Errorf("stdout = %q, want %q", got, wantStdout)
	}
	if got := e.stderr.(*strings.Builder).String(); got != wantStderr {
		t.Errorf("stderr = %q, want %q", got, wantStderr)
	}
}
