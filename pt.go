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
	"log"
	"os"
	"os/exec"
	"regexp"

	"github.com/hazaelsan/pipe-throttler/runner"
	"github.com/hazaelsan/pipe-throttler/split"
	"github.com/hazaelsan/pipe-throttler/throttler/dummy"
)

var (
	interval   = flag.Duration("interval", 0, "how long to wait after the throttler is ready before outputting the next data chunk")
	size       = flag.Uint("size", 0, "how many bytes to read from stdin, overrides --split if > 0")
	splitInput = flag.String("split", "\n", "regular expression on which to split stdin")
)

func main() {
	flag.Parse()
	var splitFunc bufio.SplitFunc
	if *size > 0 {
		splitFunc = split.BySize(int(*size))
	} else {
		inRE, err := regexp.Compile(*splitInput)
		if err != nil {
			log.Fatal(err)
		}
		splitFunc = split.ByRE(inRE)
	}
	opts := runner.Options{
		Reader:       os.Stdin,
		Throttler:    dummy.New(os.Stdout),
		SplitFunc:    splitFunc,
		WaitDuration: *interval,
	}
	r := runner.New(opts)
	if err := r.Run(); err != nil {
		var e *exec.ExitError
		if errors.As(err, &e) {
			os.Exit(e.ExitCode())
		}
		os.Exit(1)
	}
}
