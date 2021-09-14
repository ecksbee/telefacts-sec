package serializables

import (
	"fmt"
	"path/filepath"
)

func getSchemaFromFilingItems(filingItems []filingItem) (*filingItem, error) {
	var candidates []filingItem
	for _, f := range filingItems {
		s := f.Name
		ext := filepath.Ext(s)
		if ext == xsdExt {
			candidates = append(candidates, f)
		}
	}
	if len(candidates) > 1 {
		return nil, fmt.Errorf("cannot identify a single schema")
	}
	if len(candidates) <= 0 {
		return nil, fmt.Errorf("no schema found")
	}
	return &(candidates[0]), nil
}
