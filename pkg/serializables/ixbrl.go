package serializables

import (
	"fmt"
	"path/filepath"
	"strings"
)

func getIxbrlFileFromFilingItems(filingItems []filingItem, ticker string) (*filingItem, error) {
	for _, f := range filingItems {
		s := f.Name
		ext := filepath.Ext(s)
		a := (ext == ixbrlExt && strings.Index(s, ticker) == 0)
		b := len(s) >= 8
		if a && b {
			return &f, nil
		}
	}
	return nil, fmt.Errorf("cannot identify a single ixbrl file")
}
