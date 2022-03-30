package gowrap

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"text/template"
	"time"
	"unicode"

	"github.com/fitan/gowrap/generator"
	"github.com/fitan/gowrap/pkg"
	"github.com/pkg/errors"
)

//go:embed templates/*
var tmp embed.FS

//GenerateCommand implements Command interface
type GenerateCommand struct {
	BaseCommand

	interfaceName string
	template      string
	outputFile    string
	sourcePkg     string
	noGenerate    bool
	pkgNeedSyntax bool
	vars          vars
	localPrefix   string

	loader        templateLoader
	filepath      fs
	batchTemplate string
	combo         string
}

var serviceCombo []string = []string{"service", "new"}

//NewGenerateCommand creates GenerateCommand
func NewGenerateCommand(l remoteTemplateLoader) *GenerateCommand {
	gc := &GenerateCommand{
		loader: loader{fileReader: ioutil.ReadFile, remoteLoader: l},
		filepath: fs{
			Rel:       filepath.Rel,
			Abs:       filepath.Abs,
			Dir:       filepath.Dir,
			WriteFile: ioutil.WriteFile,
		},
	}

	//this flagset loads flags values to the command fields
	fs := &flag.FlagSet{}
	// load pkg needsyntax slow。
	fs.BoolVar(&gc.pkgNeedSyntax, "ps", true, "load pkg needsyntax")
	fs.BoolVar(&gc.noGenerate, "g", false, "don't put //go:generate instruction to the generated code")
	fs.StringVar(&gc.interfaceName, "i", "", `the source interface name, i.e. "Reader"`)
	fs.StringVar(&gc.batchTemplate, "bt", "", `the source interface name, i.e. "Reader"`)
	fs.StringVar(&gc.sourcePkg, "p", "", "the source package import path, i.e. \"io\", \"github.com/fitan/gowrap\" or\na relative import path like \"./generator\"")
	fs.StringVar(&gc.outputFile, "o", "", "the output file name")
	fs.StringVar(&gc.template, "t", "", "the template to use, it can be an HTTPS URL, local file or a\nreference to a template in gowrap repository,\n"+
		"run `gowrap template list` for details")
	fs.Var(&gc.vars, "v", "a key-value pair to parametrize the template,\narguments without an equal sign are treated as a bool values,\ni.e. -v foo=bar -v disableChecks")
	fs.StringVar(&gc.localPrefix, "l", "", "put imports beginning with this string after 3rd-party packages; comma-separated list")
	fs.StringVar(&gc.combo, "init", "", "initialize template")

	gc.BaseCommand = BaseCommand{
		Short: "generate decorators",
		Usage: "-p package -i interfaceName -bt",
		Flags: fs,
	}

	return gc
}

// Run implements Command interface
func (gc *GenerateCommand) Run(args []string, stdout io.Writer) error {
	totalT := time.Now()
	defer func() {
		log.Printf("total time: %v", time.Now().Sub(totalT).String())
	}()

	if err := gc.FlagSet().Parse(args); err != nil {
		return CommandLineError(err.Error())
	}

	if err := gc.checkFlags(); err != nil {
		return err
	}

	var ops []generator.Options
	var err error
	if gc.combo != "" {
		ops, err = gc.getInitOptions(gc.combo)
		if err != nil {
			return err
		}

		gens, err := generator.NewGeneratorInit(ops)
		if err != nil {
			return err
		}

		for _, gen := range gens {
			err := gen.Generate()
			if err != nil {
				return err
			}
		}
		return nil
	}

	ops, err = gc.getOptions()
	if err != nil {
		return err
	}

	t1 := time.Now()
	gens, err := generator.NewGenerator(ops)
	if err != nil {
		return err
	}
	log.Printf("generator.NewGenerator time: %v", time.Now().Sub(t1).String())

	t2 := time.Now()
	g := sync.WaitGroup{}
	for _, gen := range gens {
		g.Add(1)
		go func(gen *generator.Generator) {
			defer g.Done()
			err := gen.Generate()
			if err != nil {
				log.Fatalf("generate err: %s", err.Error())
			}
		}(gen)
	}
	g.Wait()
	log.Printf("gen.Generate time: %v", time.Now().Sub(t2).String())

	//buf := bytes.NewBuffer([]byte{})
	//
	//if err := gen.Generate(buf); err != nil {
	//	return err
	//}
	return nil

}

var (
	errNoOutputFile    = CommandLineError("output file is not specified")
	errNoInterfaceName = CommandLineError("interface name is not specified")
	errNoTemplate      = CommandLineError("no template specified")
	errMustBeOrdered   = CommandLineError("Must be ordered bt or init")
)

func (gc *GenerateCommand) checkFlags() error {
	if gc.batchTemplate == "" && gc.combo == "" {
		return errNoTemplate
	}
	//if gc.outputFile == "" {
	//	return errNoOutputFile
	//}
	//
	//if gc.interfaceName == "" {
	//	return errNoInterfaceName
	//}
	//
	//if gc.template == "" {
	//	return errNoTemplate
	//}

	return nil
}

func (gc *GenerateCommand) getInitOptions(serviceName string) ([]generator.Options, error) {
	ops := make([]generator.Options, 0, 0)
	err := os.Mkdir(serviceName, os.ModePerm)
	if err != nil {
		return nil, err
	}

	cmdDir, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "os.Getwd")
	}

	for _, v := range serviceCombo {
		bodyTemplate := v
		outputFile := fmt.Sprintf("./%s/%s_%s.go", serviceName, serviceName, v)

		options := generator.Options{
			InterfaceName:  serviceName,
			OutputFile:     outputFile,
			Funcs:          helperFuncs,
			HeaderTemplate: headerTemplate,
			HeaderVars: map[string]interface{}{
				"DisableGoGenerate": gc.noGenerate,
				"OutputFileName":    filepath.Base(outputFile),
				"VarsArgs":          varsToArgs(gc.vars),
			},
			Vars:          gc.vars.toMap(),
			LocalPrefix:   gc.localPrefix,
			PkgNeedSyntax: false,
			RunCmdDir:     cmdDir,
		}

		outputFileDir, err := gc.filepath.Abs(gc.filepath.Dir(outputFile))
		if err != nil {
			return nil, err
		}

		gc.sourcePkg = serviceName

		options.BodyTemplate, options.HeaderVars["Template"], err = gc.embedLoadTemplate(bodyTemplate, outputFileDir)
		if err != nil {
			return nil, err
		}
		ops = append(ops, options)
	}
	return ops, err
}

func (gc *GenerateCommand) getOptions() ([]generator.Options, error) {
	ops := make([]generator.Options, 0, 0)

	sourcePackage, err := pkg.Load(gc.sourcePkg, true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to load source package")
	}

	cmdDir, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "os.Getwd")
	}

	btl := strings.Split(gc.batchTemplate, " ")
	for _, v := range btl {
		vl := strings.Split(v, ":")
		if len(vl) != 2 {
			return nil, fmt.Errorf("format is wrong: %s", v)
		}
		bodyTemplate := vl[0]
		outputFile := vl[1]

		options := generator.Options{
			InterfaceName:  gc.interfaceName,
			OutputFile:     outputFile,
			Funcs:          helperFuncs,
			HeaderTemplate: headerTemplate,
			HeaderVars: map[string]interface{}{
				"DisableGoGenerate": gc.noGenerate,
				"OutputFileName":    filepath.Base(outputFile),
				"VarsArgs":          varsToArgs(gc.vars),
			},
			Vars:          gc.vars.toMap(),
			LocalPrefix:   gc.localPrefix,
			PkgNeedSyntax: gc.pkgNeedSyntax,
			RunCmdDir:     cmdDir,
		}

		outputFileDir, err := gc.filepath.Abs(gc.filepath.Dir(outputFile))
		if err != nil {
			return nil, err
		}

		if gc.sourcePkg == "" {
			gc.sourcePkg = "./"
		}

		options.SourceLoadPkg = sourcePackage

		options.SourcePackage = sourcePackage.PkgPath
		//options.BodyTemplate, options.HeaderVars["Template"], err = gc.loadTemplate(bodyTemplate, outputFileDir)
		options.BodyTemplate, options.HeaderVars["Template"], err = gc.embedLoadTemplate(bodyTemplate, outputFileDir)
		if err != nil {
			return nil, err
		}
		ops = append(ops, options)

	}

	return ops, err
}

type readerFunc func(path string) ([]byte, error)

type loader struct {
	fileReader   readerFunc
	remoteLoader templateLoader
}

func (gc *GenerateCommand) embedLoadTemplate(template string, outputFileDir string) (tmpl string, url string, err error) {
	fullName := "templates/" + template + ".tmpl"
	file, err := tmp.ReadFile(fullName)
	if err != nil {
		return "", "", err
	}
	return string(file), "", nil
}

func (gc *GenerateCommand) loadTemplate(template string, outputFileDir string) (contents, url string, err error) {
	body, url, err := gc.loader.Load(template)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to load template")
	}

	if !strings.HasPrefix(url, "https://") {
		templatePath, err := gc.filepath.Abs(url)
		if err != nil {
			return "", "", err
		}

		url, err = gc.filepath.Rel(outputFileDir, templatePath)
		if err != nil {
			return "", "", err
		}
	}

	return string(body), url, nil
}

// Load implements templateLoader
func (l loader) Load(template string) (tmpl []byte, url string, err error) {
	tmpl, err = l.fileReader(template)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}

		return l.remoteLoader.Load(template)
	}

	return tmpl, template, err
}

type templateLoader interface {
	Load(path string) (tmpl []byte, url string, err error)
}

type fs struct {
	Rel       func(string, string) (string, error)
	Abs       func(string) (string, error)
	Dir       func(string) string
	WriteFile func(string, []byte, os.FileMode) error
}

type varFlag struct {
	name  string
	value interface{}
}

//vars is a helper type that implements flag.Value to read multiple vars from the command line
type vars []varFlag

//String implements flag.Value
func (v vars) String() string {
	return fmt.Sprintf("%#v", v)
}

func (v *vars) Set(s string) error {
	chunks := strings.SplitN(s, "=", 2)
	switch len(chunks) {
	case 1:
		*v = append(*v, varFlag{name: chunks[0], value: true})
	case 2:
		*v = append(*v, varFlag{name: chunks[0], value: chunks[1]})
	}

	return nil
}

func (v vars) toMap() map[string]interface{} {
	m := make(map[string]interface{}, len(v))
	for _, vf := range v {
		m[vf.name] = vf.value
	}

	return m
}

func varsToArgs(v vars) string {
	if len(v) == 0 {
		return ""
	}

	var ss []string

	for _, vf := range v {
		switch typedValue := vf.value.(type) {
		case string:
			ss = append(ss, vf.name+"="+typedValue)
		case bool:
			ss = append(ss, vf.name)
		}
	}

	return " -v " + strings.Join(ss, " -v ")
}

var helperFuncs = template.FuncMap{
	"up":        strings.ToUpper,
	"down":      strings.ToLower,
	"upFirst":   upFirst,
	"downFirst": downFirst,
	"replace":   strings.ReplaceAll,
	"snake":     toSnakeCase,
}

func upFirst(s string) string {
	for _, v := range s {
		return string(unicode.ToUpper(v)) + s[len(string(v)):]
	}
	return ""
}

func downFirst(s string) string {
	for _, v := range s {
		return string(unicode.ToLower(v)) + s[len(string(v)):]
	}
	return ""
}

var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

func toSnakeCase(str string) string {
	result := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	result = matchAllCap.ReplaceAllString(result, "${1}_${2}")
	return strings.ToLower(result)
}

const headerTemplate = `package {{.Package.Name}}

// Code generated by gowrap. DO NOT EDIT.
// template: {{.Options.HeaderVars.Template}}
// gowrap: http://github.com/fitan/gowrap

{{if (not .Options.HeaderVars.DisableGoGenerate)}}
//{{"go:generate"}} gowrap gen -p {{.SourcePackage.PkgPath}} -i {{.Options.InterfaceName}} -t {{.Options.HeaderVars.Template}} -o {{.Options.HeaderVars.OutputFileName}}{{.Options.HeaderVars.VarsArgs}} -l "{{.Options.LocalPrefix}}"
{{end}}

`
