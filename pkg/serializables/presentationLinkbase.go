package serializables

import (
	"fmt"
	"path/filepath"
	"strings"
)

func getPresentationLinkbaseFromFilingItems(filingItems []filingItem, ticker string) (*filingItem, error) {
	for _, f := range filingItems {
		s := f.Name
		ext := filepath.Ext(s)
		a := (ext == xmlExt && strings.Index(s, ticker) == 0)
		b := len(s) >= 8
		if b {
			longExt := s[len(s)-8:]
			b = longExt == preExt
		}
		if a && b {
			return &f, nil
		}
	}
	return nil, fmt.Errorf("cannot identify a single presentation linkbase")
}
