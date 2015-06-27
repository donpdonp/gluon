#include <stdlib.h>
#include <string.h>
#include <curl/curl.h>

#include "main.h"
#define CONFIG(key) json_object_dotget_string(config, key)

ruby_vm* machines = NULL;
int machines_count = 0;
int admin_vm_idx;
redisContext *redis_pub;

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
  struct RClass *class_cextension = mrb_module_get(admin_vm->state, "Neur0n");
  mrb_define_class_method(admin_vm->state, class_cextension, "machine_add", my_machine_add, MRB_ARGS_REQ(1));
  mrb_define_class_method(admin_vm->state, class_cextension, "machine_get", my_machine_get, MRB_ARGS_REQ(1));
  mrb_define_class_method(admin_vm->state, class_cextension, "machine_list", my_machine_list, MRB_ARGS_NONE());
  mrb_define_class_method(admin_vm->state, class_cextension, "machine_eval", my_machine_eval, MRB_ARGS_REQ(2));
  mrb_define_class_method(admin_vm->state, class_cextension, "http_get", my_http_get, MRB_ARGS_REQ(1));
  mruby_parse_file(*admin_vm, "admin.rb");
}

void
mainloop(JSON_Object* config) {
  redisContext *redis_sub;
  redisReply *reply;

  printf("redis: connect to %s. subscribe to %s.\n", CONFIG("redis.host"), CONFIG("redis.channel"));
  redis_sub = redisConnect(CONFIG("redis.host"), 6379);
  redis_pub = redisConnect(CONFIG("redis.host"), 6379);
  reply = (redisReply*)redisCommand(redis_sub, "SUBSCRIBE %s", CONFIG("redis.channel"));
  while(redisGetReply(redis_sub, (void**)&reply) == REDIS_OK) {
    // consume message
    const char* json_in = reply->element[2]->str;
    ruby_vm admin_vm = machines[admin_vm_idx];
    JSON_Value *input_json = json_parse_string(json_in);
    const char* id = json_object_get_string(json_value_get_object(input_json), "id");
    mrb_value json_obj = mruby_json_parse(admin_vm, json_in);
    printf("#####\n");
    printf("<- %s (mrb type %d)\n", json_in, json_obj.tt);

    if(json_obj.tt == MRB_TT_HASH){
      int local_count = machines_count;
      for(int i=0; i < local_count; i++) {
        ruby_vm this_vm = machines[i];
        printf("    machine %d/%s -> Neur0n::dispatch\n", i, this_vm.owner);
        mrb_value result = mruby_dispatch(this_vm, json_obj);
        printf("    machine %d/%s -> type %d\n", i, this_vm.owner, result.tt);

        if (this_vm.state->exc) {
          mrb_value errstr;
          errstr = mrb_funcall(this_vm.state, mrb_obj_value(this_vm.state->exc), "inspect", 0);
          printf("    machine %d/%s -> VM EXCEPTION\n", i, this_vm.owner);
          fwrite(RSTRING_PTR(errstr), RSTRING_LEN(errstr), 1, stdout);
          putc('\n', stdout);
          this_vm.state->exc = 0;
        } else {
          if(result.tt == MRB_TT_HASH){
            const char* json = mruby_stringify_json(this_vm, result);
            JSON_Value *resp_json = json_value_init_object();
            json_object_set_string(json_value_get_object(resp_json), "id", id);
            JSON_Value *payload_json = json_parse_string(json);
            json_object_set_value(json_value_get_object(resp_json), "result", payload_json);
            const char* safe_json = json_serialize_to_string(resp_json);
            json_value_free(resp_json);
            printf("    machine %d/%s -> %s\n", i, this_vm.owner, safe_json);
            send_result(safe_json);
          }
        }
      }
    }
    json_value_free(input_json);
    freeReplyObject(reply);
  }
}

void
send_result(const char* json) {
  redisReply *reply_pub;
  reply_pub = (redisReply*)redisCommand(redis_pub, "publish %s %s", "neur0n", json);
  if(reply_pub == NULL) {
    printf("Warning: reply_pub is null\n");
    if(redis_pub->err) {
        printf("Redis post-error: %s\n", redis_pub->errstr);
    }
  } else {
    freeReplyObject(reply_pub);
  }
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
    struct RClass *class_cextension = mrb_define_module(new_vm->state, "Neur0n");
    mrb_define_class_method(new_vm->state, class_cextension, "emit", my_emit, MRB_ARGS_REQ(1));
    printf("new machine #%d allocated for %s\n", machines_count-1, name);
    return idx;
  }
}

int
machines_find(const char* name){
  for(int i=0; i < machines_count; i++) {
    if(strcmp(name, machines[i].owner) == 0) {
      return i;
    }
  }
  return -1;
}

static mrb_value
my_machine_add(mrb_state *mrb, mrb_value self) {
  mrb_value name;
  mrb_get_args(mrb, "S", &name);

  const char* mname = mrb_string_value_cstr(mrb, &name);
  printf("my_machine_add %s\n", mname);
  int midx = machines_find(mname);
  if(midx == -1) {
    midx = machines_add(mname);
    return mrb_fixnum_value(midx);
  } else {
    printf("my_machine_add %s already exists\n", mname);
  }
}

mrb_value
machine_get_as_ruby(mrb_state *mrb, int i) {
  mrb_value hash;
  hash = mrb_hash_new(mrb);
  ruby_vm machine = machines[i];
  mrb_value key = mrb_str_new_cstr(mrb, "name");
  mrb_value value = mrb_str_new_cstr(mrb, machine.owner);
  mrb_hash_set(mrb, hash, key, value);
  return hash;
}

static mrb_value
my_machine_get(mrb_state *mrb, mrb_value self) {
  mrb_value machine;
  // fix id/owner mess
  return machine;
}

static mrb_value
my_machine_list(mrb_state *mrb, mrb_value self) {
  mrb_value list;
  list = mrb_ary_new(mrb);
  for(int i=0; i < machines_count; i++) {
    mrb_value machine;
    machine = machine_get_as_ruby(mrb, i);
    mrb_ary_push(mrb, list, machine);
  }
  return list;
}

static mrb_value
my_machine_eval(mrb_state *mrb, mrb_value self) {
  mrb_value name;
  mrb_value rcode;
  mrb_get_args(mrb, "SS", &name, &rcode);

  const char* machine_name = mrb_string_value_cstr(mrb, &name);
  const char* code = mrb_string_value_cstr(mrb, &rcode);
  printf("my_machine_eval %s\n", machine_name);
  int midx = machines_find(machine_name);
  if(midx){
    ruby_vm name_vm = machines[midx];
    mruby_eval(name_vm, code);
  } else {
      printf("my_machine_eval %s does not exist\n", machine_name);
  }
  return name;
}

static mrb_value
my_emit(mrb_state *mrb, mrb_value self) {
//  mrb_value vm_id;
//  mrb_get_args(mrb, "S", &vm_id);

  mrb_value msg;
  mrb_get_args(mrb, "o", &msg);

  struct RClass* clazz = mrb_module_get(mrb, "JSON");
  mrb_value str = mrb_funcall(mrb, mrb_obj_value(clazz), "stringify", 1, msg);
  const char* json = mrb_string_value_cstr(mrb, &str);
  send_result(json);
  printf("my_emit %s", json);

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
    struct CurlStr body;
    body.memory = malloc(1);  /* will be grown as needed by the realloc above */
    body.size = 0;

    curl_easy_setopt(curl, CURLOPT_URL, url_c);
    curl_easy_setopt(curl, CURLOPT_FOLLOWLOCATION, 1L);
    curl_easy_setopt(curl, CURLOPT_SSL_VERIFYPEER, 0L);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, curl_on_page);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &body);
    curl_easy_setopt(curl, CURLOPT_USERAGENT, "gluon curl");
    res = curl_easy_perform(curl);
    if(res == CURLE_OK) {
      mrb_value rbody = mrb_str_new(mrb, body.memory, body.size);
      // ruby str copies value
      free(body.memory);
      return rbody;
    } else {
      printf("curl not ok\n");
    }
    curl_easy_cleanup(curl);
  }
}

size_t
curl_on_page(char *ptr, size_t size, size_t nmemb, void *userdata) {
  size_t realsize = size * nmemb;
  struct CurlStr *mem = (struct CurlStr *)userdata;

  mem->memory = realloc(mem->memory, mem->size + realsize + 1);
  if(mem->memory == NULL) {
    /* out of memory! */
    printf("not enough memory (realloc returned NULL)\n");
    return 0;
  }

  memcpy(&(mem->memory[mem->size]), ptr, realsize);
  mem->size += realsize;
  mem->memory[mem->size] = 0;

  return realsize;
}
