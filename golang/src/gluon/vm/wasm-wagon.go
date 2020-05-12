package vm

import "github.com/go-interpreter/wagon/exec"
import "github.com/go-interpreter/wagon/validate"
import "github.com/go-interpreter/wagon/wasm"
import "github.com/go-interpreter/wagon/wasm/leb128"

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
	}
	return &wp, err
}

func (wp *WasmProcess) GetModule() *wasm.Module { return wp.module }
func (wp *WasmProcess) GetWagon() *exec.VM      { return wp.wagon }

func wasmfactory() *WasmProcess {
	// placeholder
	//module := wasm.NewModule()
	boot_wasm = []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x05, 0x01,
		0x60, 0x00, 0x01, 0x7f, 0x03, 0x02, 0x01, 0x00, 0x07, 0x08, 0x01,
		0x04, 0x6d, 0x61, 0x69, 0x6e, 0x00, 0x00, 0x0a, 0x07, 0x01, 0x05,
		0x00, 0x41, 0x2a, 0x0f, 0x0b}
	module, _ := wasm.ReadModule(bytes.NewReader(boot_wasm), importerDummy)
	log.Printf("-WasmProcessNewVM for boot_wasm/wasmfactory\n")
	moduleSummary(module)
	wvm, _ := WasmProcessNewVM(module)
	return wvm
}

func (vm *VM) EvalWasm(dependencies map[string][]byte) (string, error) {
	code := dependencies["main"]
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

		log.Printf("-WasmProcessNewVM for main module from %d bytes", len(code))
		moduleSummary(module)
		wp, err := WasmProcessNewVM(module)
		if err != nil {
			log.Printf("exec.NewVM err: %v", err)
		} else {
			vm.Wasm = wp
		}
	}
	return vm.WasmCall("setup", nil)
}

func (vm *VM) WasmCall(ffname string, params interface{}) (string, error) {
	var err error
	result := ""
	module := vm.Wasm.GetModule()
	log.Printf("--WasmCall %s\n", ffname)
	memory := vm.Wasm.GetWagon().Memory() // byte array
	if memory != nil {
		if module.Export != nil {
			inbufExport := findExport(module.Export, "inbuf")
			if inbufExport != nil {
				log.Printf("Export found name: %s type: %s\n", inbufExport.FieldStr, inbufExport.Kind)
				functionExport := findExport(module.Export, ffname)
				if functionExport != nil {
					i := int64(functionExport.Index)
					if functionExport.Kind == wasm.ExternalFunction {
						fidx := module.Function.Types[i]
						ftype := module.Types.Entries[fidx]
						log.Printf("call %s(%#v) %#v \n", ffname, ftype.ParamTypes, ftype.ReturnTypes)

						var output interface{}
						switch len(ftype.ParamTypes) {
						case 1:
							params_json, _ := json.Marshal(params)
							outStart := findGlobalInt(module, "inbuf")
							copy(memory[outStart:], params_json)
							log.Printf("WasmCall Wagon.Exec %s w/ 1 param %s (%d) copied to inbuf[%d:]",
								ffname, string(params_json), uint64(len(params_json)), outStart)
							output, err = vm.Wasm.GetWagon().ExecCode(i, uint64(len(params_json)))
						case 0:
							log.Printf("WasmCall Wagon.Exec %s w/ 0 params", ffname)
							output, err = vm.Wasm.GetWagon().ExecCode(i)
						default:
							err = errors.New("unknown function signature")
						}
						if err != nil {
							log.Printf("wasm err=%v", err)
						} else {
							_, _ = json.Marshal(output)
							start := output.(uint32)
							len := uint32(memory[start])
							dataStart := start + 1
							result = string(memory[dataStart : dataStart+len])
							log.Printf("%s returned %d. memory[%d] = %d. str[0:%d]= %#v\n",
								ffname, output, output, len, len, result)
						}
					}
				} else {
					log.Printf("wasm function %s not found\n", ffname)
				}
			} else {
				log.Printf("wasm export inbuf not found\n")
			}
		} else {
			log.Printf("wasm call stopped. module.Export is nil\n")
		}
	} else {
		log.Printf("wasm call aborted. no memory in this module.\n")
	}
	return result, err
}

func moduleSummary(module *wasm.Module) {
	exportCount := 0
	if module.Export != nil {
		exportCount = len(module.Export.Entries)
	}
	globalCount := 0
	if module.Global != nil {
		globalCount = len(module.Global.Globals)
	}
	memoryCount := 0
	if module.Memory != nil {
		memoryCount = len(module.Memory.Entries)
	}
	log.Printf("moduleSummary: %d exports. %d globals. %d memories.",
		exportCount, globalCount, memoryCount)
	if module.Export != nil {
		for _, e := range module.Export.Entries {
			log.Printf("module.Export %s %#v #%d\n", e.Kind, e.FieldStr, e.Index)
		}
	}
	if module.Global != nil {
		for idx, e := range module.Global.Globals {
			gint, _ := leb128.ReadVarint32(bytes.NewReader(e.Init[1:]))
			log.Printf("module.Global #%d %s %#v\n", idx, e.Type.Type, gint)
		}
	}
	if module.Memory != nil {
		for _, e := range module.Memory.Entries {
			log.Printf("module.Memory %d pages\n", e.Limits.Initial)
		}
	}
}

func findGlobalInt(module *wasm.Module, name string) int {
	if module.Export != nil {
		export := findExport(module.Export, name)
		if export != nil {
			if module.Global != nil {
				global := module.Global.Globals[export.Index]
				gint, _ := leb128.ReadVarint32(bytes.NewReader(global.Init[1:]))
				return int(gint)
			}
		}
	}
	log.Printf("findGlobal %s not found!\n", name)
	return 0
}

func findExport(export *wasm.SectionExports, ffname string) *wasm.ExportEntry {
	for fname, e := range export.Entries {
		if fname == ffname {
			return &e
		}
	}
	return nil
}

func importer(dependencies map[string][]byte, name string) (*wasm.Module, error) {
	var err error
	var module *wasm.Module
	if dependencies[name] != nil {
		module, err = wasm.ReadModule(bytes.NewReader(dependencies[name]), importerDummy)
		log.Printf("--webasm imported: %#v %d bytes %#v\n", name, len(dependencies[name]), err)
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
