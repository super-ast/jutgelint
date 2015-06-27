# jutgelint

A lint for Jutge. Integrates the AST translators from C++ and Go, the current
supported languages, with our generic AST checker. It also supports using the
resulting warnings to place inline comments in the original code.

## Setup

	go get github.com/mvdan/jutgelint/cmd/jutge-lint

* The C++ AST translator `superast-cpp` must be installed and in your `$PATH`
* The Go AST translator is already bundled in the static Go binaries
* The generic AST checker `checker` must be installed and in your `$PATH`

### Usage

```
Usage: jutgelint [input] [output]

The input and output files default to standard input and standard output
if none are specified.

Options:
  -lang=auto: Language to use (auto, c++, go)

Examples:
    jutgelint input.go output.go
    jutgelint -lang=cpp <input.cc >output.cc
```
