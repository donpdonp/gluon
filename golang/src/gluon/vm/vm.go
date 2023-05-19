package vm

import (
	"donpdonp/gluon/util"
	"errors"

	"github.com/matiasinsaurralde/go-wasm3"
	"github.com/robertkrimen/otto"
)

type VM struct {
	Id    string
	Q     chan map[string]interface{}
	Owner string
	Name  string
	Url   string
	Js    *otto.Otto
	Wasm  *wasm3.Runtime
}

func Factory(owner string, lang string) *VM {
	new_vm := VM{Id: util.Snowflake(), Owner: owner}
	new_vm.Q = make(chan map[string]interface{}, 1000)
	if lang == "javascript" {
		new_vm.Js = otto.New()
	}
	if lang == "webassembly" {
		new_vm.Wasm = wasm3.NewRuntime(&wasm3.Config{
			Environment: wasm3.NewEnvironment(),
			StackSize:   64 * 1024,
		})
	}
	return &new_vm
}

func (vm *VM) Lang() string {
	if vm.Js != nil {
		return "javascript"
	}
	if vm.Wasm != nil {
		return "webassembly"
	}
	return "unknown"
}

func (vm *VM) EvalGo(params_jbytes []byte) (string, error) {
	params_json := string(params_jbytes)
	var callBytes []byte
	if vm.Lang() == "javascript" {
		callBytes = []byte("go(" + params_json + ")")
		return vm.Eval(vm.EvalDependencies(callBytes))
	}
	if vm.Lang() == "webassembly" {
	}
	return "", errors.New("")
}

func (vm *VM) EvalDependencies(code []byte) map[string][]byte {
	dependencies := map[string][]byte{}
	lang := vm.Lang()
	if lang == "webassembly" {
		//dependencies = vm.EvalDependencyWasm(code)
	}
	dependencies["main"] = code
	return dependencies
}

func (vm *VM) Eval(dependencies map[string][]byte) (string, error) {
	code := dependencies["main"]
	lang := vm.Lang()
	if lang == "javascript" {
		return vm.EvalJs(string(code))
	}
	return "", errors.New(lang)
}
