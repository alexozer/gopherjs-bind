package jsbind

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/robertkrimen/otto"
)

type Binding struct {
	name  string
	elems []Element
}

const (
	packageStmt = "package"
	importStmt  = "import \"github.com/gopherjs/gopherjs/js\""
)

func New(name string) *Binding {
	return &Binding{name, make([]Element, 0)}
}

func (b *Binding) addElement(e Element) {
	b.elems = append(b.elems, e)
}

func (b *Binding) Export(filename string) error {
	err := os.Remove(filename)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprintln(file, packageStmt, b.name)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(file, importStmt)
	if err != nil {
		return err
	}

	for _, elem := range b.elems {
		for _, line := range elem.Text() {
			_, err = file.WriteString(line + "\n")
			if err != nil {
				return err
			}
		}
		file.WriteString("\n")
	}

	return nil
}

type Element interface {
	Text() []string
}

type Interface struct {
	Name    string
	Methods []*Method
}

func (i *Interface) Text() (lines []string) {
	lines = make([]string, 0, 1)

	lines = append(lines, fmt.Sprintf(
		"type %s interface {",
		i.Name,
	))

	for _, method := range i.Methods {
		lines = append(lines, fmt.Sprintf(
			"%s(%s) %s",
			method.Name,
			method.Params.sanitized().listTypes(),
			method.Return.Type,
		))
	}

	lines = append(lines, "}")

	return
}

type Method struct {
	Name    string
	Binding *Struct
	Params  VarList
	Return  *Variable
}

func (m *Method) Text() (lines []string) {
	if capitialize(m.Name) == m.Name {
		return m.genConstructor()
	}

	var binding string
	if m.Binding != nil {
		binding = fmt.Sprintf("(self *%s)", m.Binding.Name)
	}

	cleanParams := m.Params.sanitized()

	lines = []string{
		fmt.Sprintf(
			"func %s %s(%s) %s {",
			binding,
			capitialize(m.Name),
			cleanParams.list(),
			"interface{}",
		),
	}

	jsInvocation := fmt.Sprintf(
		"self.Call(\"%s\", %s)",
		m.Name,
		cleanParams.listNames(),
	)

	if m.Return == nil {
		lines = append(lines, jsInvocation, "}")
		return
	}

	castedObj, err := m.genCastObject(jsInvocation, m.Return.Type)
	if err != nil {
		panic(err)
	}

	lines = append(lines, "return "+castedObj, "}")
	return
}

func (m *Method) genConstructor() (lines []string) {
	cleanParams := m.Params.sanitized()

	return []string{
		fmt.Sprintf(
			"func New%s(%s) *%s {",
			m.Name,
			cleanParams.list(),
			m.Name,
		),
		fmt.Sprintf(
			"return js.Global.Get(\"%s\").New(%s)",
			m.Name,
			cleanParams.listNames(),
		),
		"}",
	}
}

func (m *Method) genCastObject(obj, goType string) (string, error) {
	switch goType {
	case "float64":
		return obj + ".Float()", nil
	case "int":
		return obj + ".Int()", nil
	case "string":
		return obj + ".Str()", nil
	case "interface{}":
		return obj + ".Interface()", nil
	default:
		return "", errors.New("Function may only return float64, int, string, or interface{}")
	}
}

type Struct struct {
	Name    string
	Methods []*Method
	Fields  []*Variable
}

func (s *Struct) Text() (lines []string) {
	lines = []string{
		fmt.Sprintf(
			"type %s struct {",
			s.Name,
		),
		"js.Object",
	}

	for _, field := range s.Fields {
		lines = append(lines, fmt.Sprintf(
			"%s %s `js:\"%s\"`",
			field.capitialized().Name,
			field.Type,
			field.Name,
		))
	}

	lines = append(lines, "}")

	return
}

func (b *Binding) Add(name string, obj *otto.Object) error {
	objStruct := Struct{
		capitialize(name),
		make([]*Method, 0),
		make([]*Variable, 0),
	}

	for _, key := range obj.Keys() {
		val, err := obj.Get(key)
		if err != nil {
			return err
		}

		switch {
		case val.IsFunction():
			if key == capitialize(key) {
				// Function should be a constructor
				b.addElement(&Method{
					Name:    key,
					Binding: nil,
					Params:  parseParams(&val),
					Return:  nil, // automatically calculated
				})

				proto, err := val.Object().Get("prototype")
				if err != nil {
					return err
				}

				b.Add(key, proto.Object())
			} else {
				// Function should be a method
				if key == "constructor" {
					continue // Constructors are generated automatically
				}

				b.addElement(&Method{
					Name:    key,
					Binding: &objStruct,
					Params:  parseParams(&val),
					Return:  &Variable{Type: "interface{}"},
				})
			}

		case val.IsPrimitive():
			var field = &Variable{Name: key}

			field.setType(&val)

			objStruct.Fields = append(objStruct.Fields, field)

		case val.Class() == "Array":
			objStruct.Fields = append(objStruct.Fields, &Variable{
				Name: key,
				Type: "[]interface{}",
			})

		case val.IsObject():
			b.Add(key, val.Object())

		default:
			objStruct.Fields = append(objStruct.Fields, &Variable{key, "interface{}"})
		}
	}

	b.addElement(&objStruct)
	return nil
}

const (
	commentPattern = `((?ms:\/\/.*$)|(?ms:\/\*.*\*\/))`
	parensPattern  = `function\s*\(((?s:[^\)]*))\)`
	splitPattern   = `\s*,\s*`
)

var (
	commentRegexp = regexp.MustCompile(commentPattern)
	parensRegexp  = regexp.MustCompile(parensPattern)
	splitRegexp   = regexp.MustCompile(splitPattern)
)

func parseParams(function *otto.Value) VarList {
	header := function.String()

	header = commentRegexp.ReplaceAllString(header, "")

	header = parensRegexp.FindStringSubmatch(header)[1]

	header = strings.TrimSpace(header)
	if header == "" {
		return make(VarList, 0)
	}

	paramNames := splitRegexp.Split(header, -1)

	params := make(VarList, len(paramNames))
	for i, param := range splitRegexp.Split(header, -1) {
		params[i] = Variable{Name: param, Type: "interface{}"}
	}

	return params
}

func capitialize(str string) string {
	if len(str) == 0 {
		return str
	}

	return strings.ToUpper(str[:1]) + str[1:]
}
