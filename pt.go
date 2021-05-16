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
	"log"
	"os"
	"os/exec"
	"regexp"

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

func newSplitFunc(size uint, pat string) (bufio.SplitFunc, error) {
	if size > 0 {
		return split.BySize(int(size)), nil
	}
	re, err := regexp.Compile(pat)
	if err != nil {
		return nil, err
	}
	return split.ByRE(re), nil
}

func newThrottler() (throttler.Throttler, error) {
	if len(flag.Args()) > 0 {
		return newExpect()
	}
	return dummy.New(os.Stdout), nil
}

func newExpect() (*expect.Expect, error) {
	f, err := newSplitFunc(*expectSize, *expectSplit)
	if err != nil {
		return nil, err
	}
	opts := expect.Options{
		Command:     flag.Args(),
		MatchStderr: *expectStderr,
		SplitFunc:   f,
		Timeout:     *expectTimeout,
	}
	return expect.New(opts)
}

func main() {
	flag.Parse()
	f, err := newSplitFunc(*size, *splitInput)
	if err != nil {
		log.Fatal(err)
	}
	t, err := newThrottler()
	if err != nil {
		log.Fatal(err)
	}
	opts := runner.Options{
		Reader:       os.Stdin,
		Throttler:    t,
		SplitFunc:    f,
		WaitDuration: *interval,
	}
	r := runner.New(opts)
	if err := r.Run(); err != nil {
		var e *exec.ExitError
		if errors.As(err, &e) {
			os.Exit(e.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
