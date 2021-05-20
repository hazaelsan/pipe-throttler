package main

import (
	"errors"
	"flag"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/hazaelsan/pipe-throttler/throttler/dummy"
)

func TestExitCode(t *testing.T) {
	testdata := map[error]int{
		exec.Command("sh", "-c", "exit 123").Run(): 123,
		errors.New("some error"):                   1,
		nil:                                        0,
	}
	for err, want := range testdata {
		if got := exitCode(err); got != want {
			t.Errorf("exitCode(%v) = %v, want %v", err, got, want)
		}
	}
}

func TestNewThrottler(t *testing.T) {
	testdata := []struct {
		name  string
		args  []string
		size  int
		split string
		dummy bool
		ok    bool
	}{
		{
			name:  "dummy",
			dummy: true,
			ok:    true,
		},
		{
			name: "size",
			args: []string{"foo"},
			size: 3,
			ok:   true,
		},
		{
			name:  "split",
			args:  []string{"foo"},
			split: "...",
			ok:    true,
		},
		{
			name: "empty split",
			args: []string{"foo"},
		},
		{
			name:  "bad split",
			args:  []string{"foo"},
			split: "?bad",
		},
	}
	for _, tt := range testdata {
		pt, err := newThrottler(tt.args, tt.size, tt.split, false, 0)
		if err != nil {
			if tt.ok {
				t.Errorf("newThrottler(%v) error = %v", tt.name, err)
			}
			continue
		}
		if !tt.ok {
			t.Errorf("newThrottler(%v) error = nil", tt.name)
		}
		if _, ok := pt.(*dummy.Dummy); ok != tt.dummy {
			t.Errorf("newThrottler(%v) = %T", tt.name, pt)
		}
	}
}

func TestNewRunner(t *testing.T) {
	osArgs := os.Args
	defer func() {
		os.Args = osArgs
		flag.Parse()
	}()
	testdata := []struct {
		name     string
		size     int
		split    string
		eSplit   string
		interval time.Duration
		args     []string
		ok       bool
	}{
		{
			name:  "good",
			split: "\n",
			ok:    true,
		},
		{
			name: "empty split",
		},
		{
			name:  "bad split",
			split: "?bad",
		},
		{
			name:   "expect",
			split:  "\n",
			eSplit: "?bad",
			args:   []string{"invalid"},
		},
	}
	for _, tt := range testdata {
		os.Args = append(osArgs, tt.args...)
		flag.Parse()
		flag.Set("size", strconv.Itoa(tt.size))
		flag.Set("split", tt.split)
		flag.Set("expect_split", tt.eSplit)
		if _, err := newRunner(); err != nil {
			if tt.ok {
				t.Errorf("newRunner(%v) error = %v", tt.name, err)
			}
			continue
		}
		if !tt.ok {
			t.Errorf("newRunner(%v) error = nil", tt.name)
		}
	}
}

func TestRun(t *testing.T) {
	osArgs := os.Args
	defer func() {
		os.Args = osArgs
		flag.Parse()
	}()
	testdata := []struct {
		name  string
		split string
		args  []string
	}{
		{
			name:  "dummy",
			split: "?bad",
		},
		{
			name:  "expect",
			split: "?bad",
			args:  []string{"invalid"},
		},
	}
	for _, tt := range testdata {
		flag.Set("split", tt.split)
		flag.Set("expect_split", tt.split)
		if err := run(); err == nil {
			t.Errorf("run(%v) error = nil", tt.name)
		}
	}
}
