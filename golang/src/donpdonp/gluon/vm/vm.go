package vm

import "fmt"
import "errors"
import "github.com/robertkrimen/otto"

type VM struct {
	Owner string
	Name  string
	Url   string
	Js    *otto.Otto
	Ruby  *RubyVM
}

func Factory(owner string, lang string) *VM {
	new_vm := VM{Owner: owner}
	if lang == "ruby" {
		fmt.Printf("vm ruby go\n")
		new_vm.Ruby = rubyfactory()
	}
	if lang == "javascript" {
		fmt.Printf("vm js go\n")
		new_vm.Js = otto.New()
	}
	if lang == "webassembly" {
		fmt.Printf("vm webasm go\n")
		wasmfactory()
	}
	return &new_vm
}

func (vm *VM) Lang() string {
	if vm.Js != nil {
		return "javascript"
	}
	if vm.Ruby != nil {
		return "ruby"
	}
	return "unknown language"
}

func (vm *VM) Eval(code string) (string, error) {
	lang := vm.Lang()
	if lang == "javascript" {
		return vm.EvalJs(code)
	}
	if lang == "ruby" {
		return vm.EvalRuby(code)
	}
	return "", errors.New(lang)
}

