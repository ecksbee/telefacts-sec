package serializables

import (
	"ecksbee.com/telefacts-sec/internal/actions"
)

func Zip(workingDir string) ([]byte, error) {
	files, err := GetOSFiles(workingDir)
	if err != nil {
		return nil, err
	}
	return actions.Zip(workingDir, files)
}
