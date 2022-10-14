package generator

import (
	"bytes"
	"embed"
	"github.com/pkg/errors"
	"golang.org/x/tools/imports"
	"io/ioutil"
	"text/template"
)

//go:embed ../templates/*
var templateDir embed.FS

func getTemplate(name string) (string, error) {
	fullName := "templates/" + name + ".tmpl"
	file, err := templateDir.ReadFile(fullName)
	if err != nil {
		return "", errors.Wrap(err, "templateDir.ReadFile")
	}
	return string(file), nil
}

func GenToTemplate(templateName string, toFileName string, value Gen) error {
	templateFile,err := getTemplate(templateName)
	if err != nil {
		return errors.Wrap(err, "getTemplate")
	}
	t, err := template.New("header").Parse(templateFile)
	if err != nil {
		err = errors.Wrap(err, "template.New")
		return err
	}

	buf := bytes.NewBuffer([]byte{})

	err = t.Execute(buf, value)
	if err != nil {
		err = errors.Wrap(err, "t.Execute")
		return err
	}


	processedSource, err := imports.Process(toFileName, buf.Bytes(), nil)
	if err != nil {
		err = errors.Wrap(err, "imports.Process")
		return err
	}
	buf = bytes.NewBuffer([]byte{})
	_, err = buf.Write(processedSource)

	err = ioutil.WriteFile(toFileName, buf.Bytes(), 0664)
	if err != nil {
		err = errors.Wrap(err, "ioutil.WriteFile")
		return err
	}
	return nil
}