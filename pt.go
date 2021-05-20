// pipe-throttler throttles pipeline output.
//
// Usage:
//   $ some_producer | pt --interval=1s | some_consumer --consumer_args...
//   $ some_producer | pt --interval=1s some_consumer -- --consumer_args...
//
// See README.md for additional information between the two modes.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/hazaelsan/pipe-throttler/runner"
	"github.com/hazaelsan/pipe-throttler/split"
	"github.com/hazaelsan/pipe-throttler/throttler"
	"github.com/hazaelsan/pipe-throttler/throttler/dummy"
	"github.com/hazaelsan/pipe-throttler/throttler/expect"
)

var (
	interval   = flag.Duration("interval", 0, "how long to wait after the throttler is ready before outputting the next data chunk")
	size       = flag.Uint("size", 0, "how many bytes to read from stdin, overrides --split if > 0")
	splitInput = flag.String("split", "\n", "regular expression on which to split stdin")

	expectSize    = flag.Uint("expect_size", 0, "how many bytes to read from the wrapped command, overrides --expect_split if > 0")
	expectSplit   = flag.String("expect_split", "\n", "regular expression on which to split the wrapped command's output")
	expectStderr  = flag.Bool("expect_stderr", false, "whether to match the wrapped command's stderr as opposed to stdout")
	expectTimeout = flag.Duration("expect_timeout", 0, "how long to wait for the wrapped command to match --expect_split, waits forever if <= 0")
)

func newSplitFunc(size int, pat string) (bufio.SplitFunc, error) {
	if size > 0 {
		return split.BySize(size), nil
	}
	if pat == "" {
		return nil, errors.New("empty split pattern")
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, err
	}
	return split.ByRE(re), nil
}

func newThrottler(args []string, size int, split string, stderr bool, timeout time.Duration) (throttler.Throttler, error) {
	if len(args) == 0 {
		return dummy.New(os.Stdout), nil
	}
	f, err := newSplitFunc(size, split)
	if err != nil {
		return nil, err
	}
	opts := expect.Options{
		Command:     args,
		MatchStderr: stderr,
		SplitFunc:   f,
		Timeout:     timeout,
	}
	return expect.New(opts)
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var e *exec.ExitError
	if errors.As(err, &e) {
		return e.ExitCode()
	}
	fmt.Fprintln(os.Stderr, err)
	return 1
}

func newRunner() (*runner.Runner, error) {
	f, err := newSplitFunc(int(*size), *splitInput)
	if err != nil {
		return nil, err
	}
	t, err := newThrottler(flag.Args(), int(*expectSize), *expectSplit, *expectStderr, *expectTimeout)
	if err != nil {
		return nil, err
	}
	opts := runner.Options{
		Reader:       os.Stdin,
		Throttler:    t,
		SplitFunc:    f,
		WaitDuration: *interval,
	}
	return runner.New(opts), nil
}

func run() error {
	r, err := newRunner()
	if err != nil {
		return err
	}
	return r.Run()
}

func main() {
	flag.Parse()
	os.Exit(exitCode(run()))
}
