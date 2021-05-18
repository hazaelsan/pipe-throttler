package runner

import (
	"errors"
	"io"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hazaelsan/pipe-throttler/split"
	"github.com/hazaelsan/pipe-throttler/throttler/dummy"
	"github.com/kylelemons/godebug/pretty"
)

var (
	errRead  = errors.New("read error")
	errStart = errors.New("start error")
	errWrite = errors.New("write error")
)

type appendWriter struct {
	s   []string
	err error
}

func (a *appendWriter) Write(b []byte) (int, error) {
	a.s = append(a.s, string(b))
	return len(b), a.err
}

func (a *appendWriter) Close() error {
	return nil
}

type badReader struct{}

func (badReader) Read([]byte) (int, error) {
	return 0, errRead
}

type badThrottler struct {
	dummy.Dummy
}

func (*badThrottler) Start() error {
	return errStart
}

func newRunner(r io.Reader, w io.WriteCloser) *Runner {
	opts := Options{
		Reader:       r,
		Throttler:    dummy.New(w),
		SplitFunc:    split.ByRE(regexp.MustCompile("\n")),
		WaitDuration: time.Millisecond,
	}
	return New(opts)
}

func TestRun(t *testing.T) {
	input := "foo\nbar baz\nquux"
	testdata := []struct {
		name string
		f    func(*appendWriter) *Runner
		want []string
		err  error
	}{
		{
			name: "good",
			f: func(w *appendWriter) *Runner {
				return newRunner(strings.NewReader(input), w)
			},
			want: []string{"foo\n", "bar baz\n", "quux"},
		},
		{
			name: "start error",
			f: func(w *appendWriter) *Runner {
				return &Runner{t: new(badThrottler)}
			},
			err: errStart,
		},
		{
			name: "bad reader",
			f: func(w *appendWriter) *Runner {
				return newRunner(new(badReader), w)
			},
			err: errRead,
		},
		{
			name: "bad writer",
			f: func(w *appendWriter) *Runner {
				w.err = errWrite
				return newRunner(strings.NewReader(input), w)
			},
			want: []string{"foo\n"},
			err:  errWrite,
		},
	}
	for _, tt := range testdata {
		w := new(appendWriter)
		r := tt.f(w)
		if err := r.Run(); !errors.Is(err, tt.err) {
			t.Errorf("Run(%v) error = %v, want %v", tt.name, err, tt.err)
		}
		if diff := pretty.Compare(w.s, tt.want); diff != "" {
			t.Errorf("reader(%v) -got +want:\n%v", tt.name, diff)
		}
	}
}
