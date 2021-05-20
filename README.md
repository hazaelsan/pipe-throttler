# Pipe Throttler

[![Build Status](https://travis-ci.com/hazaelsan/pipe-throttler.svg?branch=main)](https://travis-ci.com/hazaelsan/pipe-throttler)
[![GoDoc](https://godoc.org/github.com/hazaelsan/pipe-throttler?status.svg)](https://godoc.org/github.com/hazaelsan/pipe-throttler)

`pt` is a UNIX pipeline throttler.  It takes data from `stdin` and throttles writing it out to `stdout`.

## Why?

This tool probably shouldn't exist.

There are some broken programs that don't deal well with lots of data fed through `stdin`.  Ideally those programs should be fixed, but unfortunately that's not always feasible.

## Splitting input

`pt` has two splitting input modes: `regexp` and `size`.

### `regexp` mode

`pt` takes a `--split` flag to split output by a regular expression, e.g., `--split="\n"` will split output on newlines.  This is the default behavior.

### `size` mode

`pt` can also split output in fixed-size chunks (in bytes), e.g., `--size=1024` will split output every 1024 bytes.

Note: Due to multi-byte character encodings, it's possible to split input in the middle of a character.  In practice this shouldn't be an issue since data is written out unmodified.

## Output modes

`pt` has two output modes: `throttle` and `expect`.

### `throttle` mode

In this mode, `pt` simply rate-limits output based on the `--interval` flag, e.g.,

```
$ echo -e "foo\nbar\nbaz" | pt --interval=1s
```

Will output `foo`, `bar`, and `baz` at one second intervals.

This mode is implied when there are no non-flag arguments passed to `pt`.

### `expect` mode

In this mode `pt` will spawn the given command and wait for the wrapped command to output matching either `--expect_size` or `--expect_split`, e.g.,:

```shell
$ cat command.sh
#!/bin/bash
while true; do
  echo -n "> "
  if read line; then
    echo "${line}"
    continue
  fi
  [[ -n "${line}" ]] && echo -n "${line}"
  exit 0
done

$ echo -en "foo\nbar\nbaz" | pt ./command.sh
```

Note: If the wrapped command takes any flags then you MUST precede it with `--` to prevent those flags being interpreted by `pt`:

```shell
$ some_producer | pt -- wrapped_command --flag1 --flag2 ...
```
