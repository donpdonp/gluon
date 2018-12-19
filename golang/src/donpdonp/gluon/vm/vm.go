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
		fmt.Printf("vm ruby go\n")
		new_vm.Ruby = rubyfactory()
	}
	if lang == "javascript" {
		fmt.Printf("vm js go\n")
		new_vm.Js = otto.New()
	}
	if lang == "webassembly" {
		fmt.Printf("vm webasm go\n")
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

func (vm *VM) Eval(code string) (string, error) {
	lang := vm.Lang()
	if lang == "javascript" {
		return vm.EvalJs(code)
	}
	if lang == "ruby" {
		return vm.EvalRuby(code)
	}
	if lang == "webassembly" {
		return vm.EvalWasm(code)
	}
	return "", errors.New(lang)
}
