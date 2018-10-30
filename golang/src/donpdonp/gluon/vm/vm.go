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

func (vm *VM) EvalJs(js_code string) (string, error) {
	result, err := vm.Js.Run(js_code)
	if err != nil {
		fmt.Println("evaljs Run failed", err, vm.Js.Context().Stacktrace)
		return "", err
	} else {
		fmt.Printf("evaljs JSON.stringify %v\n", result)
		otto_json, err := vm.Js.Call("JSON.stringify", nil, result)
		json := ""
		fmt.Printf("evaljs JSON.stringify out: %v err: %v\n", otto_json, err)
		if err == nil {
			fmt.Printf("evaljs JSON.stringify err: %v\n", err)
		} else {
			thing, _ := otto_json.Export()
			json = thing.(string)
		}
		return json, nil //descriptor_value json
	}
}

// ug, hack to support first time setup that returns a function
func (vm *VM) FirstEvalJs(js_code string) (string, error) {
	src, err := vm.Js.Compile("", js_code)
	if err != nil {
		fmt.Println("js compile failed!", err)
		return "", err
	} else {
		fmt.Println("js compile good!")
		setup, err := vm.Js.Run(src)
		if err != nil {
			fmt.Println("eval failed", err, vm.Js.Context().Stacktrace)
			return "", err
		} else {
			descriptor_value, err := setup.Call(setup)
			if err != nil {
				fmt.Println("js func setup eval fail %v", err)
				return "", err
			} else {
				otto_json, _ := vm.Js.Call("JSON.stringify", nil, descriptor_value)
				json, _ := otto_json.Export()
				return json.(string), nil //descriptor_value json
			}
		}
	}
}
