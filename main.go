package main

import (
	"flag"
	"log"

	"github.com/alexozer/gopherjs-bind/jsbind"
)

const (
	inputDesc = "Javascript file to parse"
	nameDesc  = "Name of resulting binding's package"
	objDesc   = "Javascript object which contains Javascript library"
)

var (
	inputFlag = flag.String("input", "", inputDesc)
	nameFlag  = flag.String("name", "", nameDesc)
	objFlag   = flag.String("object", "", objDesc)
)

func init() {
	flag.StringVar(inputFlag, "i", "", inputDesc)
	flag.StringVar(nameFlag, "n", "", nameDesc)
	flag.StringVar(objFlag, "O", "", objDesc)
}

func main() {
	flag.Parse()
	src := jsbind.NewSource()

	if *inputFlag == "" {
		log.Fatal("Must provide an input file")
	}
	script, err := src.Compile(*inputFlag, nil)
	if err != nil {
		log.Fatal(err)
	}

	_, err = src.Run(script)
	if err != nil {
		log.Fatal(err)
	}

	if *objFlag == "" {
		log.Fatal("Must provide library object name")
	}
	obj, err := src.Get(*objFlag)
	if err != nil {
		log.Fatal(err)
	}

	if *nameFlag == "" {
		log.Fatal("Must provide binding's package name")
	}
	binding := jsbind.New(*nameFlag)
	binding.Add(*nameFlag, obj.Object())

	binding.Export(*nameFlag + ".go")
}
