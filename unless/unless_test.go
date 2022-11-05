package precheck_test

import (
	"testing"

	precheck "github.com/femnad/fup/unless"
)

func TestCut(t *testing.T) {
	for _, tc := range []struct {
		input    string
		index    int
		expected string
	}{
		{
			input:    "xbaz",
			index:    1,
			expected: "baz",
		},
		{
			input:    "fooz",
			index:    -1,
			expected: "foo",
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			cut, err := precheck.Cut(tc.input, tc.index)
			if err != nil {
				t.Logf("unexpected cut error: %v", err)
			}
			if cut != tc.expected {
				t.Logf("actual output doesn't match expected output: %s != %s", cut, tc.expected)
				t.Fail()
			}
		})
	}
}
