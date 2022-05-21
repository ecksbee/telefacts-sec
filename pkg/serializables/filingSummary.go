package serializables

import (
	"encoding/xml"
	"path/filepath"
	"strings"
)

type FilingSummary struct {
	XMLName    xml.Name `xml:"FilingSummary"`
	InputFiles []struct {
		XMLName xml.Name
		File    []struct {
			XMLName  xml.Name
			XMLAttrs []xml.Attr `xml:",any,attr"`
			CharData string     `xml:",chardata"`
		} `xml:"File"`
	} `xml:"InputFiles"`
}

func (fs *FilingSummary) GetIxbrl() string {
	for _, file := range fs.InputFiles[0].File {
		for _, fattr := range file.XMLAttrs {
			if fattr.Name.Local == "original" {
				return file.CharData
			}
		}
	}
	return ""
}

func (fs *FilingSummary) GetInstance() string {
	ticker := fs.GetTicker()
	if len(ticker) <= 0 {
		return ""
	}
	for _, f := range fs.InputFiles[0].File {
		s := f.CharData
		ext := filepath.Ext(s)
		a := (ext == xmlExt && strings.Index(s, ticker) == 0)
		b := len(s) >= 8
		if b {
			longExt := s[len(s)-8:]
			b = longExt != preExt && longExt != defExt && longExt != calExt && longExt != labExt
		}
		if a && b {
			return s
		}
	}
	return ""
}

func (fs *FilingSummary) GetTicker() string {
	str := fs.GetSchema()
	x := strings.Index(str, "-")
	ticker := str[:x]
	if len(ticker) <= 0 {
		return ""
	}
	return ticker
}

func (fs *FilingSummary) GetSchema() string {
	for _, f := range fs.InputFiles[0].File {
		s := f.CharData
		ext := filepath.Ext(s)
		if ext == xsdExt {
			return s
		}
	}
	return ""
}

func (fs *FilingSummary) GetCalculationLinkbase() string {
	ticker := fs.GetTicker()
	if len(ticker) <= 0 {
		return ""
	}
	for _, f := range fs.InputFiles[0].File {
		s := f.CharData
		ext := filepath.Ext(s)
		a := (ext == xmlExt && strings.Index(s, ticker) == 0)
		b := len(s) >= 8
		if b {
			longExt := s[len(s)-8:]
			b = longExt == calExt
		}
		if a && b {
			return s
		}
	}
	return ""
}

func (fs *FilingSummary) GetDefinitionLinkbase() string {
	ticker := fs.GetTicker()
	if len(ticker) <= 0 {
		return ""
	}
	for _, f := range fs.InputFiles[0].File {
		s := f.CharData
		ext := filepath.Ext(s)
		a := (ext == xmlExt && strings.Index(s, ticker) == 0)
		b := len(s) >= 8
		if b {
			longExt := s[len(s)-8:]
			b = longExt == defExt
		}
		if a && b {
			return s
		}
	}
	return ""
}

func (fs *FilingSummary) GetLabelLinkbase() string {
	ticker := fs.GetTicker()
	if len(ticker) <= 0 {
		return ""
	}
	for _, f := range fs.InputFiles[0].File {
		s := f.CharData
		ext := filepath.Ext(s)
		a := (ext == xmlExt && strings.Index(s, ticker) == 0)
		b := len(s) >= 8
		if b {
			longExt := s[len(s)-8:]
			b = longExt == labExt
		}
		if a && b {
			return s
		}
	}
	return ""
}

func (fs *FilingSummary) GetPresentationLinkbase() string {
	ticker := fs.GetTicker()
	if len(ticker) <= 0 {
		return ""
	}
	for _, f := range fs.InputFiles[0].File {
		s := f.CharData
		ext := filepath.Ext(s)
		a := (ext == xmlExt && strings.Index(s, ticker) == 0)
		b := len(s) >= 8
		if b {
			longExt := s[len(s)-8:]
			b = longExt == preExt
		}
		if a && b {
			return s
		}
	}
	return ""
}
