package names

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func Scrape(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, resp.Body)
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func ScrapeNames() (map[string]map[string]string, error) {
	tickerURL := `https://www.sec.gov/files/company_tickers.json`
	ret := make(map[string]map[string]string)
	b, err := Scrape(tickerURL)
	if err != nil {
		return ret, err
	}
	type SECTickers map[string]struct {
		CIK    int    `json:"cik_str"`
		Ticker string `json:"ticker"`
		Title  string `json:"title"`
	}
	var f SECTickers
	err = json.Unmarshal(b, &f)
	if err != nil {
		return ret, err
	}
	cik := "http://www.sec.gov/CIK"
	ret[cik] = make(map[string]string)
	for _, obj := range f {
		ret[cik][fmt.Sprintf("%010d", obj.CIK)] = obj.Title
	}
	return ret, err
}
