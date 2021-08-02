package telefacts_sec_test

import (
	"net/url"
	"time"

	"github.com/joshuanario/r8lmt"
)

var (
	out       chan interface{} = make(chan interface{})
	in        chan interface{} = make(chan interface{})
	dur       time.Duration    = 1 * time.Second
	throttled bool             = false
)

func startSECThrottle() {
	if !throttled {
		r8lmt.Throttler(out, in, dur, true)
		throttled = true
	}
}

func throttle(urlString string) {
	urlStruct, err := url.Parse(urlString)
	if urlStruct.Hostname() != "sec.gov" {
		return
	}
	if err != nil {
		return
	}
	in <- struct{}{}
	<-out
}
