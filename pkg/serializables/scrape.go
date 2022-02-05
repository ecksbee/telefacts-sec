package serializables

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"ecksbee.com/telefacts-sec/internal/actions"
)

func Scrape(filingURL string, throttle func(string)) ([]byte, error) {
	isSEC, _ := regexp.MatchString(regexSEC, filingURL)
	if !isSEC {
		return nil, fmt.Errorf("not an acceptable SEC address, " + filingURL)
	}
	body, err := actions.Scrape(filingURL+"/index.json", throttle)
	if err != nil {
		return nil, err
	}
	filing := struct {
		Directory struct {
			Item      []filingItem `json:"item"`
			Name      string       `json:"name"`
			ParentDir string       `json:"parent-dir"`
		} `json:"directory"`
	}{}
	err = json.Unmarshal(body, &filing)
	items := filing.Directory.Item
	if len(items) <= 0 || err != nil {
		return nil, fmt.Errorf("empty filing at "+filingURL+". %s\n\n%v", string(body), err)
	}
	schemaItem, err := getSchemaFromFilingItems(items)
	if err != nil {
		return nil, err
	}
	str := schemaItem.Name
	x := strings.Index(str, "-")
	ticker := str[:x]
	if len(ticker) <= 0 {
		return nil, fmt.Errorf("ticker symbol not found")
	}
	instance, err := getInstanceFromFilingItems(items, ticker)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	wg.Add(6)
	var (
		schema        []byte
		instanceBytes []byte
		presentation  []byte
		definition    []byte
		calculation   []byte
		label         []byte
		preItem       *filingItem
		defItem       *filingItem
		calItem       *filingItem
		labItem       *filingItem
	)
	go func() {
		defer wg.Done()
		targetUrl := filingURL + "/" + schemaItem.Name
		schema, err = actions.Scrape(targetUrl, throttle)
	}()
	go func() {
		defer wg.Done()
		targetUrl := filingURL + "/" + instance.Name
		instanceBytes, err = actions.Scrape(targetUrl, throttle)
	}()
	go func() {
		defer wg.Done()
		preItem, err = getPresentationLinkbaseFromFilingItems(items, ticker)
		if err != nil {
			return
		}
		targetUrl := filingURL + "/" + preItem.Name
		presentation, err = actions.Scrape(targetUrl, throttle)
	}()
	go func() {
		defer wg.Done()
		defItem, err = getDefinitionLinkbaseFromFilingItems(items, ticker)
		if err != nil {
			return
		}
		targetUrl := filingURL + "/" + defItem.Name
		definition, err = actions.Scrape(targetUrl, throttle)
	}()
	go func() {
		defer wg.Done()
		calItem, err = getCalculationLinkbaseFromFilingItems(items, ticker)
		if err != nil {
			return
		}
		targetUrl := filingURL + "/" + calItem.Name
		calculation, err = actions.Scrape(targetUrl, throttle)
	}()
	go func() {
		defer wg.Done()
		labItem, err = getLabelLinkbaseFromFilingItems(items, ticker)
		if err != nil {
			return
		}
		targetUrl := filingURL + "/" + labItem.Name
		label, err = actions.Scrape(targetUrl, throttle)
	}()
	wg.Wait()
	if err != nil {
		return nil, err
	}
	return zipData([]struct {
		Name string
		Body []byte
	}{
		{schemaItem.Name, schema},
		{instance.Name, instanceBytes},
		{preItem.Name, presentation},
		{defItem.Name, definition},
		{calItem.Name, calculation},
		{labItem.Name, label},
	})
}

func zipData(files []struct {
	Name string
	Body []byte
}) ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)
	for _, file := range files {
		zipFile, err := zipWriter.Create(file.Name)
		if err != nil {
			return nil, err
		}
		_, err = zipFile.Write(file.Body)
		if err != nil {
			return nil, err
		}
	}
	err := zipWriter.Close()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
