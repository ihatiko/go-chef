package utils

import (
	"bytes"
	"os"
	"strings"
	"text/template"
)

const (
	functionTemplate string = "function.tmpl"
)

type FunctionSignature struct {
	Params      []string
	Return      string
	Name        string
	InheritType string
	Content     []string
	ReturnType  string
}
type InternalFunctionSignature struct {
	Params      string
	Return      string
	Name        string
	InheritType string
	Content     string
	ReturnType  string
}

func GetFunction(signature FunctionSignature) string {
	internalSignature := InternalFunctionSignature{
		Content:     strings.Join(signature.Content, "\n"),
		Return:      signature.Return,
		Params:      strings.Join(signature.Params, ","),
		ReturnType:  signature.ReturnType,
		Name:        signature.Name,
		InheritType: signature.InheritType,
	}
	tmpl, err := os.ReadFile("utils/" + functionTemplate)
	if err != nil {
		panic(err)
	}
	t, err := template.New("").Parse(string(tmpl))
	if err != nil {
		panic(err)
	}
	buf := bytes.NewBufferString("")
	err = t.ExecuteTemplate(buf, "", internalSignature)
	if err != nil {
		panic(err)
	}
	return buf.String()
}
