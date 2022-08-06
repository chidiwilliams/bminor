package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	paths, err := filepath.Glob(filepath.Join("testdata", "*.test"))
	if err != nil {
		t.Fatal(err)
	}

	for _, path := range paths {
		_, filename := filepath.Split(path)
		testName := filename[:len(filename)-len(filepath.Ext(path))]

		t.Run(testName, func(t *testing.T) {
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal("error reading test source file:", err)
			}

			goldenFile := filepath.Join("testdata", testName+".expected")
			want, err := os.ReadFile(goldenFile)
			if err != nil {
				t.Fatal("error reading golden file", err)
			}

			wantStdOut, wantStdErr := splitWantOutput(string(want))

			stdErr := bytes.Buffer{}
			stdOut := bytes.Buffer{}
			RunAndReportError(string(source), &stdErr, &stdOut)

			gotStdErr := stdErr.String()
			if wantStdErr != gotStdErr {
				t.Errorf("expected stderr to be %s, got %s", strconv.Quote(wantStdErr), strconv.Quote(gotStdErr))
			}

			gotStdOut := stdOut.String()
			if wantStdOut != gotStdOut {
				t.Errorf("expected stdout to be %s, got %s", strconv.Quote(wantStdOut), strconv.Quote(gotStdOut))
			}
		})
	}
}

func splitWantOutput(output string) (stdOut string, stdErr string) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if len(line) >= 5 && line[:5] == "err: " {
			stdErr += line[5:] + "\n"
		} else {
			stdOut += line + "\n"
		}
	}

	return
}
