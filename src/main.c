#include <string.h>
#include <hiredis/hiredis.h>

#include "mruby.h"
#include "mruby/compile.h"
#include "mruby/string.h"

const char* do_mruby_json(char*);
const char* stringify_json(mrb_state* mrb, mrb_value val);

int
main() {
  redisContext *redis;
  redisContext *redis_pub;
  redisReply *reply;
  redisReply *reply_pub;

  redis = redisConnect("127.0.0.1", 6379);
  redis_pub = redisConnect("127.0.0.1", 6379);
  puts("Subscribe neur0n");
  reply = redisCommand(redis, "SUBSCRIBE %s", "neur0n");
  while(redisGetReply(redis, (void**)&reply) == REDIS_OK) {
    // consume message
    char* code = reply->element[2]->str;
    puts(code);
    const char* json;
    json = do_mruby_json(code);

    if(json){
      puts(json);
    } else {
      puts("bad code");
    }
    reply_pub = redisCommand(redis_pub, "publish %s %s", "neur0n", json);
    freeReplyObject(reply_pub);
  }
  freeReplyObject(reply);
}

const char*
do_mruby_json(char* code){
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
    return;
  }

  struct RProc* proc;
  proc = mrb_generate_code(state, parser_state);
  if (proc == NULL) {
    fputs("codegen error\n", stderr);
    return;
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
  const char* json = json = stringify_json(state, result);
  mrb_close(state);
  return json;
}

const char*
stringify_json(mrb_state* mrb, mrb_value val) {
  struct RClass* clazz = mrb_module_get(mrb, "JSON");
  mrb_value str = mrb_funcall(mrb, mrb_obj_value(clazz), "stringify", 1, val);
  return mrb_string_value_cstr(mrb, &str);
}
