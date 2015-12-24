// node
var net = require("net")

// npm
var redisLib = require("redis"),
    redisSub = redisLib.createClient(),
    redisPub = redisLib.createClient()

var pubsub_channel = 'gluon'

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
  var json = JSON.stringify(msg)
  console.log('redis>', json)
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
    var value = redisPub.get(payload.params.key, function(err, value) {
      redis_pub({id: payload.id, result: value})
    })
  }
  if(cmd == 'set') {
    var value = redisPub.set(payload.params.key, payload.params.value, function(err, value) {
      redis_pub({id: payload.id, result: true})
    })
  }
  if(cmd == 'delete') {
    var value = redisPub.del(payload.params.key, function(err, value) {
      redis_pub({id: payload.id, result: value})
    })
  }
}

