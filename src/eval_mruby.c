#include "main.h"
#include "string.h"

const char*
mruby_eval(ruby_vm vm, const char* code){
  printf("eval_mruby_json vm:%s code: %s\n", vm.owner, code);

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
  printf("run result type #%d\n", result.tt);
  if(result.tt == MRB_TT_EXCEPTION){
    mrb_value exv = mrb_obj_value(vm.state->exc);
    exv = mrb_funcall(vm.state, exv, "inspect", 0);
    vm.state->exc = 0;
    puts(mrb_string_value_cstr(vm.state, &exv));
    return "{\"error\":\"ruby exception\"}";
  }
  const char* json = mruby_stringify_json(vm.state, result);
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
mruby_stringify_json(mrb_state* mrb, mrb_value val) {
  struct RClass* clazz = mrb_module_get(mrb, "JSON");
  mrb_value str = mrb_funcall(mrb, mrb_obj_value(clazz), "stringify", 1, val);
  return mrb_string_value_cstr(mrb, &str);
}
