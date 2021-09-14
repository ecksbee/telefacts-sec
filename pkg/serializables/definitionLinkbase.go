package serializables

import (
	"fmt"
	"path/filepath"
	"strings"
)

func getDefinitionLinkbaseFromFilingItems(filingItems []filingItem, ticker string) (*filingItem, error) {
	for _, f := range filingItems {
		s := f.Name
		ext := filepath.Ext(s)
		a := (ext == xmlExt && strings.Index(s, ticker) == 0)
		b := len(s) >= 8
		if b {
			longExt := s[len(s)-8:]
			b = longExt == defExt
		}
		if a && b {
			return &f, nil
		}
	}
	return nil, fmt.Errorf("cannot identify a single definition linkbase")
}
