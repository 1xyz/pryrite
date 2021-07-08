package kmd

import (
	"testing"
)

var flagtests = []struct {
	in          []string
	expectedErr error
	expectedMsg string
	expecteCmd  string
}{
	{[]string{"-m", "foo", "hello", "world"}, nil, "foo", "hello world"},
	{[]string{"-m", "\"foo", "says", "bar\"", "hello", "world"}, nil, "foo says bar", "hello world"},
	{[]string{"-m", "\"foo", "says", "bar", "hello", "world"}, ErrEOL, "", ""},
	{[]string{"-m", "\"foo", "says", "bar", "hello", "world\""}, ErrCmdMissing, "", ""},
	{[]string{"says", "bar", "hello", "world"}, nil, "", "says bar hello world"},
}

func TestParseSaveArgs(t *testing.T) {
	for _, tt := range flagtests {
		s, err := ParseSaveArgs(tt.in)
		if err != tt.expectedErr {
			t.Errorf("expected error to match")
		}
		if s == nil {
			return
		}
		if tt.expectedMsg != s.Message {
			t.Errorf("mismatch %s != %s", tt.expectedMsg, s.Message)
		}
		if tt.expecteCmd != s.Command {
			t.Errorf("mismatch %s != %s", tt.expecteCmd, s.Command)
		}
	}
}
