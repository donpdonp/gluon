// node
var NetSocket = require("net").Socket

// npm
var redisLib = require("redis"),
    redisSub = redisLib.createClient(),
    redisPub = redisLib.createClient()

// local
var sessions = require('./lib/sessions')
var irc = require('./lib/irc')

redisSub.on("subscribe", function (channel, count) {
  console.log("redis subscribe "+channel)
})

redisSub.on("message", function (channel, message) {
  console.log("<redis", message);
  var payload = JSON.parse(message)
  if(payload.type && payload.type.match(/^irc\./)) {
    dispatch(payload)
  }
})

redisSub.subscribe("neur0n")

function dispatch(payload) {
  // manage irc sessions
  var cmd = payload.type.split('.')[1]
  if(cmd == 'connect') {
    var session = sessions.generate(payload.server, payload.nick,
                                    payload.nick, redis_pub)
    irc.connect(session, socketer(session))
  }
  if(cmd == 'list') {
    console.log("irc sessions:", sessions.list())
  }
  if(cmd == 'join') {
    irc.join(sessions.get(payload.network), payload.channel)
  }
  if(cmd == 'privmsg') {
    if(!payload.nick) {
      irc.privmsg(sessions.get(payload.network), payload.channel, ':'+payload.message)
    }
  }
}

function redis_pub(msg){
  var json = JSON.stringify(msg)
  console.log('redis>', json)
  redisPub.publish('neur0n', json)
}

function socketer(session) {
    var socket = new NetSocket()
    socket.on('closed', function (message) {
      console.log(server, 'closed')
    })

    socket.on('error', function(err){
      console.log('sock err', err)
      // what
      //irc.connect(session, socketer(session))
      return true
    })

    return socket
}