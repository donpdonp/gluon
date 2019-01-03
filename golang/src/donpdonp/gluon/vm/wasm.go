package vm

import "github.com/go-interpreter/wagon/exec"
import "github.com/go-interpreter/wagon/validate"
import "github.com/go-interpreter/wagon/wasm"

import "log"
import "bytes"
import "encoding/json"

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
  result := ""
	module, err := wasm.ReadModule(bytes.NewReader(code), importer)
	if err != nil {
		log.Printf("could not read module from %d bytes %v", len(code), err)
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
      for idx, e := range module.Memory.Entries {
        log.Printf("Module Memory #%d  %#v\n", idx, e.Limits);
      }

    	for fname, e := range module.Export.Entries {
    		log.Printf("Module Export: %s %#v\n", fname, e)
    	}

    	for fname, e := range module.Export.Entries {
    		i := int64(e.Index)
    		fidx := module.Function.Types[int(i)]
    		ftype := module.Types.Entries[int(fidx)]
    		log.Printf("call %s(%#v) %#v \n", fname, ftype.ParamTypes, ftype.ReturnTypes)

    		output, err := wvm.ExecCode(i, 7, 8)
    		if err != nil {
    			log.Printf("wasm err=%v", err)
    		} else {
          _, _ = json.Marshal(output)
          memory := wvm.Memory() // byte array
          len := int(memory[0])
          log.Printf("wasm out: memory[0] %#v \n", len)
          start := 1
          result = string(memory[start:start+len])
        }
    		log.Printf("wasm out: %d. %d bytes memory. %#v %#v\n", output, len(result), result )
    	}
    }
  }
	return result, err
}

func importer(name string) (*wasm.Module, error) {
	log.Printf("webasm importer: %s\n", name)
	return nil, nil
}
