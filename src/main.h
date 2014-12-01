#include "mruby.h"
#include "mruby/compile.h"
#include "mruby/string.h"
#include "mruby-json/src/parson.h"

#ifndef MAIN_H
#define MAIN_H

struct ruby_vm_t {
  mrb_state* state;
  const char* owner;
};

typedef struct ruby_vm_t ruby_vm;

const char* eval_mruby_json(ruby_vm vm, const char* code);
const char* mruby_stringify_json(mrb_state* mrb, mrb_value val);
void mainloop(JSON_Object* config);
void setup();
static mrb_value my_c_method(mrb_state *mrb, mrb_value self);
#endif