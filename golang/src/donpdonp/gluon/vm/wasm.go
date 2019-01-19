package vm

import "github.com/go-interpreter/wagon/exec"
import "github.com/go-interpreter/wagon/validate"
import "github.com/go-interpreter/wagon/wasm"

import "log"
import "bytes"
import "errors"
import "encoding/json"

var boot_wasm []byte

// once a wasm VM is created with a module, the module is inaccessable
// so keep a reference here
type WasmProcess struct {
	wagon  *exec.VM
	module *wasm.Module
}

func WasmProcessNewVM(module *wasm.Module) (*WasmProcess, error) {
	var wp WasmProcess
	wvm, err := exec.NewVM(module)
	if err != nil {
		log.Printf("WasmProcess NewVM failed : %#v", err)
	} else {
		wp = WasmProcess{wagon: wvm, module: module}
		exportCount := 0
		if module.Export != nil { exportCount = len(module.Export.Entries) }
		globalCount := 0
		if module.Global != nil { globalCount = len(module.Global.Globals) }
		log.Printf("WasmProcess NewVM: %d exports. %d globals.", exportCount, globalCount)
    log.Printf("WasmProcess NewVM: %#v\n", module.Export.Entries)
	}
	return &wp, err
}

func (wp *WasmProcess) GetModule() *wasm.Module { return wp.module }
func (wp *WasmProcess) GetWagon() *exec.VM { return wp.wagon }

func wasmfactory() *WasmProcess {
	// placeholder
	//module := wasm.NewModule()
	boot_wasm = []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x05, 0x01,
		                 0x60, 0x00, 0x01, 0x7f, 0x03, 0x02, 0x01, 0x00, 0x07, 0x08, 0x01,
		                 0x04, 0x6d, 0x61, 0x69, 0x6e, 0x00, 0x00, 0x0a, 0x07, 0x01, 0x05,
		                 0x00, 0x41, 0x2a, 0x0f, 0x0b}
	module, _ := wasm.ReadModule(bytes.NewReader(boot_wasm), importerDummy)
	log.Printf("-WasmProcessNewVM for boot_wasm/wasmfactory\n")
	wvm, _ := WasmProcessNewVM(module)
	return wvm
}

func (vm *VM) EvalWasm(dependencies map[string][]byte) (string, error) {
	code := dependencies["main"]
	log.Printf("evalwasm main module from %d bytes", len(code))
	module, err := wasm.ReadModule(bytes.NewReader(code),
		func(name string) (*wasm.Module, error) {
			return importer(dependencies, name)
		})
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

		if module.Memory == nil {
			log.Printf("module has no memory section")
		} else {
			for idx, e := range module.Memory.Entries {
				log.Printf("module.Memory #%d  %#v\n", idx, e.Limits)
			}
		}
		if module.Export == nil {
			log.Printf("module has no export section")
		}

  	log.Printf("-WasmProcessNewVM for main\n")
		wp, err := WasmProcessNewVM(module)
		if err != nil {
  		vm.Wasm = wp
			log.Printf("exec.NewVM: %v", err)
		}
	}
	return vm.EvalGoWasm("setup")
}

func (vm *VM) EvalGoWasm(ffname string) (string, error) {
	var err error
	result := ""
	module := vm.Wasm.GetModule()
	log.Printf("--evalGoWasm %s\n", ffname)
	log.Printf("module.Function.Types %#v\n", module.Function.Types)
	log.Printf("module.Types.Entries %#v\n", module.Types.Entries)
	if module.Export != nil {
		for fname, e := range module.Export.Entries {
			i := int64(e.Index)
			log.Printf("module.Export #%d %#v %#v\n", i, fname, e.Kind.String())
			if e.Kind == wasm.ExternalFunction {
				fidx := module.Function.Types[int(i)]
				ftype := module.Types.Entries[int(fidx)]
				log.Printf("call %s(%#v) %#v \n", fname, ftype.ParamTypes, ftype.ReturnTypes)

				var output interface{}
				switch len(ftype.ParamTypes) {
				case 2:
					log.Printf("Wagon.Exec %s w/ 2 params (7,8)", fname)
  				output, err = vm.Wasm.GetWagon().ExecCode(i, 7, 8)
				case 1:
					log.Printf("Wagon.Exec %s w/ 1 params (7)", fname)
  				output, err = vm.Wasm.GetWagon().ExecCode(i, 7)
				default:
					log.Printf("Wagon.Exec %s w/ 0 params", fname)
  				output, err = vm.Wasm.GetWagon().ExecCode(i)
			  }
				if err != nil {
					log.Printf("wasm err=%v", err)
				} else {
					_, _ = json.Marshal(output)
					memory := vm.Wasm.GetWagon().Memory() // byte array
					if memory != nil {
						len := int(memory[0])
						log.Printf("wasm out: memory[0] %#v \n", len)
						start := 1
						result = string(memory[start : start+len])
						log.Printf("wasm returned %d. memory[0] = %d. str[0:len]= %#v\n", output, len, result)
					} else {
						log.Printf("wasm returned %d. no memory in this module.\n", output)
					}
				}
			}
		}
	} else {
		log.Printf("module.Export is nil\n")
	}
  return result, err
}

func importer(dependencies map[string][]byte, name string) (*wasm.Module, error) {
	var err error
	var module *wasm.Module
	if dependencies[name] != nil {
		module, err = wasm.ReadModule(bytes.NewReader(dependencies[name]), importerDummy)
		log.Printf("webasm imported: %s %d bytes %#v\n", name, len(dependencies[name]), err)
	} else {
		err = errors.New(name + " not found")
		log.Printf("webasm import: %s not found\n", name)
	}
	return module, err
}

func importerDummy(name string) (*wasm.Module, error) {
	log.Printf("webasm importDummy ignoring module %s\n", name)
	return nil, nil
}

func (vm *VM) EvalDependencyWasm(code []byte) map[string][]byte {
	dependencies := map[string][]byte{}
	wasm.ReadModule(bytes.NewReader(code), func(name string) (*wasm.Module, error) {
		dependencies[name] = nil
		return wasm.NewModule(), nil
	})
	return dependencies
}
