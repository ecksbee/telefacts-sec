package web

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	neturl "net/url"
	"strconv"
	"strings"
	"time"

	"ecksbee.com/telefacts-sec/internal/actions"
	"ecksbee.com/telefacts-sec/pkg/throttle"
	"golang.org/x/net/html/charset"
)

type EdgarRss struct {
	XMLName  xml.Name   `xml:"feed"`
	XMLAttrs []xml.Attr `xml:",any,attr"`
	Updated  []struct {
		XMLName  xml.Name
		XMLAttrs []xml.Attr `xml:",any,attr"`
		CharData string     `xml:",chardata"`
	} `xml:"updated"`
	Entry []struct {
		XMLName  xml.Name
		XMLAttrs []xml.Attr `xml:",any,attr"`
		Title    []struct {
			XMLName  xml.Name
			XMLAttrs []xml.Attr `xml:",any,attr"`
			CharData string     `xml:",chardata"`
		} `xml:"title"`
		Link []struct {
			XMLName  xml.Name
			XMLAttrs []xml.Attr `xml:",any,attr"`
		} `xml:"link"`
		Summary []struct {
			XMLName  xml.Name
			XMLAttrs []xml.Attr `xml:",any,attr"`
			CharData string     `xml:",chardata"`
		} `xml:"summary"`
		Updated []struct {
			XMLName  xml.Name
			XMLAttrs []xml.Attr `xml:",any,attr"`
			CharData string     `xml:",chardata"`
		} `xml:"updated"`
		Category []struct {
			XMLName  xml.Name
			XMLAttrs []xml.Attr `xml:",any,attr"`
			CharData string     `xml:",chardata"`
		} `xml:"category"`
		Id []struct {
			XMLName  xml.Name
			XMLAttrs []xml.Attr `xml:",any,attr"`
			CharData string     `xml:",chardata"`
		} `xml:"id"`
	} `xml:"entry"`
}

type Edgar2025Response struct {
	Hits struct {
		Hits []struct {
			Source struct {
				Adsh         string   `json:"adsh"`
				Ciks         []string `json:"ciks"`
				DisplayNames []string `json:"display_names"`
				FileDate     string   `json:"file_date"`
				FileType     string   `json:"file_type"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type SearchResult struct {
	SearchText string
	Results    []struct {
		Title                  string
		Summary                string
		EdgarUrl               string
		PercentEncodedEdgarUrl string
	}
}

func getFilings(search string, formType string, year string) (*SearchResult, error) {
	yearInt, err := strconv.Atoi(year)
	if err != nil {
		return nil, err
	}
	if yearInt < 2025 {
		return getLegacyFilings(search, formType, year)
	}
	return getModernFilings(search, strings.ToUpper(formType), year)
}

func getModernFilings(search string, formType string, year string) (*SearchResult, error) {
	yearInt, err := strconv.Atoi(year)
	if err != nil {
		return nil, err
	}
	startDate := year + `-01-01`
	endDate := year + `-12-31`
	now := time.Now()
	if now.Year() == yearInt {
		endDate = now.Format("2006-01-02")
	}
	queryText := `q=` + neturl.QueryEscape(search) + `&dateRange=custom&category=custom&startdt=` +
		startDate + `&enddt=` + endDate + `&forms=` + formType
	finalUrl := neturl.URL{
		Scheme:   `https`,
		Host:     `efts.sec.gov`,
		Path:     `LATEST/search-index`,
		RawQuery: queryText,
	}
	body, err := actions.Scrape(finalUrl.String(), throttle.Throttle)
	if err != nil {
		return nil, err
	}
	var resp Edgar2025Response
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	data := SearchResult{
		SearchText: search,
		Results: make([]struct {
			Title                  string
			Summary                string
			EdgarUrl               string
			PercentEncodedEdgarUrl string
		}, 0),
	}
	for _, hit := range resp.Hits.Hits {
		if len(hit.Source.Ciks) <= 0 {
			continue
		}
		if hit.Source.FileType != formType {
			continue
		}
		name := "UNKNOWN"
		if len(hit.Source.DisplayNames) > 0 {
			name = hit.Source.DisplayNames[0]
		}
		fileDateStr := ""
		fileDate, err := time.Parse("2006-01-02", hit.Source.FileDate)
		if err != nil {
			fileDateStr = "UNKNOWN"
		}
		fileDateStr = fileDate.Format("01/02/2006")
		trimmedCik := strings.TrimLeft(hit.Source.Ciks[0], "0")
		edgarUrl := `/Archives/edgar/data/` + trimmedCik + `/` + strings.ReplaceAll(hit.Source.Adsh, "-", "") + `/` + hit.Source.Adsh +
			`-index.htm`
		data.Results = append(data.Results, struct {
			Title                  string
			Summary                string
			EdgarUrl               string
			PercentEncodedEdgarUrl string
		}{
			Title: hit.Source.FileType + ` - ` + name,
			Summary: `Filed Date: ` + fileDateStr + ` Accession Number: ` +
				hit.Source.Adsh,
			EdgarUrl:               edgarUrl,
			PercentEncodedEdgarUrl: neturl.QueryEscape(edgarUrl),
		})
	}
	return &data, nil
}

func getLegacyFilings(search string, formType string, year string) (*SearchResult, error) {
	edgarSearchText := `company-name%3D"` + neturl.QueryEscape(search) + `"%20AND%20form-type%3D%28` + neturl.QueryEscape(formType) + `%2A%29`
	queryText := `text=` + edgarSearchText + `&start=1&count=80&first=` + year + `&last=1994&output=atom`
	finalUrl := neturl.URL{
		Scheme:   `https`,
		Host:     `www.sec.gov`,
		Path:     `cgi-bin/srch-edgar`,
		RawQuery: queryText,
	}
	body, err := actions.Scrape(finalUrl.String(), throttle.Throttle)
	if err != nil {
		return nil, err
	}
	reader := bytes.NewReader(body)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	rss := EdgarRss{}
	err = decoder.Decode(&rss)
	if err != nil {
		return nil, err
	}
	data := SearchResult{
		SearchText: search,
		Results: make([]struct {
			Title                  string
			Summary                string
			EdgarUrl               string
			PercentEncodedEdgarUrl string
		}, 0),
	}
	for _, entry := range rss.Entry {
		edgarUrl := ""
		for _, link := range entry.Link {
			for _, linkAttr := range link.XMLAttrs {
				if linkAttr.Name.Local == "href" {
					edgarUrl = linkAttr.Value
					break
				}
			}
		}
		if edgarUrl == "" {
			continue
		}
		data.Results = append(data.Results, struct {
			Title                  string
			Summary                string
			EdgarUrl               string
			PercentEncodedEdgarUrl string
		}{
			Title:                  entry.Title[0].CharData,
			Summary:                entry.Summary[0].CharData,
			EdgarUrl:               edgarUrl,
			PercentEncodedEdgarUrl: neturl.QueryEscape(edgarUrl),
		})
	}
	return &data, nil
}
