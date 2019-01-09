package vm

import "errors"
import "encoding/json"

import "github.com/robertkrimen/otto"
import "github.com/go-interpreter/wagon/exec"

import "donpdonp/gluon/util"

type VM struct {
	Id    string
	Q     chan map[string]interface{}
	Owner string
	Name  string
	Url   string
	Js    *otto.Otto
	Ruby  *RubyVM
	Wasm  *exec.VM
}

func Factory(owner string, lang string) *VM {
	new_vm := VM{Id: util.Snowflake(), Owner: owner}
	new_vm.Q = make(chan map[string]interface{}, 1000)
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

func (vm *VM) EvalGo(params_jbytes []byte) (string, error) {
	params_json := string(params_jbytes)
	var callBytes []byte
	if vm.Lang() == "javascript" {
		callBytes = []byte("go(" + params_json + ")")
	}
	if vm.Lang() == "ruby" {
		params_double_jbytes, _ := json.Marshal(params_json)
		params_double_json := string(params_double_jbytes)
		callBytes = []byte("go(JSON.parse(" + params_double_json + "))")
	}
	return vm.Eval(callBytes)
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
