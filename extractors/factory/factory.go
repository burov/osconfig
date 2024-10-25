package factory

import (
	"github.com/GoogleCloudPlatform/osconfig/extractors"
	"github.com/GoogleCloudPlatform/osconfig/extractors/scalibr"
)

func GetExtractor() extractors.Extractor {
	return &scalibr.ScalibrExtractor{}
}
