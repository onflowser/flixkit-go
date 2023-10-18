package bindings

import (
	"bytes"
	"os"
	"sort"
	"text/template"

	"github.com/onflow/flixkit-go"
	"github.com/stoewer/go-strcase"
)

type SimpleParameter struct {
	Name        string
	JsType      string
	Description string
	FclType     string
	CadType     string
}

type TemplateData struct {
	Version         string
	Parameters      []SimpleParameter
	Title           string
	Description     string
	Location        string
	IsScript        bool
	IsLocalTemplate bool
}

type FclJSGenerator struct {
	TemplateDir string
}

func (g FclJSGenerator) Generate(flix *flixkit.FlowInteractionTemplate, templateLocation string, isLocal bool) (string, error) {
	files, err := os.ReadDir(g.TemplateDir)
	if err != nil {
		return "", err
	}
	templateFiles := make([]string, 0)
	for _, file := range files {
		if !file.IsDir() {
			templateFiles = append(templateFiles, g.TemplateDir+"/"+file.Name())
		}
	}
	tmpl, err := template.ParseFiles(templateFiles...)
	if err != nil {
		return "", err
	}

	methodName := strcase.LowerCamelCase(flix.Data.Messages.GetTitleValue("Request"))
	description := flix.GetDescription()
	data := TemplateData{
		Version:         flix.FVersion,
		Parameters:      transformArguments(flix.Data.Arguments),
		Title:           methodName,
		Description:     description,
		Location:        templateLocation,
		IsScript:        flix.IsScript(),
		IsLocalTemplate: isLocal,
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	return buf.String(), err
}

func transformArguments(args flixkit.Arguments) []SimpleParameter {
	simpleArgs := []SimpleParameter{}
	var keys []string
	// get keys for sorting
	for k := range args {
		keys = append(keys, k)
	}

	sort.SliceStable(keys, func(i, j int) bool {
		return args[keys[i]].Index < args[keys[j]].Index
	})
	for _, key := range keys {
		arg := args[key]
		isArray, cType, jsType := isArrayParameter(arg)
		desciption := arg.Messages.GetTitleValue("")
		if isArray {
			simpleArgs = append(simpleArgs, SimpleParameter{Name: key, CadType: cType, JsType: jsType, FclType: "Array(t." + cType + ")", Description: desciption})
		} else {
			jsType := convertCadenceTypeToJS(arg.Type)
			simpleArgs = append(simpleArgs, SimpleParameter{Name: key, CadType: arg.Type, JsType: jsType, FclType: arg.Type, Description: desciption})
		}
	}
	return simpleArgs
}

func isArrayParameter(arg flixkit.Argument) (bool, string, string) {
	if arg.Type == "" || arg.Type[0] != '[' {
		return false, "", ""
	}
	cadenceType := arg.Type[1 : len(arg.Type)-1]
	jsType := "Array<" + convertCadenceTypeToJS(cadenceType) + ">"
	return true, cadenceType, jsType
}

func convertCadenceTypeToJS(cadenceType string) string {
	// need to determine js type based on fcl supported types
	// looking at fcl types and how arguments work as parameters
	// https://github.com/onflow/fcl-js/blob/master/packages/types/src/types.js
	switch cadenceType {
	case "Bool":
		return "boolean"
	case "Void":
		return "void" // return type only
	case "Dictionary":
		return "object" // TODO: support Collection type, test to see what fcl
	case "Struct":
		return "object" // TODO: support Composite type, test to see what fcl
	case "Enum":
		return "object" // TODO: support Composite type, test to see what fcl
	default:
		return "string"
	}
}
