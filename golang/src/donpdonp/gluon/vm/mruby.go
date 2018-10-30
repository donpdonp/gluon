package vm

/*
#cgo CFLAGS: -I../../../../../mruby/include
#cgo LDFLAGS: -L../../../../../mruby/build/host/lib -lmruby -lm
#include "stdlib.h"
#include "mruby.h"
#include "mruby/compile.h"
#include "mruby/string.h"
#include "mruby/hash.h"
#include "mruby/array.h"
#include "mruby/variable.h"

extern mrb_value Emit(mrb_state*, mrb_value);
//MRB_API void mrb_define_class_method(mrb_state *, struct RClass *, const char *, mrb_func_t, mrb_aspec);
static void go_mrb_define_class_method(mrb_state *a, struct RClass *b, const char *c, int d, mrb_aspec e) {
  mrb_define_class_method(a,b,c,Emit,e);
}

static const char*
mruby_stringify_json_cstr(mrb_state* vm, mrb_value val) {
  struct RClass* clazz = mrb_module_get(vm, "JSON");
  mrb_value str = mrb_funcall(vm, mrb_obj_value(clazz), "generate", 1, val);
  return mrb_string_value_cstr(vm, &str);
}

// cgo doesnt like this define
//#define MRB_ARGS_REQ(n)     ((mrb_aspec)((n)&0x1f) << 18)
static mrb_aspec args_req(int n) { return ((n)&0x1f) << 18; }
*/
import "C"
import "unsafe"

import "fmt"
import "errors"

type RubyVM struct {
	state *C.mrb_state
}

func rubyfactory() *RubyVM {
	state := C.mrb_open()
	ruby_class := C.mrb_define_module(state, C.CString("Gluon"))

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
	fmt.Printf("go ruby emit callback ok\n")
	return C.mrb_value{}
}

func (vm *VM) EvalRuby(code string) (string, error) {
	context := C.mrbc_context_new(vm.Ruby.state)
  code_cstr := C.CString(code)
	parser_state := C.mrb_parse_string(vm.Ruby.state, code_cstr, context)
	if parser_state == nil {
		return "", errors.New("parse err")
	}
	proc := C.mrb_generate_code(vm.Ruby.state, parser_state)
	C.mrb_parser_free(parser_state)
	C.free(unsafe.Pointer(code_cstr))
	root_object := C.mrb_top_self(vm.Ruby.state)
	result := C.mrb_run(vm.Ruby.state, proc, root_object)
	if result.tt == C.MRB_TT_EXCEPTION {
		fmt.Println("ruby func setup eval fail")
		return "", errors.New("setup err")
	}
	fmt.Printf("ruby func setup eval good, return ruby type: %v\n", result.tt)
	json := C.mruby_stringify_json_cstr(vm.Ruby.state, result)
	return C.GoString(json), nil
}
