#include <stdlib.h>
#include <string.h>
#include <hiredis/hiredis.h>

#include "mruby.h"
#include "mruby/compile.h"
#include "mruby/string.h"
#include "mruby-json/src/parson.h"

#define CONFIG(key) json_object_dotget_string(config, key)

struct ruby_vm_t {
  mrb_state* state;
  const char* owner;
};
typedef struct ruby_vm_t ruby_vm;
ruby_vm* machines = NULL;
int machines_count = 0;
ruby_vm admin_vm;

const char* eval_mruby_json(ruby_vm, const char*);
const char* mruby_stringify_json(mrb_state* mrb, mrb_value val);
void mainloop(JSON_Object* config);
void setup();
static mrb_value my_c_method(mrb_state *mrb, mrb_value self);

int
main() {
  JSON_Value *config_json = json_parse_file("config.json");
  if(json_value_get_type(config_json) == JSONObject){
    JSON_Object* config = json_value_get_object(config_json);
    setup();
    mainloop(config);
  } else {
    puts("error reading/parsing config.json");
  }
}

void
setup() {
  admin_vm.state = mrb_open();
  admin_vm.owner = "admin";
  struct RClass *class_cextension = mrb_define_module(admin_vm.state, "Neuron");
  mrb_define_class_method(admin_vm.state, class_cextension, "go", my_c_method, MRB_ARGS_REQ(1));
}

void
mainloop(JSON_Object* config) {
  redisContext *redis;
  redisContext *redis_pub;
  redisReply *reply;
  redisReply *reply_pub;

  printf("redis connect %s subscribe %s\n", CONFIG("redis.host"), CONFIG("redis.channel"));
  redis = redisConnect(CONFIG("redis.host"), 6379);
  redis_pub = redisConnect(CONFIG("redis.host"), 6379);
  reply = (redisReply*)redisCommand(redis, "SUBSCRIBE %s", CONFIG("redis.channel"));
  while(redisGetReply(redis, (void**)&reply) == REDIS_OK) {
    // consume message
    const char* json_in = reply->element[2]->str;
    printf("<- %s\n", json_in);
    JSON_Value* jvalue = json_parse_string(json_in);
    if(json_value_get_type(jvalue) == JSONObject){
      JSON_Object* obj = json_value_get_object(jvalue);
      const char* code = json_object_get_string(obj, "code");

      if(code){
        const char* json_result;
        int i;
        for(i=0; i < machines_count; i++) {
          ruby_vm this_vm = machines[i];
          printf("machine %d/%s %p <- %s\n", i, this_vm.owner, &this_vm, code);
          json_result = eval_mruby_json(this_vm, code);
          printf("machine %d -> %s\n", i, json_result);
        }

        json_result = eval_mruby_json(admin_vm, code);
        printf("admin -> %s\n", json_result);

        reply_pub = (redisReply*)redisCommand(redis_pub, "publish %s %s", "neur0n", json_result);
        freeReplyObject(reply_pub);
      }
    }
    json_value_free(jvalue);
  }
  freeReplyObject(reply);
}

const char*
eval_mruby_json(ruby_vm vm, const char* code){
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

const char*
mruby_stringify_json(mrb_state* mrb, mrb_value val) {
  struct RClass* clazz = mrb_module_get(mrb, "JSON");
  mrb_value str = mrb_funcall(mrb, mrb_obj_value(clazz), "stringify", 1, val);
  return mrb_string_value_cstr(mrb, &str);
}

void
machines_add(const char* name){
  int idx = machines_count;
  machines_count = machines_count + 1;
  int new_size = sizeof(ruby_vm)*machines_count;
  printf("realloc %p size %ld * %d = %d\n", machines, sizeof(ruby_vm), machines_count, new_size);
  machines = (ruby_vm*)realloc(machines, new_size);
  printf("post realloc %p \n", machines);
  if(machines){
    printf("new machine #%d allocated for %s\n", machines_count, name);
    printf("machines %p. machines[%d] %p.\n", machines, idx, &machines[idx]);
    ruby_vm* new_vm = &machines[idx];
    new_vm->state = mrb_open();
    new_vm->owner = name;
    printf("new machine #%d allocated for %s @ %p\n", machines_count, name, &new_vm);
  }
}

static mrb_value
my_c_method(mrb_state *mrb, mrb_value self) {
  mrb_value x;
  mrb_get_args(mrb, "S", &x);

  printf("Neuron::go adding machine: %s\n", mrb_string_value_cstr(mrb, &x));
  machines_add(mrb_string_value_cstr(mrb, &x));
  return x;
}
