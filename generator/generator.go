package generator

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/dave/jennifer/jen"
	"github.com/fitan/gowrap/pkg"
	"github.com/fitan/gowrap/printer"
	"github.com/fitan/gowrap/xtype"
	"github.com/pkg/errors"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

// Generator generates decorators for the interface types
type Generator struct {
	Options

	headerTemplate *template.Template
	bodyTemplate   *template.Template
	srcPackage     *packages.Package
	dstPackage     *packages.Package
	genTemplates   []genTemplate
	methods        methodsList
	//GenFn          *GenFn
	Enum          *Enum
	doc           *ast.CommentGroup
	interfaceType string
	localPrefix   string
}

type genTemplate struct {
	bodyTemplate *template.Template
	dstPackage   *packages.Package
}

// TemplateInputs information passed to template for generation
type TemplateInputs struct {
	// Interface information for template
	Interface TemplateInputInterface
	// Vars additional vars to pass to the template, see Options.Vars
	Vars    map[string]interface{}
	Imports []string

	//GenFn *GenFn
	Enum *Enum
}

func (t TemplateInputInterface) hashToID(s string) int64 {
	hash := sha256.Sum256([]byte(s))
	return int64(binary.BigEndian.Uint64(hash[:8]) % uint64(int64(math.Pow(10, 15))))
}

func (t TemplateInputInterface) CEPermissionSql() string {
	var (
		parentID          int64
		parentIcon        string
		parentMenu        int
		parentMethod      string
		parentAlias       string
		parentName        string
		parentPath        string
		parentDescription string
		sqls              []string
	)

	//INSERT INTO spider_dev.sys_permission (id, parent_id, icon, menu, method, alias, name, path, description, created_at, updated_at, deleted_at) VALUES (878, 877, '', 1, 'GET', 'Redis实例', 'menu.cdb.redis', '/cdb/index', '', '2022-12-29 14:21:22', '2022-12-29 14:21:22', null);

	parentMenu = 1
	parentMethod = "GET"
	parentAlias = t.Annotation()
	parentName = strings.ToLower(strings.Join([]string{"menu", strings.Trim(strings.Replace(t.BasePath(), "/", ".", -1), "."), t.Name}, "."))
	parentPath = t.BasePath() + "/index"
	parentDescription = t.Annotation()
	parentID = t.hashToID(parentName)
	parentSql := fmt.Sprintf(`INSERT INTO sys_permission (id, parent_id, icon, menu, method, alias, name, path, description,created_at, updated_at) VALUES (%v, %v, '%v', %v, '%v', '%v', '%v', '%v', '%v', '%v', '%v');`,
		parentID, 0, parentIcon, parentMenu, parentMethod, strings.TrimSpace(parentAlias), parentName, parentPath, strings.TrimSpace(parentDescription), time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))
	sqls = append(sqls, parentSql)

	for _, v := range t.Methods {
		var (
			id          int64
			icon        string
			menu        int
			method      string
			alias       string
			name        string
			mPath       string
			description string
		)

		mPath = t.MethodPath(v.Name)
		menu = 0
		method = strings.ToUpper(v.RawKit.Conf.UrlMethod)
		if method == "" {
			continue
		}
		alias = v.Annotation()
		name = strings.Trim(strings.ToLower(strings.Join([]string{strings.Replace(t.BasePath(), "/", ".", -1), v.Name, method}, ".")), ".")
		description = v.Annotation()
		id = t.hashToID(name)
		sql := fmt.Sprintf(`INSERT INTO sys_permission (id, parent_id, icon, menu, method, alias, name, path, description,created_at, updated_at) VALUES (%v, %v, '%v', %v, '%v', '%v', '%v', '%v', '%v', '%v', '%v');`,
			id, parentID, icon, menu, method, strings.TrimSpace(alias), name, mPath, strings.TrimSpace(description), time.Now().Format("2006-01-02 15:04:05"), time.Now().Format("2006-01-02 15:04:05"))

		sqls = append(sqls, sql)
	}

	return strings.Join(sqls, "\n")
}

// Import generates an import statement using a list of imports from the source file
// along with the ones from the template itself
func (t TemplateInputs) Import(imports ...string) string {
	allImports := make(map[string]struct{}, len(imports)+len(t.Imports))

	for _, i := range t.Imports {
		allImports[strings.TrimSpace(i)] = struct{}{}
	}

	for _, i := range imports {
		if len(i) == 0 {
			continue
		}

		i = strings.TrimSpace(i)

		//if i[len(i)-1] != '"' {
		//	i += `"`
		//}
		//
		//if i[0] != '"' {
		//	i = `"` + i
		//}

		allImports[i] = struct{}{}
	}

	out := make([]string, 0, len(allImports))

	for i := range allImports {
		out = append(out, i)
	}

	sort.Strings(out)

	return "import (\n" + strings.Join(out, "\n") + ")\n"
}

type ParseDoc ast.CommentGroup

func (p *ParseDoc) ValidVersion() string {
	if p != nil {
		for _, v := range p.List {
			fields := strings.Fields(v.Text)
			if strings.HasPrefix(strings.Join(fields, " "), "// @validVersion") {
				field := strings.Fields(v.Text)
				if len(field) >= 2 {
					return field[2]
				}
			}
		}
	}
	return ""
}

// TemplateInputInterface subset of interface information used for template generation
type TemplateInputInterface struct {
	Name string
	// Type of the interface, with package name qualifier (e.g. sort.Interface)
	Type string
	// Methods name keyed map of method information
	Methods map[string]Method
	Doc     *ast.CommentGroup

	Type2Ast *Type2ast
}

func (t TemplateInputInterface) KitServerOption() (res string) {
	tagStr := "// @kit-server-option"
	for _, v := range t.Doc.List {
		docFormat := DocFormat(v.Text)
		if strings.HasPrefix(docFormat, tagStr) {
			return strings.Join(strings.Fields(strings.TrimPrefix(docFormat, tagStr)), ",")
		}
	}
	return
}

func (t TemplateInputInterface) Annotation() string {
	if t.Doc == nil {
		return ""
	}
	for _, c := range t.Doc.List {
		docFormat := DocFormat(c.Text)
		if strings.HasPrefix(docFormat, "// "+t.Name) {
			return strings.TrimPrefix(docFormat, "// "+t.Name)
		}
	}
	return strings.TrimPrefix(DocFormat(t.Doc.List[0].Text), "// ")
}

func (t TemplateInputInterface) Tags() string {
	if t.Doc != nil {
		for _, v := range t.Doc.List {
			fields := strings.Fields(v.Text)
			if strings.HasPrefix(strings.Join(fields, " "), "// @tags") {
				return v.Text
			}
		}
	}
	return "// @tags " + t.Name
}

func (t TemplateInputInterface) ValidVersion() string {
	if t.Doc != nil {
		for _, v := range t.Doc.List {
			fields := strings.Fields(v.Text)
			if strings.HasPrefix(strings.Join(fields, " "), "// @validVersion") {
				if len(fields) >= 3 {
					return fields[2]
				}
			}
		}
	}
	return ""
}

func (t TemplateInputInterface) BasePath() string {
	if t.Doc != nil {
		for _, v := range t.Doc.List {
			if strings.HasPrefix(DocFormat(v.Text), "// @basePath") {
				return strings.TrimPrefix(DocFormat(v.Text), "// @basePath ")
			}

		}
	}
	return ""
}

func (t TemplateInputInterface) EnableSwag(name string) bool {
	if t.Doc != nil {
		for _, v := range t.Doc.List {
			if strings.HasPrefix(DocFormat(v.Text), "// @Swag false") {
				return false
			}

		}
	}

	return t.Methods[name].EnableSwag()
}

func (t TemplateInputInterface) HasMethodPath(name string) bool {
	return t.Methods[name].RawKit.Conf.Url != ""
}

func (t TemplateInputInterface) ClientTypes() string {
	code := jen.Null()
	for _, v := range t.Methods {
		if t.HasMethodPath(v.Name) {
			code.Line().Id(t.Type2Ast.Parse(xtype.TypeOf(v.KitRequest.RequestTypeOf), v.Name+"ClientReq"))
		}
	}
	fmt.Println("last Type2ast", t.Type2Ast.current)
	return code.GoString()
}

func (t TemplateInputInterface) ClientInterface() string {
	return jen.Comment(t.Doc.Text()).Line().Type().Id("HttpClientI").InterfaceFunc(func(group *jen.Group) {
		for _, v := range t.Methods {
			if t.HasMethodPath(v.Name) {
				group.Comment(strings.Join(v.Doc, "\n")).Line().Add(v.ClientInterfaceFunc())
			}
		}
	}).GoString()
}

func (t TemplateInputInterface) MethodPath(name string) string {
	return strings.TrimSuffix(path.Join(t.BasePath(), t.Methods[name].RawKit.Conf.Url), "/")
}

type methodsList map[string]Method

// Options of the NewGenerator constructor
type Options struct {
	//InterfaceName is a name of interface type
	InterfaceName string

	InterfaceDoc *ast.CommentGroup

	//Imports from the file with interface definition
	Imports []string

	SourceLoadPkg *packages.Package

	TokenFileSet *token.FileSet

	AstPackage *ast.Package

	TemplateInputs *TemplateInputs

	//SourcePackage is an import path or a relative path of the package that contains the source interface
	SourcePackage string

	//SourcePackageAlias is an import selector defauls is source package name
	SourcePackageAlias string

	//OutputFile name which is used to detect destination package name and also to fix imports in the resulting source
	OutputFile string

	//HeaderTemplate is used to generate package clause and comment over the generated source
	HeaderTemplate string

	//BodyTemplate generates import section, decorator constructor and methods
	BodyTemplate string

	//Vars additional vars that are passed to the templates from the command line
	Vars map[string]interface{}

	//HeaderVars header specific variables
	HeaderVars map[string]interface{}

	//Funcs is a map of helper functions that can be used within a template
	Funcs template.FuncMap

	//LocalPrefix is a comma-separated string of import path prefixes, which, if set, instructs Process to sort the import
	//paths with the given prefixes into another group after 3rd-party packages.
	LocalPrefix string

	PkgNeedSyntax bool

	BatchTemplate []BatchTemplate

	RunCmdDir string
	InitType  string

	Type2ast *Type2ast
}

type BatchTemplate struct {
	OutputFile   string
	BodyTemplate string
}

var errEmptyInterface = errors.New("interface has no methods")
var errUnexportedMethod = errors.New("unexported method")

var methods methodsList
var importSpecs []*ast.ImportSpec
var doc *ast.CommentGroup

// var srcPackage *packages.Package
// var dstPackage *packages.Package
var srcPackageAST *ast.Package

var globalOption Options

func NewGeneratorInit(ops []Options) ([]*Generator, error) {
	if len(ops) == 0 {
		return nil, nil
	}

	globalOption = ops[0]

	gs := make([]*Generator, 0, 0)

	for _, options := range ops {
		headerTemplate, err := template.New("header").Funcs(options.Funcs).Parse("")
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse header template")
		}

		bodyTemplate, err := template.New("body").Funcs(options.Funcs).Parse(options.BodyTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse body template")
		}

		if options.Vars == nil {
			options.Vars = make(map[string]interface{})
		}

		options.Vars["instance"] = makeInstance(globalOption.RunCmdDir)
		options.Vars["initType"] = options.InitType
		if options.SourceLoadPkg != nil {
			options.Vars["pkgName"] = options.SourceLoadPkg.Name
		}

		dstPackagePath := filepath.Dir(options.OutputFile)
		if !strings.HasPrefix(dstPackagePath, "/") && !strings.HasPrefix(dstPackagePath, "./") {
			dstPackagePath = "./" + dstPackagePath
		}
		gs = append(gs, &Generator{
			Options:        options,
			headerTemplate: headerTemplate,
			bodyTemplate:   bodyTemplate,
			srcPackage:     nil,
			dstPackage:     nil,
			interfaceType:  "",
			methods:        methods,
			localPrefix:    options.LocalPrefix,
			genTemplates:   make([]genTemplate, 0, 0),
		})
	}

	return gs, nil
}

//func NewGeneratorFn(ops []Options) ([]*Generator, error) {
//	if len(ops) == 0 {
//		return nil, nil
//	}
//
//	globalOption = ops[0]
//
//	gs := make([]*Generator, 0, 0)
//	for _, options := range ops {
//		if options.Funcs == nil {
//			options.Funcs = make(template.FuncMap)
//		}
//
//		headerTemplate, err := template.New("header").Funcs(options.Funcs).Parse(options.HeaderTemplate)
//		if err != nil {
//			return nil, errors.Wrap(err, "failed to parse header template")
//		}
//
//		bodyTemplate, err := template.New("body").Funcs(options.Funcs).Parse(options.BodyTemplate)
//		if err != nil {
//			return nil, errors.Wrap(err, "failed to parse body template")
//		}
//
//		if options.Vars == nil {
//			options.Vars = make(map[string]interface{})
//		}
//
//		options.Vars["instance"] = makeInstance(globalOption.RunCmdDir)
//
//		fs := token.NewFileSet()
//
//		//if srcPackage == nil {
//		//	srcPackage, err = pkg.Load(options.SourcePackage, options.PkgNeedSyntax)
//		//	if err != nil {
//		//		return nil, errors.Wrap(err, "failed to load source package")
//		//	}
//		//}
//
//		dstPackagePath := filepath.Dir(options.OutputFile)
//		if !strings.HasPrefix(dstPackagePath, "/") && !strings.HasPrefix(dstPackagePath, "./") {
//			dstPackagePath = "./" + dstPackagePath
//		}
//
//		//if dstPackage == nil {
//		//	dstPackage, err = loadDestinationPackage(dstPackagePath)
//		//	if err != nil {
//		//		return nil, errors.Wrapf(err, "failed to load destination package: %s", dstPackagePath)
//		//	}
//		//}
//
//		if srcPackageAST == nil {
//			srcPackageAST, err = pkg.AST(fs, options.SourceLoadPkg)
//			if err != nil {
//				return nil, errors.Wrap(err, "failed to parse source package")
//			}
//		}
//
//		interfaceType := options.SourceLoadPkg.Name + "." + options.InterfaceName
//		if options.SourceLoadPkg.PkgPath == options.SourceLoadPkg.PkgPath {
//			interfaceType = options.InterfaceName
//			srcPackageAST.Name = ""
//		} else {
//			if options.SourcePackageAlias != "" {
//				srcPackageAST.Name = options.SourcePackageAlias
//			}
//
//			options.Imports = append(options.Imports, `"`+options.SourceLoadPkg.PkgPath+`"`)
//		}
//
//		for _, f := range srcPackageAST.Files {
//			importSpecs = append(importSpecs, f.Imports...)
//		}
//
//		options.Imports = append(options.Imports, makeImports(importSpecs)...)
//
//		jenF := jen.NewFile("genfn")
//
//		genFn := NewGenFn(options.SourceLoadPkg, jenF, NewGenFnCopy())
//		genFn.Parse()
//		genFn.Run()
//
//		gs = append(gs, &Generator{
//			Options:        options,
//			headerTemplate: headerTemplate,
//			bodyTemplate:   bodyTemplate,
//			srcPackage:     options.SourceLoadPkg,
//			dstPackage:     options.SourceLoadPkg,
//			interfaceType:  interfaceType,
//			methods:        methods,
//			GenFn:          genFn,
//			doc:            doc,
//			localPrefix:    options.LocalPrefix,
//			genTemplates:   make([]genTemplate, 0, 0),
//		})
//
//	}
//	return gs, nil
//}

// NewGenerator returns Generator initialized with options
func NewGenerator(ops []Options) ([]*Generator, error) {
	if len(ops) == 0 {
		return nil, nil
	}

	globalOption = ops[0]

	gs := make([]*Generator, 0, 0)
	for _, options := range ops {
		if options.Funcs == nil {
			options.Funcs = make(template.FuncMap)
		}

		headerTemplate, err := template.New("header").Funcs(options.Funcs).Parse(options.HeaderTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse header template")
		}

		bodyTemplate, err := template.New("body").Funcs(options.Funcs).Parse(options.BodyTemplate)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse body template")
		}

		if options.Vars == nil {
			options.Vars = make(map[string]interface{})
		}

		options.Vars["instance"] = makeInstance(globalOption.RunCmdDir)

		fs := token.NewFileSet()

		//if srcPackage == nil {
		//	srcPackage, err = pkg.Load(options.SourcePackage, options.PkgNeedSyntax)
		//	if err != nil {
		//		return nil, errors.Wrap(err, "failed to load source package")
		//	}
		//}

		dstPackagePath := filepath.Dir(options.OutputFile)
		if !strings.HasPrefix(dstPackagePath, "/") && !strings.HasPrefix(dstPackagePath, "./") {
			dstPackagePath = "./" + dstPackagePath
		}

		//if dstPackage == nil {
		//	dstPackage, err = loadDestinationPackage(dstPackagePath)
		//	if err != nil {
		//		return nil, errors.Wrapf(err, "failed to load destination package: %s", dstPackagePath)
		//	}
		//}

		if srcPackageAST == nil {
			srcPackageAST, err = pkg.AST(fs, options.SourceLoadPkg)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse source package")
			}
		}

		interfaceType := options.SourceLoadPkg.Name + "." + options.InterfaceName
		if options.SourceLoadPkg.PkgPath == options.SourceLoadPkg.PkgPath {
			interfaceType = options.InterfaceName
			srcPackageAST.Name = ""
		} else {
			if options.SourcePackageAlias != "" {
				srcPackageAST.Name = options.SourcePackageAlias
			}

			options.Imports = append(options.Imports, `"`+options.SourceLoadPkg.PkgPath+`"`)
		}

		if options.InterfaceName != "" {

			if methods == nil && importSpecs == nil {

				t1 := time.Now()
				methods, importSpecs, doc, err = findInterface(fs, options.SourceLoadPkg, srcPackageAST, options.InterfaceName, options)
				log.Printf("findInterface time: %v", time.Now().Sub(t1).String())
				if err != nil {
					return nil, errors.Wrap(err, "failed to parse interface declaration")
				}

				if len(methods) == 0 {
					return nil, errEmptyInterface
				}
			}

			for _, m := range methods {
				if srcPackageAST.Name != "" && []rune(m.Name)[0] == []rune(strings.ToLower(m.Name))[0] {
					return nil, errors.Wrap(errUnexportedMethod, m.Name)
				}
			}
		}

		mainImports, err := LoadMainImports()
		if err != nil {
			return nil, errors.Wrap(err, "failed to load main imports")
		}

		options.Imports = append(options.Imports, makeImports(importSpecs)...)
		options.Imports = append(options.Imports, makeImports(mainImports)...)

		gs = append(gs, &Generator{
			Options:        options,
			headerTemplate: headerTemplate,
			bodyTemplate:   bodyTemplate,
			srcPackage:     options.SourceLoadPkg,
			dstPackage:     options.SourceLoadPkg,
			interfaceType:  interfaceType,
			methods:        methods,
			Enum:           NewEnumGen(options.SourceLoadPkg),
			doc:            doc,
			localPrefix:    options.LocalPrefix,
			genTemplates:   make([]genTemplate, 0, 0),
		})

	}
	return gs, nil
}

func makeInstance(dirPath string) string {
	var dirS []string

	if runtime.GOOS == "windows" {
		dirS = strings.Split(dirPath, "\\")
	} else {
		dirS = strings.Split(dirPath, "/")
	}
	var instance string
	if len(dirS) >= 2 {
		instance = strings.Join(dirS[len(dirS)-2:], ".")
	}
	return instance
}

func makeImports(imports []*ast.ImportSpec) []string {
	result := make([]string, len(imports))
	for _, i := range imports {
		var name string
		if i.Name != nil {
			name = i.Name.Name
		}
		result = append(result, name+" "+i.Path.Value)
		extra := makeExtraImport(i.Doc)
		result = append(result, extra...)
	}

	return result
}

func LoadMainImports() (res []*ast.ImportSpec, err error) {
	mainName := "main.go"
	readMain := func(path string) (res []*ast.ImportSpec, err error) {
		fset := token.NewFileSet()
		var f *ast.File
		f, err = parser.ParseFile(fset, path, nil, parser.ParseComments|parser.ImportsOnly)
		if err != nil {
			return
		}

		res = f.Imports
		return
	}

	workPath, err := os.Getwd()
	if err != nil {
		return
	}

	for {
		_, err = os.Stat(filepath.Join(workPath, mainName))
		if err == nil {
			return readMain(filepath.Join(workPath, mainName))
		}

		if err != nil {
			if !os.IsNotExist(err) {
				return
			}
		}

		_, err = os.Stat(filepath.Join(workPath, "cmd", mainName))
		if err == nil {
			return readMain(filepath.Join(workPath, "cmd", mainName))
		}

		if err != nil {
			if !os.IsNotExist(err) {
				return
			}
		}

		preWorkPath := filepath.Dir(workPath)
		if preWorkPath == workPath {
			err = errors.New("not found main.go")
			return
		}

		workPath = preWorkPath

	}
}

func makeExtraImport(doc *ast.CommentGroup) (res []string) {
	if doc == nil {
		return
	}
	for _, c := range doc.List {
		if strings.HasPrefix(c.Text, "// @extra ") {
			res = append(res, strings.TrimSpace(strings.TrimPrefix(c.Text, "// @extra ")))
		}
	}
	return
}

func loadDestinationPackage(path string) (*packages.Package, error) {
	dstPackage, err := pkg.Load(path, false)
	if err != nil {
		//using directory name as a package name
		dstPackage, err = makePackage(path)
	}

	return dstPackage, err
}

var errNoPackageName = errors.New("failed to determine the destination package name")

func makePackage(path string) (*packages.Package, error) {
	name := filepath.Base(path)
	if name == string(filepath.Separator) || name == "." {
		return nil, errNoPackageName
	}

	return &packages.Package{
		Name: name,
	}, nil
}

// Generate generates code using header and body templates
func (g Generator) Generate(fix bool) error {
	buf := bytes.NewBuffer([]byte{})

	err := g.headerTemplate.Execute(buf, map[string]interface{}{
		"SourcePackage": g.srcPackage,
		"Package":       g.dstPackage,
		"Vars":          g.Options.Vars,
		"Options":       g.Options,
	})
	if err != nil {
		return err
	}

	err = g.bodyTemplate.Execute(buf, TemplateInputs{
		Interface: TemplateInputInterface{
			Name:    g.Options.InterfaceName,
			Type:    g.interfaceType,
			Methods: g.methods,
			Doc:     g.doc,

			Type2Ast: g.Options.Type2ast,
		},
		Imports: g.Options.Imports,
		Vars:    g.Options.Vars,

		//GenFn: g.GenFn,
		Enum: g.Enum,
	})
	if err != nil {
		return err
	}

	t1 := time.Now()
	imports.LocalPrefix = g.localPrefix
	var importsOpt *imports.Options
	if fix {
		importsOpt = &imports.Options{
			Comments:   true,
			TabWidth:   8,
			TabIndent:  true,
			FormatOnly: false,
		}
	} else {
		importsOpt = &imports.Options{
			Comments:   true,
			TabWidth:   8,
			TabIndent:  true,
			FormatOnly: true,
		}
	}
	processedSource, err := imports.Process(g.Options.OutputFile, buf.Bytes(), importsOpt)
	log.Printf("outoutFile: %v. imports.Porcess time: %v", g.Options.OutputFile, time.Now().Sub(t1).String())
	if err != nil {
		return errors.Wrapf(err, "failed to format generated code:\n%s", buf)
	}

	buf = bytes.NewBuffer([]byte{})
	_, err = buf.Write(processedSource)
	if err != nil {
		return nil
	}
	err = ioutil.WriteFile(g.Options.OutputFile, buf.Bytes(), 0664)
	return err

	//_, err = w.Write(processedSource)
	//return err
}

var errInterfaceNotFound = errors.New("interface type declaration not found")

// findInterface looks for the interface declaration in the given directory
// and returns a list of the interface's methods and a list of imports from the file
// where interface type declaration was found
func findInterface(fs *token.FileSet, currentPackage *packages.Package, p *ast.Package, interfaceName string, o Options) (methods methodsList, imports []*ast.ImportSpec, doc *ast.CommentGroup, err error) {
	var found bool
	var types []*ast.TypeSpec
	var it *ast.InterfaceType
	var currentFile *ast.File

	//looking for the source interface declaration in all files in the dir
	//while doing this we also store all found type declarations to check if some of the
	//interface methods use unexported types
	for _, f := range currentPackage.Syntax {
		tsList, gdDoc := typeSpecs(f)
		for index, ts := range tsList {
			types = append(types, ts)

			if i, ok := ts.Type.(*ast.InterfaceType); ok {
				if ts.Name.Name == interfaceName && !found {
					doc = gdDoc[index]
					imports = f.Imports
					currentFile = f
					it = i
					found = true
				}
			}
		}
	}

	o.InterfaceDoc = doc

	if !found {
		return nil, nil, doc, errors.Wrap(errInterfaceNotFound, interfaceName)
	}

	methods, err = processInterface(interfaceName, fs, currentPackage, currentFile, it, types, p.Name, imports, o)
	if err != nil {
		return nil, nil, doc, err
	}

	return methods, imports, doc, err
}

func typeSpecs(f *ast.File) ([]*ast.TypeSpec, []*ast.CommentGroup) {
	result := []*ast.TypeSpec{}
	gdDoc := []*ast.CommentGroup{}

	for _, decl := range f.Decls {
		if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.TYPE {
			for _, spec := range gd.Specs {
				if ts, ok := spec.(*ast.TypeSpec); ok {
					result = append(result, ts)
					gdDoc = append(gdDoc, gd.Doc)
				}
			}
		}
	}

	return result, gdDoc
}

func processInterface(interfaceName string, fs *token.FileSet, currentPackage *packages.Package, currentFile *ast.File, it *ast.InterfaceType, typeSpecs []*ast.TypeSpec, typesPrefix string, imports []*ast.ImportSpec, o Options) (methods methodsList, err error) {
	if it.Methods == nil {
		return nil, nil
	}

	//interfaceType := currentPackage.Types.Scope().Lookup(interfaceName).Type().Underlying().(*types.Interface)

	methods = make(methodsList, len(it.Methods.List))

	for _, field := range it.Methods.List {
		var embeddedMethods methodsList

		var kit Kit
		switch v := field.Type.(type) {
		case *ast.FuncType:

			if field.Doc != nil {
				kit, err = NewKit(interfaceName, field.Names[0].Name, currentPackage, field)
				if err != nil {
					err = errors.Wrap(err, "NewKit")
					return
				}
			}

			var method *Method
			method, err = NewMethod(currentPackage, currentFile, field.Names[0].Name, field, printer.New(fs, typeSpecs, typesPrefix), o)

			if err == nil {

				if globalOption.PkgNeedSyntax {
					httpMethod := NewHttpMethod(field.Names[0].Name, currentPackage, currentFile, field)
					has, err := httpMethod.Parse()
					if err != nil {
						return nil, err
					}

					if has {
						method.Gin = httpMethod.GinParams
						method.Kit = httpMethod.KitParams
						method.HasGin = true
					}
				}
				method.RawKit = kit
				if kit.Conf.HttpRequestName != "" {
					kitRequest := NewKitRequest(
						currentPackage, field.Names[0].Name, method.RawKit.Conf.HttpRequestName,
						method.RawKit.Conf.HttpRequestBody, o,
					)
					if kit.Conf.HttpRequestName != "nil" {
						kitRequest.ParseRequest()
					}
					method.KitRequest = kitRequest
					method.KitRequestDecode = kitRequest.DecodeRequest()
				}

				//if kit.Conf.HttpResponseName != "" {
				//	response := NewResponse(currentPackage, interfaceType.Method(index), kit.Conf.HttpResponseName)
				//	method.KitResponse = response
				//}

				methods[field.Names[0].Name] = *method
			}
		case *ast.SelectorExpr:
			embeddedMethods, err = processSelector(fs, currentPackage, v, imports, o)
		case *ast.Ident:
			embeddedMethods, err = processIdent(fs, currentPackage, currentFile, v, typeSpecs, typesPrefix, imports, o)
		}

		if err != nil {
			return nil, err
		}

		methods, err = mergeMethods(methods, embeddedMethods)
		if err != nil {
			return nil, err
		}
	}

	return methods, nil
}

func processSelector(fs *token.FileSet, currentPackage *packages.Package, se *ast.SelectorExpr, imports []*ast.ImportSpec, o Options) (methodsList, error) {
	interfaceName := se.Sel.Name
	packageSelector := se.X.(*ast.Ident).Name

	importPath, err := findImportPathForName(packageSelector, imports, currentPackage)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to find package %s", packageSelector)
	}

	p, ok := currentPackage.Imports[importPath]
	if !ok {
		return nil, fmt.Errorf("unable to find package %s", packageSelector)
	}

	astPkg, err := pkg.AST(fs, p)
	if err != nil {
		return nil, errors.Wrap(err, "failed to import package")
	}

	methods, _, _, err = findInterface(fs, p, astPkg, interfaceName, o)

	return methods, err
}

var errDuplicateMethod = errors.New("embedded interface has same method")

// mergeMethods merges two methods list, if there is a duplicate method name
// errDuplicateMethod is returned
func mergeMethods(ml1, ml2 methodsList) (methodsList, error) {
	if ml1 == nil || ml2 == nil {
		return ml1, nil
	}

	result := make(methodsList, len(ml1)+len(ml2))
	for k, v := range ml1 {
		result[k] = v
	}

	for name, signature := range ml2 {
		if _, ok := ml1[name]; ok {
			return nil, errors.Wrap(errDuplicateMethod, name)
		}

		result[name] = signature
	}

	return result, nil
}

var errEmbeddedInterfaceNotFound = errors.New("embedded interface not found")
var errNotAnInterface = errors.New("embedded type is not an interface")

func processIdent(fs *token.FileSet, currentPackage *packages.Package, currentFile *ast.File, i *ast.Ident, types []*ast.TypeSpec, typesPrefix string, imports []*ast.ImportSpec, o Options) (methodsList, error) {
	var embeddedInterface *ast.InterfaceType
	var interfaceName string
	for _, t := range types {
		if t.Name.Name == i.Name {
			var ok bool
			embeddedInterface, ok = t.Type.(*ast.InterfaceType)
			interfaceName = t.Name.Name
			if !ok {
				return nil, errors.Wrap(errNotAnInterface, t.Name.Name)
			}
			break
		}
	}

	if embeddedInterface == nil {
		return nil, errors.Wrap(errEmbeddedInterfaceNotFound, i.Name)
	}

	return processInterface(interfaceName, fs, currentPackage, currentFile, embeddedInterface, types, typesPrefix, imports, o)
}

var errUnknownSelector = errors.New("unknown selector")

func findImportPathForName(name string, imports []*ast.ImportSpec, currentPackage *packages.Package) (string, error) {
	for _, i := range imports {
		if i.Name != nil && i.Name.Name == name {
			return unquote(i.Path.Value), nil
		}
	}

	for path, pkg := range currentPackage.Imports {
		if pkg.Name == name {
			return path, nil
		}
	}

	return "", errors.Wrapf(errUnknownSelector, name)
}

func unquote(s string) string {
	if s[0] == '"' {
		s = s[1:]
	}

	if s[len(s)-1] == '"' {
		s = s[0 : len(s)-1]
	}

	return s
}
