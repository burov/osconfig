package packages

import (
	"context"
	"encoding/json"

	"github.com/GoogleCloudPlatform/osconfig/clog"

)

type comparisonResults struct {
	legacyExtractorItemsCount int `json:"extracted_items_count"`
	modernExtractorItemsCount    int `json:"extracted_items_count"`

	legacyExtractorExtra    []*PkgInfo `json:"legacy_extractor_extra"`
	modernExtractorExtra    []*PkgInfo `json:"new_extractor_extra"`
}

func compareExtractedPackages(legacyExtractor, modernExtractor []*PkgInfo) comparisonResults {
	legacyExtractorIndex := indexByKey(legacyExtractor)
	modernExtractorIndex := indexByKey(modernExtractor)

	var legacyExtractorExtra, modernExtractorExtra []*PkgInfo
	for key, pkg := range legacyExtractorIndex {
		if _, ok := modernExtractorIndex[key]; !ok {
			legacyExtractorExtra = append(legacyExtractorExtra, pkg)
		}
	}

	for key, pkg := range modernExtractorIndex {
		if _, ok := legacyExtractorIndex[key]; !ok {
			modernExtractorExtra = append(modernExtractorExtra, pkg)
		}
	}

	return comparisonResults{
		legacyExtractorItemsCount: len(legacyExtractor),
		modernExtractorItemsCount: len(modernExtractor),

		legacyExtractorExtra: legacyExtractorExtra,
		modernExtractorExtra: modernExtractorExtra,
	}
}

func indexByKey(pkgs []*PkgInfo) map[string]*PkgInfo {
	results := make(map[string]*PkgInfo, len(pkgs))

	for _, pkg := range pkgs {
		results[pkg.key()] = pkg
	}

	return results
}

func printComparisonResult(ctx context.Context, report comparisonResults) {
	raw, err := json.MarshalIndent(report, "", "    ")
	if err != nil {
		clog.Errorf(ctx, "unable comparisonResults, err - %v", err)
		return
	}

	clog.Infof(ctx, "Comparison results after inventory extraction")
	clog.Infof(ctx, string(raw))
}
