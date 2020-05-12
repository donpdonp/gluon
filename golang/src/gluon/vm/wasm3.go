package vm

import wasm3 "github.com/matiasinsaurralde/go-wasm3"
import "log"

type Wasm3Process struct {
	runtime *wasm3.Runtime
	module  *wasm3.Module
}

func wasm3factory() *Wasm3Process {
	runtime := wasm3.NewRuntime(&wasm3.Config{
		Environment: wasm3.NewEnvironment(),
		StackSize:   64 * 1024,
	})
	return &Wasm3Process{runtime: runtime, module: nil}
}

func (vm *VM) Wasm3Call(fnName string, params interface{}) (string, error) {
	fn, err := vm.Wasm3.runtime.FindFunction(fnName)
	if err != nil {
		panic(err)
	}
	log.Printf("Found '%s' function (using runtime.FindFunction)", fnName)
	memoryLength := vm.Wasm3.runtime.GetAllocatedMemoryLength()
	log.Printf("Allocated memory (before function call) is: %d\n", memoryLength)
	result, _ := fn()
	log.Printf("%#v", result)
	return "{}", nil
}

func (vm *VM) EvalWasm3(dependencies map[string][]byte) (string, error) {
	main_bytes := dependencies["main"]
	module, err := vm.Wasm3.runtime.ParseModule(main_bytes)
	if err != nil {
		panic(err)
	}
	module, err = vm.Wasm3.runtime.LoadModule(module)
	if err != nil {
		panic(err)
	}
	vm.Wasm3.module = module
	return vm.Wasm3Call("setup", nil)
}
