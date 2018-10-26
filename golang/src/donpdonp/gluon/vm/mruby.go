package vm

/*
#cgo CFLAGS: -I../../../../../mruby/include
#cgo LDFLAGS: -L../../../../../mruby/build/host/lib -lmruby -lm
#include "mruby.h"
*/
import "C"

import "fmt"

func justhere() {
  a := C.mrb_state{}
  fmt.Printf("%v\n", a.top_self)
  C.mrb_open()
}
