package vm

import "github.com/go-interpreter/wagon/exec"
import "github.com/go-interpreter/wagon/validate"
import "github.com/go-interpreter/wagon/wasm"

import "log"
import "bytes"

var boot_wasm []byte

func wasmfactory() *exec.VM {
  // placeholder
  //module := wasm.NewModule()
  boot_wasm = []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x05, 0x01, 0x60, 0x00, 0x01, 0x7f, 0x03, 0x02, 0x01, 0x00, 0x07, 0x08, 0x01, 0x04, 0x6d, 0x61, 0x69, 0x6e, 0x00, 0x00, 0x0a, 0x07, 0x01, 0x05, 0x00, 0x41, 0x2a, 0x0f, 0x0b}
	module, _ := wasm.ReadModule(bytes.NewReader(boot_wasm), importer)
	vm, _ := exec.NewVM(module)
	return vm
}

func (vm *VM) EvalWasm(code string) (string, error) {
	module, err := wasm.ReadModule(bytes.NewReader(boot_wasm), importer)
	if err != nil {
		log.Fatalf("could not read module: %v", err)
	}

	verify := true
	if verify {
		err = validate.VerifyModule(module)
		if err != nil {
			log.Fatalf("could not verify module: %v", err)
		}
	}

	if module.Export == nil {
		log.Fatalf("module has no export section")
	}

	_, err = exec.NewVM(module)
	for name, e := range module.Export.Entries {
		log.Printf("EvalWasm Export entry: %#v %#v\n", e, name)
	}
	return "wasm-boss", nil
}

func importer(name string) (*wasm.Module, error) {
	log.Printf("importer: %s\n", name)
	return nil, nil
}
