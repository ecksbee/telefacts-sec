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

const ixbrlExt = ".htm"
const xmlExt = ".xml"
const xsdExt = ".xsd"
const preExt = "_pre.xml"
const defExt = "_def.xml"
const calExt = "_cal.xml"
const labExt = "_lab.xml"
const regexSEC = "https://www.sec.gov/Archives/edgar/data/([0-9]+)/([0-9]+)"

func Scrape(filingURL string, dir string, throttle func(string)) error {
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
	schemaItem, err := getSchemaFromFilingItems(items) //todo scrape and unzip .zip filing item
	if err != nil {
		return err
	}
	str := schemaItem.Name
	x := strings.Index(str, "-")
	ticker := str[:x]
	if len(ticker) <= 0 {
		return fmt.Errorf("ticker symbol not found")
	}
	var entry string
	ixbrlItem, err := getIxbrlFileFromFilingItems(items, ticker)
	if err != nil && err.Error() == "cannot identify a single ixbrl file" {
		instance, err := getInstanceFromFilingItems(items, ticker)
		if err != nil {
			return err
		}
		entry = instance.Name
	} else {
		entry = ixbrlItem.Name
	}
	underscore.VolumePath = dir
	id, err := underscore.NewFolder(underscore.Underscore{
		Entry:    ixbrlItem.Name,
		Checksum: "",
		Note:     filingURL,
	})
	if err != nil {
		return err
	}
	workingDir := path.Join(dir, "folders", id)
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
		entryFile, err := actions.Scrape(filingURL+"/"+entry, throttle)
		if err != nil {
			return
		}
		dest := path.Join(workingDir, ixbrlItem.Name)
		err = actions.WriteFile(dest, entryFile)
	}()
	go func() {
		defer wg.Done()
		preItem, err := getPresentationLinkbaseFromFilingItems(items, ticker)
		if err != nil {
			return
		}
		presentation, err := actions.Scrape(filingURL+"/"+preItem.Name, throttle)
		if err != nil {
			return
		}
		dest := path.Join(workingDir, preItem.Name)
		err = actions.WriteFile(dest, presentation)
	}()
	go func() {
		defer wg.Done()
		defItem, err := getDefinitionLinkbaseFromFilingItems(items, ticker)
		if err != nil {
			return
		}
		definition, err := actions.Scrape(filingURL+"/"+defItem.Name, throttle)
		if err != nil {
			return
		}
		dest := path.Join(workingDir, defItem.Name)
		err = actions.WriteFile(dest, definition)
	}()
	go func() {
		defer wg.Done()
		calItem, err := getCalculationLinkbaseFromFilingItems(items, ticker)
		if err != nil {
			return
		}
		calculation, err := actions.Scrape(filingURL+"/"+calItem.Name, throttle)
		if err != nil {
			return
		}
		dest := path.Join(workingDir, calItem.Name)
		err = actions.WriteFile(dest, calculation)
	}()
	go func() {
		defer wg.Done()
		labItem, err := getLabelLinkbaseFromFilingItems(items, ticker)
		if err != nil {
			return
		}
		label, err := actions.Scrape(filingURL+"/"+labItem.Name, throttle)
		if err != nil {
			return
		}
		dest := path.Join(workingDir, labItem.Name)
		err = actions.WriteFile(dest, label)
	}()
	wg.Wait()
	return err
}
