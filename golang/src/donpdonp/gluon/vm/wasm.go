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

func (vm *VM) EvalWasm(code []byte) (string, error) {
	module, err := wasm.ReadModule(bytes.NewReader(code), importer)
	if err != nil {
		log.Printf("could not read module: %v", err)
	} else {
  	verify := true
  	if verify {
  		err = validate.VerifyModule(module)
  		if err != nil {
  			log.Printf("could not verify module: %v", err)
  		}
  	}

  	if module.Export == nil {
  		log.Printf("module has no export section")
  	}

  	wvm, err := exec.NewVM(module)
    if err != nil {
      log.Printf("exec.NewVM: %v", err)
    } else {
    	for name, e := range module.Export.Entries {
    		log.Printf("EvalWasm Export entry: %#v %#v\n", e, name)
    	}

      log.Printf("Module Memory %#v\n", module.Memory);
    	for name, e := range module.Export.Entries {
    		i := int64(e.Index)
    		fidx := module.Function.Types[int(i)]
    		ftype := module.Types.Entries[int(fidx)]
    		log.Printf("%s(%#v) %#v \n", name, ftype.ParamTypes, ftype.ReturnTypes)

    		output, err := wvm.ExecCode(i)
    		if err != nil {
    			log.Printf("wasm err=%v", err)
    		}
        memory := wvm.Memory()
    		log.Printf("wasm out: %d. %d bytes memory\n", output, len(memory))
    	}
    }
  }
	return "wasm-boss", nil
}

func importer(name string) (*wasm.Module, error) {
	log.Printf("importer: %s\n", name)
	return nil, nil
}
