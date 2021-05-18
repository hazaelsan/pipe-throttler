package throttler

import (
	"errors"
	"testing"
	"time"

	"github.com/hazaelsan/pipe-throttler/throttler/dummy"
)

var (
	errWait  = errors.New("wait error")
	errWrite = errors.New("write error")
)

type writeCloser struct {
	s   string
	err error
}

func (wc *writeCloser) Write(b []byte) (int, error) {
	wc.s += string(b)
	return len(b), wc.err
}

func (wc *writeCloser) Close() error {
	return nil
}

type throttler struct {
	*dummy.Dummy
	err error
}

func (t throttler) Wait() error {
	return t.err
}

func TestWrite(t *testing.T) {
	data := []byte("foo\nbar baz\nquux")
	testdata := []struct {
		name string
		f    func(*throttler)
		err  error
	}{
		{
			name: "good",
			f:    func(*throttler) {},
		},
		{
			name: "bad wait",
			f: func(t *throttler) {
				t.err = errWait
			},
			err: errWait,
		},
		{
			name: "bad write",
			f: func(t *throttler) {
				t.err = errWrite
			},
			err: errWrite,
		},
	}
	for _, tt := range testdata {
		wc := new(writeCloser)
		d := throttler{Dummy: dummy.New(wc)}
		tt.f(&d)
		if err := Write(d, data, time.Millisecond); !errors.Is(err, tt.err) {
			t.Errorf("Write(%v) error = %v, want %v", tt.name, err, tt.err)
		}
	}
}
