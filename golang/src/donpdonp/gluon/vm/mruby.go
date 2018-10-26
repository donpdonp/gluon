package vm

/*
#cgo CFLAGS: -I../../../../../mruby/include
#cgo LDFLAGS: -L../../../../../mruby/build/host/lib -lmruby -lm
#include "mruby.h"

extern mrb_value Emit(mrb_state*, mrb_value);
//MRB_API void mrb_define_class_method(mrb_state *, struct RClass *, const char *, mrb_func_t, mrb_aspec);
static void go_mrb_define_class_method(mrb_state *a, struct RClass *b, const char *c, int d, mrb_aspec e) {
  mrb_define_class_method(a,b,c,Emit,e);
}

// cgo doesnt like this define
//#define MRB_ARGS_REQ(n)     ((mrb_aspec)((n)&0x1f) << 18)
static mrb_aspec args_req(int n) { return ((n)&0x1f) << 18; }
*/
import "C"

import "fmt"

type RubyVM struct {
  state *C.mrb_state
}

func rubyfactory() *RubyVM {
  state := C.mrb_open()
  fmt.Printf("%v\n", state.top_self)
  ruby_class := C.mrb_define_module(state, C.CString("Gluon"));

  C.go_mrb_define_class_method(
    state,
    ruby_class,
    C.CString("emit"),
    0,
    C.args_req(1))
  return &RubyVM{state: state}
}

//export Emit
func Emit(state *C.mrb_state, value C.mrb_value) C.mrb_value {
  return C.mrb_value{}
}
