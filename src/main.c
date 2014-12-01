#include <stdlib.h>
#include <string.h>
#include <hiredis/hiredis.h>

#include "main.h"
#define CONFIG(key) json_object_dotget_string(config, key)

ruby_vm* machines = NULL;
int machines_count = 0;
ruby_vm* admin_vm;


int
main() {
  JSON_Value *config_json = json_parse_file("config.json");
  if(json_value_get_type(config_json) == JSONObject){
    JSON_Object* config = json_value_get_object(config_json);
    admin_setup();
    mainloop(config);
  } else {
    puts("error reading/parsing config.json");
  }
}

void
admin_setup() {
  admin_vm = machines_add("admin");
  struct RClass *class_cextension = mrb_define_module(admin_vm->state, "Neuron");
  mrb_define_class_method(admin_vm->state, class_cextension, "add_machine", my_c_method, MRB_ARGS_REQ(1));
  mrb_define_class_method(admin_vm->state, class_cextension, "dispatch", my_dispatch, MRB_ARGS_REQ(1));
  mruby_parse_file(*admin_vm, "admin.rb");
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
        json_result = mruby_eval(*admin_vm, code);
        printf("admin -> %s\n", json_result);
        int i;
        for(i=0; i < machines_count; i++) {
          ruby_vm this_vm = machines[i];
          printf("machine %d/%s %p <- %s\n", i, this_vm.owner, &this_vm, code);
          json_result = mruby_eval(this_vm, code);
          printf("machine %d -> %s\n", i, json_result);

          reply_pub = (redisReply*)redisCommand(redis_pub, "publish %s %s", "neur0n", json_result);
          freeReplyObject(reply_pub);
        }
      }
    }
    json_value_free(jvalue);
  }
  freeReplyObject(reply);
}


ruby_vm*
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
    return &machines[idx];
  }
}

static mrb_value
my_c_method(mrb_state *mrb, mrb_value self) {
  mrb_value x;
  mrb_get_args(mrb, "S", &x);

  printf("Neuron::add_machine %s\n", mrb_string_value_cstr(mrb, &x));
  machines_add(mrb_string_value_cstr(mrb, &x));
  return x;
}

static mrb_value
my_dispatch(mrb_state *mrb, mrb_value self) {
//  mrb_value vm_id;
//  mrb_get_args(mrb, "S", &vm_id);

  mrb_value msg;
  mrb_get_args(mrb, "o", &msg);

  return msg;
};