#include <stdlib.h>
#include <string.h>
#include <hiredis/hiredis.h>
#include <curl/curl.h>

#include "main.h"
#define CONFIG(key) json_object_dotget_string(config, key)

ruby_vm* machines = NULL;
int machines_count = 0;
int admin_vm_idx;

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
  admin_vm_idx = machines_add("admin");
  ruby_vm* admin_vm = &machines[admin_vm_idx];
  struct RClass *class_cextension = mrb_define_module(admin_vm->state, "Neur0n");
  mrb_define_class_method(admin_vm->state, class_cextension, "machine_add", my_machine_add, MRB_ARGS_REQ(1));
  mrb_define_class_method(admin_vm->state, class_cextension, "machine_eval", my_machine_eval, MRB_ARGS_REQ(1));
  mrb_define_class_method(admin_vm->state, class_cextension, "http_get", my_http_get, MRB_ARGS_REQ(1));
  mrb_define_class_method(admin_vm->state, class_cextension, "dispatch", my_dispatch, MRB_ARGS_REQ(1));
  mruby_parse_file(*admin_vm, "admin.rb");
}

void
mainloop(JSON_Object* config) {
  redisContext *redis_sub;
  redisContext *redis_pub;
  redisReply *reply;
  redisReply *reply_pub;

  printf("redis: connect to %s. subscribe to %s.\n", CONFIG("redis.host"), CONFIG("redis.channel"));
  redis_sub = redisConnect(CONFIG("redis.host"), 6379);
  redis_pub = redisConnect(CONFIG("redis.host"), 6379);
  reply = (redisReply*)redisCommand(redis_sub, "SUBSCRIBE %s", CONFIG("redis.channel"));
  while(redisGetReply(redis_sub, (void**)&reply) == REDIS_OK) {
    // consume message
    const char* json_in = reply->element[2]->str;
    printf("<-(raw) %s \n", json_in);
    ruby_vm admin_vm = machines[admin_vm_idx];
    mrb_value json_obj = mruby_json_parse(admin_vm, json_in);
    printf("<- %s (mrb type %d)\n", json_in, json_obj.tt);

    if(json_obj.tt == MRB_TT_HASH){
      int local_count = machines_count;
      for(int i=0; i < local_count; i++) {
        ruby_vm this_vm = machines[i];
        printf("    machine %d/%s -> Neur0n::dispatch\n", i, this_vm.owner);
        mrb_value result = mruby_dispatch(this_vm, json_obj);
        printf("    machine %d/%s -> return type %d\n", i, this_vm.owner, result.tt);

        if(result.tt == MRB_TT_HASH){
          const char* json = mruby_stringify_json(admin_vm, result);
          printf("    machine %d/%s -> publish json %s\n", i, this_vm.owner, json);
          //reply_pub = (redisReply*)redisCommand(redis_pub, "publish %s %s", "neur0n", "{}");
          //freeReplyObject(reply_pub);
        } else {
        }
      }
    }
  }
  freeReplyObject(reply);
}


int
machines_add(const char* name){
  int idx = machines_count;
  machines_count = machines_count + 1;
  int new_size = sizeof(ruby_vm)*machines_count;
  machines = (ruby_vm*)realloc(machines, new_size);
  if(machines){
    ruby_vm* new_vm = &machines[idx];
    new_vm->state = mrb_open();
    new_vm->owner = name;
    printf("new machine #%d allocated for %s\n", machines_count-1, name);
    return idx;
  }
}

static mrb_value
my_machine_add(mrb_state *mrb, mrb_value self) {
  mrb_value x;
  mrb_get_args(mrb, "S", &x);

  printf("Neuron::machine_add %s\n", mrb_string_value_cstr(mrb, &x));
  machines_add(mrb_string_value_cstr(mrb, &x));
  return x;
}

static mrb_value
my_machine_eval(mrb_state *mrb, mrb_value self) {
  mrb_value x;
  mrb_get_args(mrb, "S", &x);

  const char* machine_name = mrb_string_value_cstr(mrb, &x);
  printf("my_machine_eval %s\n", machine_name);
  for(int i=0; i < machines_count; i++){
    if(strcmp(machines[i].owner, machine_name) == 0){
      ruby_vm name_vm = machines[i];
      printf("my_machine_eval %s found. sending test i live\n", machine_name);
      mruby_eval(name_vm, "module Neur0n; def self.dispatch(msg); puts 'i live'; {:n=>1}; end;end");
    }
  }
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

static mrb_value
my_db_get(mrb_state *mrb, mrb_value self) {
  mrb_value key;
  mrb_get_args(mrb, "S", &key);

  return key;
};

static mrb_value
my_db_del(mrb_state *mrb, mrb_value self) {
  mrb_value key;
  mrb_get_args(mrb, "S", &key);

  return key;
};

static mrb_value
my_db_set(mrb_state *mrb, mrb_value self) {
  mrb_value key;
  mrb_get_args(mrb, "S", &key);
  mrb_value value;
  mrb_get_args(mrb, "S", &value);

  return value;
};

static mrb_value
my_http_get(mrb_state *mrb, mrb_value self) {
  CURL* curl;
  mrb_value url;
  mrb_get_args(mrb, "S", &url);

  const char* url_c  = mrb_string_value_cstr(mrb, &url);
  printf("my_http_get %s\n", url_c);

  curl = curl_easy_init();
  if(curl){
    CURLcode res;

    curl_easy_setopt(curl, CURLOPT_URL, url_c);
    curl_easy_setopt(curl, CURLOPT_FOLLOWLOCATION, 1L);
    curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, 0L);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, curl_on_page);
    res = curl_easy_perform(curl);
    if(res == CURLE_OK) {
      printf("curl ok\n");

    } else {
      printf("curl not ok\n");
    }
    curl_easy_cleanup(curl);
  }
  return url;
}

size_t
curl_on_page(char *ptr, size_t size, size_t nmemb, void *userdata){
  printf("curl on_page %d\n", (int)(size*nmemb));
  printf("%s\n", ptr);
  return size*nmemb;
}
