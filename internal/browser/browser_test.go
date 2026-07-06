package browser

import (
	"bytes"
	"errors"
	"testing"
)

type failingOpener struct{}

func (failingOpener) Open(string) error {
	return errors.New("no browser")
}

func TestOpenOrPrintHeadlessPrintsURL(t *testing.T) {
	buffer := bytes.Buffer{}
	err := OpenOrPrint(failingOpener{}, "https://example.test", Environment{GOOS: "linux"}, &buffer)
	if err != nil {
		t.Fatal(err)
	}
	if got := buffer.String(); got != "https://example.test\n" {
		t.Fatalf("output = %q", got)
	}
}
