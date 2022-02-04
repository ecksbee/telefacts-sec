package names

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

func UnmarshalNames() (map[string]map[string]string, error) {
	if NamePath == "" {
		return nil, fmt.Errorf("empty NamePath")
	}
	names := make(map[string]map[string]string)
	if _, err := os.Stat(NamePath); os.IsNotExist(err) {
		return names, nil
	}
	b, err := ioutil.ReadFile(NamePath)
	if err != nil {
		return names, err
	}
	err = json.Unmarshal(b, &names)
	return names, err
}
