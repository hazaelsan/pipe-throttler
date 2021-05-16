// Package split defines splitters for use with bufio.Scanner().
package split

import (
	"bufio"
	"regexp"
)

// ByRE is a closure for a SplitFunc that splits on a regular expression.
// Returns all remaining data up to (and including) the matched expression.
func ByRE(re *regexp.Regexp) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF {
			if len(data) == 0 {
				return 0, nil, nil
			}
			return len(data), data, nil
		}
		if loc := re.FindIndex(data); loc != nil {
			return loc[1], data[0:loc[1]], nil
		}
		return 0, nil, nil
	}
}

// BySize is a closure for a SplitFunc that splits on a fixed byte size.
func BySize(size int) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF {
			return 0, nil, nil
		}
		if size > len(data) {
			return len(data), data, nil
		}
		return size, data[0:size], nil
	}
}
