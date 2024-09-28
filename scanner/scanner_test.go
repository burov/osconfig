package scanner

import (
	"testing"
)

func TestDummy(t *testing.T) {
	scanner := NewScanner([]extractor{DPKG})
}
