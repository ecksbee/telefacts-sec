package serializables

import (
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"
	"sync"

	"ecksbee.com/telefacts-sec/internal/actions"
	underscore "ecksbee.com/telefacts/pkg/serializables"
)

type filingItem struct {
	LastModified string `json:"last-modified"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Size         string `json:"size"`
}

const xmlExt = ".xml"
const xsdExt = ".xsd"
const preExt = "_pre.xml"
const defExt = "_def.xml"
const calExt = "_cal.xml"
const labExt = "_lab.xml"
const regexSEC = "https://www.sec.gov/Archives/edgar/data/([0-9]+)/([0-9]+)"

func Download(filingURL string, dir string, throttle func(string)) error {
	isSEC, _ := regexp.MatchString(regexSEC, filingURL)
	if !isSEC {
		return fmt.Errorf("not an acceptable SEC address, " + filingURL)
	}
	body, err := actions.Scrape(filingURL+"/index.json", throttle)
	if err != nil {
		return err
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
		return fmt.Errorf("empty filing at "+filingURL+". %s\n\n%v", string(body), err)
	}
	schemaItem, err := getSchemaFromFilingItems(items)
	if err != nil {
		return err
	}
	str := schemaItem.Name
	x := strings.Index(str, "-")
	ticker := str[:x]
	if len(ticker) <= 0 {
		return fmt.Errorf("ticker symbol not found")
	}
	instance, err := getInstanceFromFilingItems(items, ticker)
	if err != nil {
		return err
	}
	underscore.VolumePath = dir
	id, err := underscore.NewFolder(underscore.Underscore{
		Entry:    instance.Name,
		Checksum: "",
		Ixbrl:    "",
		Note:     filingURL,
	})
	if err != nil {
		return err
	}
	workingDir := path.Join(dir, "folders", id)
	//todo make sure folders exists
	var wg sync.WaitGroup
	wg.Add(6)
	go func() {
		defer wg.Done()
		schema, err := actions.Scrape(filingURL+"/"+schemaItem.Name, throttle)
		if err != nil {
			return
		}
		dest := path.Join(workingDir, schemaItem.Name)
		err = actions.WriteFile(dest, schema)
	}()
	go func() {
		defer wg.Done()
		targetUrl := filingURL + "/" + instance.Name
		dest := path.Join(workingDir, instance.Name)
		err = scrapeAndWrite(targetUrl, dest, throttle)
	}()
	go func() {
		defer wg.Done()
		preItem, err := getPresentationLinkbaseFromFilingItems(items, ticker)
		if err != nil {
			return
		}
		targetUrl := filingURL + "/" + preItem.Name
		dest := path.Join(workingDir, preItem.Name)
		err = scrapeAndWrite(targetUrl, dest, throttle)
	}()
	go func() {
		defer wg.Done()
		defItem, err := getDefinitionLinkbaseFromFilingItems(items, ticker)
		if err != nil {
			return
		}
		targetUrl := filingURL + "/" + defItem.Name
		dest := path.Join(workingDir, defItem.Name)
		err = scrapeAndWrite(targetUrl, dest, throttle)
	}()
	go func() {
		defer wg.Done()
		calItem, err := getCalculationLinkbaseFromFilingItems(items, ticker)
		if err != nil {
			return
		}
		targetUrl := filingURL + "/" + calItem.Name
		dest := path.Join(workingDir, calItem.Name)
		err = scrapeAndWrite(targetUrl, dest, throttle)
	}()
	go func() {
		defer wg.Done()
		labItem, err := getLabelLinkbaseFromFilingItems(items, ticker)
		if err != nil {
			return
		}
		targetUrl := filingURL + "/" + labItem.Name
		dest := path.Join(workingDir, labItem.Name)
		err = scrapeAndWrite(targetUrl, dest, throttle)
	}()
	wg.Wait()
	return err
}

func scrapeAndWrite(url string, dest string, throttle func(string)) error {
	scraped, err := actions.Scrape(url, throttle)
	if err != nil {
		return err
	}
	return actions.WriteFile(dest, scraped)
}
