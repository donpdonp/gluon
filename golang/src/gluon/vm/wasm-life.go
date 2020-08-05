package vm

func (vm *VM) EvalWasm(code []byte) (string, error) {
	return "wasm loaded " + string(len(code)), nil
}
