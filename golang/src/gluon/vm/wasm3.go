package vm

func (vm *VM) EvalWasm(code []byte) (string, error) {
	module, err := vm.Wasm.ParseModule(code)
	if err != nil {
		panic(err)
	}
	module, err = vm.Wasm.LoadModule(module)
	if err != nil {
		panic(err)
	}
	return "wasm loaded " + string(len(code)), nil
}
