#include <string.h>
#include <hiredis/hiredis.h>

#include "mruby.h"
#include "mruby/compile.h"
#include "mruby/string.h"
#include "mruby-json/src/parson.h"

#define CONFIG(key) json_object_dotget_string(config, key)

const char* do_mruby_json(const char*);
const char* mruby_stringify_json(mrb_state* mrb, mrb_value val);
void mainloop(JSON_Object* config);

int
main() {
  JSON_Value *config_json;
  config_json = json_parse_file("config.json");
  if(json_value_get_type(config_json) == JSONObject){
    JSON_Object* config = json_value_get_object(config_json);
    mainloop(config);
  } else {
    puts("error reading/parsing config.json");
  }
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
  reply = redisCommand(redis, "SUBSCRIBE %s", CONFIG("redis.channel"));
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
        json_result = do_mruby_json(code);

        const char* answer;
        if(json_result){
          printf("-> %s\n", json_result);
          answer = json_result;
        } else {
          puts("bad code");
          answer = "{\"error\":\"bad code\"}";
        }
        reply_pub = redisCommand(redis_pub, "publish %s %s", "neur0n", answer);
        freeReplyObject(reply_pub);
      }
    }
    json_value_free(jvalue);
  }
  freeReplyObject(reply);
}

const char*
do_mruby_json(const char* code){
  mrb_state* state;
  state = mrb_open();

  mrbc_context* context;
  context = mrbc_context_new(state);

  struct mrb_parser_state* parser_state;
  parser_state = mrb_parse_string(state, code, context);
  if (parser_state == NULL) {
    fputs("parse error\n", stderr);
    return;
  }
  if (0 < parser_state->nerr) {
    printf("line %d: %s\n", parser_state->error_buffer[0].lineno,
                            parser_state->error_buffer[0].message);
    return NULL;
  }

  struct RProc* proc;
  proc = mrb_generate_code(state, parser_state);
  if (proc == NULL) {
    fputs("codegen error\n", stderr);
    return NULL;
  }
  mrb_parser_free(parser_state);

  mrb_value root_object;
  root_object = mrb_top_self(state);

  mrb_value result;
  result = mrb_run(state, proc, root_object);
  if(result.tt == MRB_TT_EXCEPTION){
    fputs("EXCEPTION\n", stderr);
    return NULL;
  }
  const char* json = mruby_stringify_json(state, result);
  mrb_close(state);
  return json;
}

const char*
mruby_stringify_json(mrb_state* mrb, mrb_value val) {
  struct RClass* clazz = mrb_module_get(mrb, "JSON");
  mrb_value str = mrb_funcall(mrb, mrb_obj_value(clazz), "stringify", 1, val);
  return mrb_string_value_cstr(mrb, &str);
}
