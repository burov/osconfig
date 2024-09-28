package scanner

import (
	scalibr "github.com/google/osv-scalibr"
	el "github.com/google/osv-scalibr/extractor/filesystem/list"
	"github.com/google/osv-scalibr/extractor/filesystem"
	"github.com/google/osv-scalibr/extractor"
	"github.com/google/osv-scalibr/filesystem/os/dpkg"
)

type extractor extractor.Extractor

const (
	DPKG = extractor(dpkg.New(dpkg.DefaultConfig()))
)

type Scanner struct {
	extractors []extractor
}

func NewScanner(extractors []extractor) *Scanner {
	return &Scanner{
		extractors: extractors,
	}
}
