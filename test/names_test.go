package telefacts_sec_test

import (
	"testing"

	"ecksbee.com/telefacts-sec/pkg/names"
)

func TestMergeNames(t *testing.T) {
	startSECThrottle()
	err := names.MergeNames(throttle)
	if err != nil {
		t.Fatalf(err.Error())
	}
}
