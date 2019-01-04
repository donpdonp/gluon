package vm

import "fmt"
import "errors"
import "github.com/robertkrimen/otto"
import "github.com/go-interpreter/wagon/exec"

type VM struct {
	Owner string
	Name  string
	Url   string
	Js    *otto.Otto
	Ruby  *RubyVM
	Wasm  *exec.VM
}

func Factory(owner string, lang string) *VM {
	new_vm := VM{Owner: owner}
	if lang == "ruby" {
		new_vm.Ruby = rubyfactory()
	}
	if lang == "javascript" {
		new_vm.Js = otto.New()
	}
	if lang == "webassembly" {
		new_vm.Wasm = wasmfactory()
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
	if vm.Wasm != nil {
		return "webassembly"
	}
	return "unknown"
}

func (vm *VM) Eval(code []byte) (string, error) {
	lang := vm.Lang()
	if lang == "javascript" {
		return vm.EvalJs(string(code))
	}
	if lang == "ruby" {
		return vm.EvalRuby(string(code))
	}
	if lang == "webassembly" {
		return vm.EvalWasm(code)
	}
	return "", errors.New(lang)
}
