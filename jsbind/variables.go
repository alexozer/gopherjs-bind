package jsbind

import "github.com/robertkrimen/otto"

type Variable struct {
	Name, Type string
}

func (v *Variable) capitialized() *Variable {
	return &Variable{
		Name: capitialize(v.Name),
		Type: v.Type,
	}
}

func (v *Variable) sanitized() *Variable {
	name := v.Name
	if reservedKeywords[name] {
		name += "_"
	}

	return &Variable{name, v.Type}
}

func (v *Variable) setType(val *otto.Value) {
	switch {
	case val.IsString():
		v.Type = "string"
	case val.IsBoolean():
		v.Type = "bool"
	case val.IsNumber():
		v.Type = "float64"
	default:
		v.Type = "interface{}"
	}
}

type VarList []Variable

func (vl VarList) sanitized() VarList {
	cleanList := make(VarList, len(vl))
	for i, val := range vl {
		cleanList[i] = *val.sanitized()
	}

	return cleanList
}

var reservedKeywords = map[string]bool{
	"break":       true,
	"default":     true,
	"func":        true,
	"interface":   true,
	"select":      true,
	"case":        true,
	"defer":       true,
	"go":          true,
	"map":         true,
	"struct":      true,
	"chan":        true,
	"else":        true,
	"goto":        true,
	"package":     true,
	"switch":      true,
	"const":       true,
	"fallthrough": true,
	"if":          true,
	"range":       true,
	"type":        true,
	"continue":    true,
	"for":         true,
	"import":      true,
	"return":      true,
	"var":         true,
}

func (vl VarList) list() string {
	if len(vl) == 0 {
		return ""
	}

	var listStr string

	for i := 0; i < len(vl)-1; i++ {
		listStr += vl[i].Name + " " + vl[i].Type + ", "
	}

	lastVar := vl[len(vl)-1]
	listStr += lastVar.Name + " " + lastVar.Type

	return listStr
}

func (vl VarList) listNames() string {
	if len(vl) == 0 {
		return ""
	}

	var listStr string

	for i := 0; i < len(vl)-1; i++ {
		listStr += vl[i].Name + ", "
	}

	listStr += vl[len(vl)-1].Name

	return listStr
}

func (vl VarList) listTypes() string {
	if len(vl) == 0 {
		return ""
	}

	var listStr string

	for i := 0; i < len(vl)-1; i++ {
		listStr += vl[i].Type + ", "
	}

	listStr += vl[len(vl)-1].Type

	return listStr
}
