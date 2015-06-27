#include <hiredis/hiredis.h>

#include "mruby.h"
#include "mruby/compile.h"
#include "mruby/string.h"
#include "mruby/hash.h"
#include "mruby/array.h"
#include "mruby/variable.h"
#include "mruby-json/src/parson.h"

#if __STDC_VERSION__ < 199901L
#pragma message "C compiler is older than C99"
#endif

#ifndef MAIN_H
#define MAIN_H

struct ruby_vm_t {
  mrb_state* state;
  const char* owner;
};

typedef struct ruby_vm_t ruby_vm;

/* main.c */
void admin_setup();
void mainloop(JSON_Object* config);
void send_result(redisContext *redis_pub, const char* id, const char* json);


/* admin callbacks */
static mrb_value my_machine_add(mrb_state *mrb, mrb_value self);
static mrb_value my_machine_get(mrb_state *mrb, mrb_value self);
static mrb_value my_machine_list(mrb_state *mrb, mrb_value self);
static mrb_value my_machine_eval(mrb_state *mrb, mrb_value self);
static mrb_value my_emit(mrb_state *mrb, mrb_value self);
/* callbacks */
static mrb_value my_db_get(mrb_state *mrb, mrb_value self);
static mrb_value my_db_set(mrb_state *mrb, mrb_value self);
static mrb_value my_db_del(mrb_state *mrb, mrb_value self);
static mrb_value my_http_get(mrb_state *mrb, mrb_value self);


/* VM list */
int machines_add(const char* name);
int machines_find(const char* name);
int machines_get_as_ruby(const char* name, int i);

/* mruby calls */
const char* mruby_eval(ruby_vm vm, const char* code);
const char* mruby_stringify_json(ruby_vm vm, mrb_value val);
void mruby_parse_file(ruby_vm vm, const char* filename);
mrb_value mruby_json_parse(ruby_vm vm, const char* json);
mrb_value mruby_dispatch(ruby_vm vm, mrb_value msg);

/* libcurl */
size_t curl_on_page(char *ptr, size_t size, size_t nmemb, void *userdata);
struct CurlStr {
  char *memory;
  size_t size;
};

#endif
