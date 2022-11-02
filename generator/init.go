package generator

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/tools/imports"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

func Init(files map[string]string, outputDir string, name string, process bool) error {
	dir := filepath.Join(outputDir, name)
	_, err := os.Stat(dir)
	if err == nil {
		return fmt.Errorf("dir %s already exists", dir)
	}

	if !os.IsNotExist(err) {
		return errors.Wrap(err, "stat dir")
	}

	for k, v := range files {
		toFileName := filepath.Join(dir, v)

		templateFile, err := getTemplate(k)
		if err != nil {
			return errors.Wrap(err, "getTemplate")
		}

		t, err := template.New("header").Parse(templateFile)
		if err != nil {
			err = errors.Wrap(err, "template.New")
			return err
		}

		buf := bytes.NewBuffer([]byte{})

		err = t.Execute(buf, map[string]interface{}{
			"name": name,
		})
		if err != nil {
			err = errors.Wrap(err, "t.Execute")
			return err
		}

		if process {
			processedSource, err := imports.Process(toFileName, buf.Bytes(), nil)
			if err != nil {
				//log.Println(buf.String())
				err = errors.Wrap(err, "imports.Process")
				log.Println(err.Error())
			}

			buf = bytes.NewBuffer([]byte{})
			_, err = buf.Write(processedSource)
		}

		err = ioutil.WriteFile(toFileName, buf.Bytes(), 0664)
		if err != nil {
			err = errors.Wrap(err, "ioutil.WriteFile")
			return err
		}

	}
	return nil
}
