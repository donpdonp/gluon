package vm

import "fmt"
import "github.com/robertkrimen/otto"

type VM struct {
	Owner string
	Name  string
	Url   string
	Js    *otto.Otto
}

func Factory(owner string) *VM {
	new_vm := VM{Owner: owner,
		Js: otto.New()}
	return &new_vm
}

func (vm *VM) Eval(js_code string) error {
	fmt.Println(string(js_code)[0:15])
	fmt.Println("--eval begins--")

	src, err := vm.Js.Compile("", js_code)

	if err != nil {
		fmt.Println("compile failed!", err)
		return err
	} else {
		fmt.Println("compile good!")
		setup, err := vm.Js.Run(src)
		if err != nil {
			fmt.Println("eval failed", err, vm.Js.Context().Stacktrace)
			return err
		} else {
			descriptor_value, err := setup.Call(setup)
			if err != nil {
				fmt.Println("js func setup eval fail")
				return err
			} else {
				descriptor_map, _ := descriptor_value.Export()
				descriptor := descriptor_map.(map[string]interface{})
				vm.Name = descriptor["name"].(string)
				return nil
			}
		}
	}
}
