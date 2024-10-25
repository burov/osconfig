package scalibr

import (
	"context"
	"fmt"
	"runtime"

	"github.com/GoogleCloudPlatform/osconfig/extractors"
	scalibr "github.com/google/osv-scalibr"
	scalibr_extractor "github.com/google/osv-scalibr/extractor"
	"github.com/google/osv-scalibr/extractor/filesystem"
	"github.com/google/osv-scalibr/extractor/filesystem/os/dpkg"
	scalibrfs "github.com/google/osv-scalibr/fs"
	scalibr_plugin "github.com/google/osv-scalibr/plugin"
)

var (
	dpkgExtractor = dpkg.New(dpkg.DefaultConfig())
)

var _ extractors.Extractor = &ScalibrExtractor{}

type ScalibrExtractor struct{}

func (s *ScalibrExtractor) ExtractInventory(ctx context.Context, extractionSources ...extractors.ExtractionSource) ([]extractors.Inventory, error) {

	cfg, err := scalibrScanConfig(extractionSources)
	if err != nil {
		return nil, err
	}

	results := scalibr.New().Scan(ctx, cfg)

	return deconstructScanResult(results)
}

func deconstructScanResult(results *scalibr.ScanResult) ([]extractors.Inventory, error) {
	inventories := make([]extractors.Inventory, 0, len(results.Inventories))

	scanStatus := results.Status
	if scanStatus.Status == scalibr_plugin.ScanStatusFailed {
		return nil, fmt.Errorf("scan failed, failure reason: %s", scanStatus.FailureReason)
	}

	for _, inv := range results.Inventories {
		item, err := inventoryFrom(inv)
		if err != nil {
			//TODO: log unexpected error here
			continue
		}
		inventories = append(inventories, item)
	}

	if scanStatus.Status != scalibr_plugin.ScanStatusSucceeded {
		return inventories, fmt.Errorf("scan partially failed, failure reason: %s", scanStatus.FailureReason)
	}

	return inventories, nil
}

func inventoryFrom(inventory *scalibr_extractor.Inventory) (extractors.Inventory, error) {
	item := extractors.Inventory{
		Name:    inventory.Name,
		Version: inventory.Version,
	}

	switch metadata := inventory.Metadata.(type) {
	case *dpkg.Metadata:
		source, purl, err := extractAdditionalFieldsDpkg(inventory, metadata)
		if err != nil {
			return extractors.Inventory{}, fmt.Errorf("unable to extract additional fields, err: %v", err)
		}

		item.Source = source
		item.Purl = purl
	default:
		//TODO: consider to return just name and version if possible.
		return extractors.Inventory{}, fmt.Errorf("unsupported inventory item")
	}

	return item, nil
}

func extractAdditionalFieldsDpkg(inventory *scalibr_extractor.Inventory, metadata *dpkg.Metadata) (extractors.Source, string, error) {
	source := extractors.Source{
		Name:    metadata.SourceName,
		Version: metadata.SourceVersion,
	}

	purl, err := dpkgExtractor.ToPURL(inventory)
	if err != nil {
		return extractors.Source{}, "", fmt.Errorf("unable to extract purl, %v", err)
	}

	return source, purl.String(), nil
}

func scalibrScanConfig(sources ...extractors.ExtractionSource) (*scalibr.ScanConfig, error) {
	return &scalibr.ScanConfig{
		ScanRoots:            scalibrfs.RealFSScanRoots(fsRootDir()),
		FilesystemExtractors: extractorsFrom(sources),
	}, nil
}

func extractorsFrom(sources ...extractors.ExtractionSource) []filesystem.Extractor {
	extractors := make([]filesystem.Extractor, 0, len(sources))

	for _, s := range sources {
		extractors = append(extractors, extractorFrom(s))
	}

	return extractors
}

func extractorFrom(es extractors.ExtractionSource) filesystem.Extractor {
	return dpkg.New(dpkg.DefaultConfig())
}

func fsRootDir() string {
	if runtime.GOOS == "windows" {
		return "C:"
	}
	return "/"
}
