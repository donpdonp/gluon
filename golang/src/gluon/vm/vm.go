package vm

import "errors"

import "github.com/robertkrimen/otto"
import "github.com/perlin-network/life/exec"

import "donpdonp/gluon/util"

type VM struct {
	Id    string
	Q     chan map[string]interface{}
	Owner string
	Name  string
	Url   string
	Js    *otto.Otto
	Wasm  *exec.VirtualMachine
}

func Factory(owner string, lang string) *VM {
	new_vm := VM{Id: util.Snowflake(), Owner: owner}
	new_vm.Q = make(chan map[string]interface{}, 1000)
	if lang == "javascript" {
		new_vm.Js = otto.New()
	}
	if lang == "webassembly" {
		life, _ := exec.NewVirtualMachine([]byte{}, exec.VMConfig{}, &exec.NopResolver{}, nil)
		new_vm.Wasm = life
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

func (vm *VM) EvalCallGo(params_json []byte) (string, error) {
	json := string(params_json)
	var callBytes []byte
	lang := vm.Lang()
	if lang == "javascript" {
		callBytes = []byte("go(" + json + ")")
		return vm.Eval(vm.EvalDependencies(callBytes))
	}
	if lang == "webassembly" {
		id, err_bool := vm.Wasm.GetFunctionExport("go")
		if err_bool {
			panic("go func not found")
		}
		result, err := vm.Wasm.Run(id)
		return string(result), err
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
	if lang == "webassembly" {
		return vm.EvalWasm(code)
	}
	return "", errors.New(lang)
}
