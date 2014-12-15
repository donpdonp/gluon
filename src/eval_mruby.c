#include "main.h"
#include "string.h"

const char*
mruby_eval(ruby_vm vm, const char* code){
  printf("mruby_eval vm:%s code: %s\n", vm.owner, code);

  mrbc_context* context;
  context = mrbc_context_new(vm.state);

  struct mrb_parser_state* parser_state;
  parser_state = mrb_parse_string(vm.state, code, context);
  if (parser_state == NULL) {
    fputs("parse error\n", stderr);
    return "{\"error\":\"parser error\"}";
  }

  if (0 < parser_state->nerr) {
    static char err[2048];
    sprintf(err, "{\"error\":\"line %d: %s\"}", parser_state->error_buffer[0].lineno+1,
                                                parser_state->error_buffer[0].message);
    return err;
  }

  struct RProc* proc;
  proc = mrb_generate_code(vm.state, parser_state);
  if (proc == NULL) {
    fputs("codegen error\n", stderr);
    return "{\"error\":\"codegen\"}";
  }
  mrb_parser_free(parser_state);

  mrb_value root_object;
  root_object = mrb_top_self(vm.state);

  mrb_value result;
  result = mrb_run(vm.state, proc, root_object);
  printf("eval result type #%d\n", result.tt);
  if(result.tt == MRB_TT_EXCEPTION){
    mrb_value exv = mrb_obj_value(vm.state->exc);
    exv = mrb_funcall(vm.state, exv, "inspect", 0);
    vm.state->exc = 0;
    puts(mrb_string_value_cstr(vm.state, &exv));
    return "{\"error\":\"ruby exception\"}";
  }
  const char* json = mruby_stringify_json(vm, result);
  return json;
}

void
mruby_parse_file(ruby_vm vm, const char* filename){
  FILE* admin_rb = fopen(filename, "r");
  mrbc_context* context = mrbc_context_new(vm.state);
  struct mrb_parser_state* parser_state;
  parser_state = mrb_parse_file(vm.state, admin_rb, context);
  fclose(admin_rb);

  struct RProc* proc;
  proc = mrb_generate_code(vm.state, parser_state);
  mrb_parser_free(parser_state);

  mrb_run(vm.state, proc, mrb_top_self(vm.state));
}

mrb_value
mruby_json_parse(ruby_vm vm, const char* json){
  struct RClass* clazz = mrb_module_get(vm.state, "JSON");
  mrb_value val = mrb_str_new_cstr(vm.state, json);
  mrb_value rjson = mrb_funcall(vm.state, mrb_obj_value(clazz), "parse", 1, val);
  return rjson;
}

const char*
mruby_stringify_json(ruby_vm vm, mrb_value val) {
  struct RClass* clazz = mrb_module_get(vm.state, "JSON");
  mrb_value str = mrb_funcall(vm.state, mrb_obj_value(clazz), "stringify", 1, val);
  return mrb_string_value_cstr(vm.state, &str);
}

mrb_value
mruby_dispatch(ruby_vm vm, mrb_value msg){
  mrb_value pre_check = mrb_const_get(vm.state, mrb_obj_value(vm.state->object_class), mrb_intern_cstr(vm.state, "Neur0n"));
  if(pre_check.tt == MRB_TT_MODULE) {
    struct RClass* clazz = mrb_module_get(vm.state, "Neur0n");
    mrb_value ret = mrb_funcall(vm.state, mrb_obj_value(clazz), "dispatch", 1, msg);
    return ret;
  }
}
