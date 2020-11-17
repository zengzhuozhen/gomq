package utils

import (
	"testing"
)

var tests = []struct {
	in  int
	out int
}{
	{64, 1},
	{321, 2},
}



func TestRemainingLengthAlgorithm(t *testing.T) {
	for i, tt := range tests {
		s := EncodeRemainingLengthAlg(tt.in)
		if BytesToInt(s) != tt.out {
			t.Errorf("%d. %q => %q, wanted: %q", i, tt.in, s, tt.out)
		}
	}
}


