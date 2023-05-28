package vm

import (
	"donpdonp/gluon/util"
	"errors"
	"fmt"

	"github.com/matiasinsaurralde/go-wasm3"
	"github.com/robertkrimen/otto"
)

type VM struct {
	Id       string
	Q        chan map[string]interface{}
	Owner    string
	Name     string
	Url      string
	Js       *otto.Otto
	Wasm     *wasm3.Runtime
	WasmGoFn *wasm3.Function
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
			EnableWASI:  true,
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
		return vm.Eval(callBytes)
	}
	if vm.Lang() == "webassembly" {
		return vm.Eval(callBytes)
	}
	return "", errors.New("")
}

func (vm *VM) EvalWasm(code []byte) {
	module, err := vm.Wasm.ParseModule(code)
	if err != nil {
		fmt.Printf("vm.EvalDependencies wasm ParseModule err %v\n", err)
	}
	module, err = vm.Wasm.LoadModule(module)
	if err != nil {
		fmt.Printf("vm.EvalDependencies wasm LoadModule %v\n", err)
	}
	vm.Wasm.FindFunction("go") // TODO
	fn, err := module.GetFunctionByName("go")
	if err == nil {
		fmt.Printf("module found go fn. saving.\n")
		vm.WasmGoFn = fn
	} else {
		fmt.Printf("vm.Eval wasm findFunction err %s: %v\n", "go", err)
	}
}

func (vm *VM) Eval(code []byte) (string, error) {
	lang := vm.Lang()
	if lang == "javascript" {
		return vm.EvalJs(string(code))
	} else if lang == "webassembly" {
		fmt.Printf("vm.Eval calling wasm go fn.\n")
		vm.WasmGoFn.CallWithArgs("A")
		//fmt.Printf("vm.Eval done wasm go fn. e: %v\n", err)
		// result, err :=
		// fmt.Printf("vm.Eval wasm %v: %v (err %v)\n", string(code), result, err)
		return "wasm fn return todo", nil
	}
	return "", errors.New(fmt.Sprintf("unknown language: %s", lang))
}
