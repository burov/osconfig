package extractors

import (
	"context"
)

var (
	DpkgSource = DpkgExtractionSource{}
)

type Inventory struct {
	Name    string
	Version string

	RawArch string

	Source Source
	Purl   string
}

type Source struct {
	Name    string
	Version string
}

type Extractor interface {
	ExtractInventory(ctx context.Context, extractionSources ...ExtractionSource) ([]Inventory, error)
}

type ExtractionSource interface{}

type DpkgExtractionSource struct{}
