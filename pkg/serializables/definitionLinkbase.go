package serializables

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func getDefinitionLinkbaseFromOSfiles(files []os.FileInfo) (os.FileInfo, error) {
	var xmls []os.FileInfo
	for _, file := range files {
		s := file.Name()
		if file.IsDir() || filepath.Ext(s) == "" {
			continue
		}
		if s[len(s)-len(defExt):] == defExt {
			xmls = append(xmls, file)
		}
	}
	if len(xmls) <= 0 {
		return nil, fmt.Errorf("no definition linkbase found")
	}
	return xmls[0], nil
}

func getDefinitionLinkbaseFromUnzipfiles(unzipFiles []*zip.File) (*zip.File, error) {
	var xmls []*zip.File
	for _, unzipFile := range unzipFiles {
		s := unzipFile.Name
		if s[len(s)-len(defExt):] == defExt {
			xmls = append(xmls, unzipFile)
		}
	}
	if len(xmls) <= 0 {
		return nil, fmt.Errorf("no definition linkbase found")
	}
	return xmls[0], nil
}

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
