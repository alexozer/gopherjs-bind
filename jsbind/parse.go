package jsbind

import "github.com/robertkrimen/otto"

const hacks = `
function Float32Array() {}
function Float64Array() {}

Float32Array.prototype = Array.prototype;
Float64Array.prototype = Array.prototype;

var self = {}`

type Source struct {
	otto.Otto
}

func NewSource() *Source {
	source := Source{*otto.New()}

	_, err := source.Run(hacks)
	if err != nil {
		panic(err) // Should never happen; why wouldn't hard-coded JS work?
	}

	return &source
}
