package dummy

import (
	"errors"
	"strings"
	"testing"
)

type writeCloser struct {
	closed bool
	strings.Builder
}

func (wc *writeCloser) Close() error {
	if wc.closed {
		return errors.New("already closed")
	}
	wc.closed = true
	return nil
}

func TestWrite(t *testing.T) {
	data := "foo bar baz"
	wc := new(writeCloser)
	s := New(wc)
	if err := s.Start(); err != nil {
		t.Errorf("Start() error = %v", err)
	}
	if _, err := s.Write([]byte(data)); err != nil {
		t.Errorf("Write(%v) error = %v", data, err)
	}
	if got := wc.String(); got != data {
		t.Errorf("Write(%v) = %v", data, got)
	}
	if err := s.DoneRead(); err != nil {
		t.Errorf("DoneRead() error = %v", err)
	}
	if err := s.Wait(); err != nil {
		t.Errorf("Wait() error = %v", err)
	}
	if err := s.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}
	// Second Stop() should fail.
	if err := s.Stop(); err == nil {
		t.Error("Close() error = nil")
	}
}
