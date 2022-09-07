package serializables

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"ecksbee.com/telefacts-sec/internal/actions"
	"ecksbee.com/telefacts-taxonomy-package/pkg/taxonomies"
	underscore "ecksbee.com/telefacts/pkg/serializables"
	"golang.org/x/net/html/charset"
)

const xmlExt = ".xml"
const xsdExt = ".xsd"
const preExt = "_pre.xml"
const defExt = "_def.xml"
const calExt = "_cal.xml"
const labExt = "_lab.xml"
const regexSEC = "https://www.sec.gov/Archives/edgar/data/([0-9]+)/([0-9]+)"

func Download(filingURL string, wd string, gts string, throttle func(string)) (string, error) {
	isSEC, _ := regexp.MatchString(regexSEC, filingURL)
	if !isSEC {
		return "", fmt.Errorf("not an acceptable SEC address, " + filingURL)
	}
	body, err := actions.Scrape(filingURL+"/FilingSummary.xml", throttle)
	if err != nil {
		return "", err
	}
	reader := bytes.NewReader(body)
	decoder := xml.NewDecoder(reader)
	decoder.CharsetReader = charset.NewReaderLabel
	filingSummary := FilingSummary{}
	err = decoder.Decode(&filingSummary)
	if len(filingSummary.InputFiles) <= 0 || len(filingSummary.InputFiles[0].File) <= 0 || err != nil {
		return "", fmt.Errorf("empty filing at "+filingURL+". %s\n\n%v", string(body), err)
	}
	entry := filingSummary.GetInstance()
	srcDoc := filingSummary.GetIxbrl()
	if srcDoc != "" {
		entry = srcDoc
	}
	err = os.MkdirAll(filepath.Join(wd, "folders"), 0755)
	if err != nil {
		return "", err
	}
	underscore.WorkingDirectoryPath = wd
	id, err := underscore.NewFolder(underscore.Underscore{
		Entry: entry,
		Note:  filingURL,
	})
	if err != nil {
		return "", err
	}
	workingDir := filepath.Join(wd, "folders", id)
	var wg sync.WaitGroup
	wg.Add(6)
	go func() {
		defer wg.Done()
		schemaName := filingSummary.GetSchema()
		schema, err := actions.Scrape(filingURL+"/"+schemaName, throttle)
		if err != nil {
			return
		}
		dest := path.Join(workingDir, schemaName)
		err = actions.WriteFile(dest, schema)
		if err == nil {
			decoded, err := underscore.DecodeSchemaFile(schema)
			if err != nil {
				return
			}
			taxonomies.VolumePath = gts
			underscore.GlobalTaxonomySetPath = gts
			taxonomies.ImportSchema(decoded)
		}
	}()
	go func() {
		defer wg.Done()
		targetUrl := filingURL + "/" + entry
		dest := path.Join(workingDir, entry)
		err = scrapeAndWrite(targetUrl, dest, throttle)
		if srcDoc == entry {
			insDoc := strings.Replace(srcDoc, ".htm", "_htm", 1)
			targetUrl = filingURL + "/" + insDoc + ".xml"
			err = scrapeAndWrite(targetUrl, dest+".xml", throttle)
		}
	}()
	go func() {
		defer wg.Done()
		preItem := filingSummary.GetPresentationLinkbase()
		targetUrl := filingURL + "/" + preItem
		dest := path.Join(workingDir, preItem)
		err = scrapeAndWrite(targetUrl, dest, throttle)
	}()
	go func() {
		defer wg.Done()
		defItem := filingSummary.GetDefinitionLinkbase()
		targetUrl := filingURL + "/" + defItem
		dest := path.Join(workingDir, defItem)
		err = scrapeAndWrite(targetUrl, dest, throttle)
	}()
	go func() {
		defer wg.Done()
		calItem := filingSummary.GetCalculationLinkbase()
		targetUrl := filingURL + "/" + calItem
		dest := path.Join(workingDir, calItem)
		err = scrapeAndWrite(targetUrl, dest, throttle)
	}()
	go func() {
		defer wg.Done()
		labItem := filingSummary.GetLabelLinkbase()
		targetUrl := filingURL + "/" + labItem
		dest := path.Join(workingDir, labItem)
		err = scrapeAndWrite(targetUrl, dest, throttle)
	}()
	wg.Wait()
	return id, err
}

func scrapeAndWrite(url string, dest string, throttle func(string)) error {
	scraped, err := actions.Scrape(url, throttle)
	if err != nil {
		return err
	}
	return actions.WriteFile(dest, scraped)
}
