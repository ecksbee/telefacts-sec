package main

import (
	"fmt"
	"path"

	"ecksbee.com/telefacts-sec/pkg/names"
	"ecksbee.com/telefacts-sec/pkg/throttle"
)

func main() {
	names.NamePath = path.Join(".", "data", "/names.json")
	throttle.StartSECThrottle()
	err := names.MergeNames(throttle.Throttle)
	if err != nil {
		fmt.Println(err.Error())
	}
}
