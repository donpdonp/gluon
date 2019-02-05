// node
var net = require("net")
var uuid = require('node-uuid')

// npm
var redisLib = require("redis"),
    redisSub = redisLib.createClient(),
    redisPub = redisLib.createClient()

var pubsub_channel = 'gluon'
var my_uuid = uuid.v4()

redisSub.on("subscribe", function (channel, count) {
  console.log("redis subscribe "+channel)
})

redisSub.on("message", function (channel, message) {
  try {
    var payload = JSON.parse(message)
    if(payload.method && payload.method.match(/^db\./)) {
      dispatch(payload)
    }
  } catch(err) {
      console.log("redis json err %s", err);
  }
})

redisSub.subscribe(pubsub_channel)

function redis_pub(msg){
  msg["from"] = my_uuid
  msg["id"] = msg["id"] || uuid.v4()
  var json = JSON.stringify(msg)
  //console.log('->', json)
  redisPub.publish(pubsub_channel, json)
  if(msg.method === 'irc.connected') {
    var session = sessions.get(msg.params.irc_session_id)
    session.channels.forEach(function(channel){
      console.log('!! rejoining ', channel)
      irc.join(session, channel)
    })
  }
}

function dispatch(payload) {
  var cmd = payload.method.split('.')[1]
  if(cmd == 'get') {
    var value = redisPub.hget(payload.params.group, payload.params.key, function(err, value) {
      console.log('HGET', payload.params.group, payload.params.key, '->', err || value)
      redis_pub({id: payload.id, result: value})
    })
  }
  if(cmd == 'set') {
    var value = redisPub.hset(payload.params.group, payload.params.key, payload.params.value, function(err, value) {
      console.log('HSET', payload.params.group, payload.params.key, payload.params.value, '->', err || value)
      redis_pub({id: payload.id, result: true})
    })
  }
  if(cmd == 'del') {
    var value = redisPub.hdel(payload.params.group, payload.params.key, function(err, value) {
      console.log('HDEL', payload.params.group, payload.params.key, '->', err || value)
      redis_pub({id: payload.id, result: value})
    })
  }
  if(cmd == 'len') {
    var value = redisPub.hlen(payload.params.group, function(err, value) {
      console.log('HLEN', payload.params.group, '->', err || value)
      redis_pub({id: payload.id, result: value})
    })
  }
  if(cmd == 'scan') {
    var value = redisPub.hscan(payload.params.group, payload.params.cursor,
                               'MATCH', payload.params.match,  'COUNT', payload.params.count,
                               function(err, value) {
      console.log('HSCAN', payload.params.group, payload.params.cursor,
                  'MATCH', payload.params.match, 'COUNT', payload.params.count, '->', err || value)
      redis_pub({id: payload.id, result: value})
    })
  }
}

