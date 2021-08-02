package actions

import (
	"archive/zip"
	"bytes"
	"io/fs"
	"io/ioutil"
	"path"
)

func Zip(dir string, files []fs.FileInfo) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := zip.NewWriter(buf)
	for _, file := range files {
		filename := file.Name()
		filepath := path.Join(dir, filename)
		data, err := ioutil.ReadFile(filepath)
		if err != nil {
			return nil, err
		}
		f, err := writer.Create(file.Name())
		if err != nil {
			return nil, err
		}
		_, err = f.Write([]byte(data))
		if err != nil {
			return nil, err
		}

	}
	err := writer.Close()
	return buf.Bytes(), err
}
