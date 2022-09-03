package main

import (
	"context"
	"flag"
	"fmt"
	"path"
	"time"

	"ecksbee.com/telefacts-sec/internal/web"
	"ecksbee.com/telefacts-sec/pkg/names"
	"ecksbee.com/telefacts-sec/pkg/serializables"
	"ecksbee.com/telefacts-sec/pkg/throttle"
)

func main() {
	namesPtr := flag.Bool("install-names", false, "if true, this application will run a command to install the names registry from US SEC's EDGAR system")
	var edgarUrl string
	flag.StringVar(&edgarUrl, "EDGAR-URL", "", "if set to a url in US SEC's EDGAR system, this application will run a command that scrapes XBRL files from that URL.")
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
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
	var ctx = context.Background()
	web.SetupAndListen(ctx, wait)
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
