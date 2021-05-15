# Pipe Throttler

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
