package split

import (
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/kylelemons/godebug/pretty"
)

func mkString(n int, sep string, tail bool) string {
	var lines []string
	for i := 1; i <= n; i++ {
		lines = append(lines, fmt.Sprintf("line %v", i))
	}
	if tail {
		return strings.Join(lines, sep) + sep
	}
	return strings.Join(lines, sep)
}

func TestByRE(t *testing.T) {
	testdata := map[string][]string{
		mkString(3, "\n", true): {
			"line 1\n",
			"line 2\n",
			"line 3\n",
		},
		mkString(3, "\n", false): {
			"line 1\n",
			"line 2\n",
			"line 3",
		},
	}
	for tt, want := range testdata {
		s := bufio.NewScanner(strings.NewReader(tt))
		s.Split(ByRE(regexp.MustCompile("\n")))
		var got []string
		for s.Scan() {
			got = append(got, s.Text())
		}
		if err := s.Err(); err != nil {
			t.Errorf("s.Err() = %v", err)
		}
		if diff := pretty.Compare(got, want); diff != "" {
			t.Errorf("ByRE() -got +want:\n%v", diff)
		}
	}
}

func TestBySize(t *testing.T) {
	testdata := map[string][]string{
		mkString(3, "\n", true): {
			"line 1\n",
			"line 2\n",
			"line 3\n",
		},
		mkString(3, "\n", false): {
			"line 1\n",
			"line 2\n",
			"line 3",
		},
	}
	for tt, want := range testdata {
		s := bufio.NewScanner(strings.NewReader(tt))
		s.Split(BySize(7))
		var got []string
		for s.Scan() {
			got = append(got, s.Text())
		}
		if err := s.Err(); err != nil {
			t.Errorf("s.Err() = %v", err)
		}
		if diff := pretty.Compare(got, want); diff != "" {
			t.Errorf("BySize() -got +want:\n%v", diff)
		}
	}
}
