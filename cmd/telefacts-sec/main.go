package main

import (
	"flag"
	"fmt"
	"path"

	"ecksbee.com/telefacts-sec/pkg/names"
	"ecksbee.com/telefacts-sec/pkg/serializables"
	"ecksbee.com/telefacts-sec/pkg/throttle"
)

func main() {
	namesPtr := flag.Bool("install-names", false, "if true, this application will run a command to install the names registry from US SEC's EDGAR system")
	var edgarUrl string
	flag.StringVar(&edgarUrl, "EDGAR-URL", "", "if set to a url in US SEC's EDGAR system, this application will run a command that scrapes XBRL files from that URL.")
	flag.Parse()
	if edgarUrl != "" && *namesPtr {
		panic("telefacts-sec cannot have both flags enabled: EDGAR-URL and install-names")
	}
	if *namesPtr {
		fmt.Println("initiating install-names")
		installNamesRegistry()
		fmt.Println("install-names complete")
		return
	}
	if edgarUrl != "" {
		fmt.Println("scraping EDGAR")
		scrapeEDGAR(edgarUrl)
		fmt.Println("scraping complete")
		return
	}
	panic("telefacts-sec is a cli")
}

func installNamesRegistry() {
	names.NamePath = path.Join(".", "data", "/names.json")
	throttle.StartSECThrottle()
	err := names.MergeNames(throttle.Throttle)
	if err != nil {
		panic(err)
	}
}

func scrapeEDGAR(url string) {
	throttle.StartSECThrottle()
	wd := path.Join(".", "wd")
	gts := path.Join(".", "gts")
	err := serializables.Download(url, wd, gts, throttle.Throttle)
	if err != nil {
		panic(err)
	}
}
