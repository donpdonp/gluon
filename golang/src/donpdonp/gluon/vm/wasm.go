package vm

/*
#cgo CFLAGS: -I../../../../../../webasm/WAVM/Include
#cgo CFLAGS: -I../../../../../../webasm/WAVM
#cgo LDFLAGS: -L../../../../../../webasm/WAVM/Lib/WASM -lWASM
#cgo LDFLAGS: -L../../../../../../webasm/WAVM/ -lgobridge
#include "gobridge.h"
*/
import "C"

import "fmt"

func wasmfactory() {
	ir := C.irModule{}
	fmt.Printf("wasm irModule: %v\n", ir)
	C.loadModule(C.CString("a"))
}
