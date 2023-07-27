package unless

import (
	"testing"
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
			cutOut, err := cut(tc.input, tc.index)
			if err != nil {
				t.Logf("unexpected cut error: %v", err)
			}
			if cutOut != tc.expected {
				t.Logf("actual output doesn't match expected output: %s != %s", cutOut, tc.expected)
				t.Fail()
			}
		})
	}
}
