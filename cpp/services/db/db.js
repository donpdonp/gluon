// node
var net = require("net")

// npm
var redisLib = require("redis"),
    redisSub = redisLib.createClient(),
    redisPub = redisLib.createClient()

// local
var irc = require('./lib/irc')(redis_pub)

redisSub.on("subscribe", function (channel, count) {
  console.log("redis subscribe "+channel)
})

redisSub.on("message", function (channel, message) {
  console.log("db from redis", message);
  try {
    var payload = JSON.parse(message)
    if(payload.method && payload.method.match(/^db\./)) {
      dispatch(payload)
    }
  } catch(err) {
      console.log("redis json err %s", err);
  }
})

redisSub.subscribe("neur0n")

function redis_pub(msg){
  var json = JSON.stringify(msg)
  console.log('redis>', json)
  redisPub.publish('neur0n', json)
  if(msg.method === 'irc.connected') {
    var session = sessions.get(msg.params.irc_session_id)
    session.channels.forEach(function(channel){
      console.log('!! rejoining ', channel)
      irc.join(session, channel)
    })
  }
}

function dispatch(payload) {
  // manage irc sessions
  var cmd = payload.method.split('.')[1]
  if(cmd == 'get') {
    var session = sessions.generate(payload.params.server, payload.params.nick,
                                    payload.params.nick, payload.id)
    start(session)
  }
  if(cmd == 'set') {
    var session_list = sessions.list()
    console.log("irc sessions:", session_list)
    redis_pub({id: payload.id, result: session_list})
  }
}

