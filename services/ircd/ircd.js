// node
var net = require("net")

// npm
var redisLib = require("redis"),
    redisSub = redisLib.createClient(),
    redisPub = redisLib.createClient()

// local
var sessions = require('./lib/sessions')
var irc = require('./lib/irc')(redis_pub)

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
    start(session)
  }
  if(cmd == 'list') {
    var session_list = sessions.list()
    console.log("irc sessions:", session_list)
    redis_pub({id: payload.id, result: session_list})
  }
  if(cmd == 'join') {
    var session = sessions.get(payload.irc_session_id)
    if(session) {
      irc.join(session, payload.channel)
    } else {
      console.log('join: bad irc session id', payload.irc_session_id)
    }
  }
  if(cmd == 'privmsg') {
    if(!payload.nick) {
      var session = sessions.get(payload.irc_session_id)
      if(session) {
        irc.privmsg(session, payload.channel, ':'+payload.message)
      } else {
        console.log('privmsg: bad irc session id', payload.irc_session_id)
      }
    }
  }
}

function redis_pub(msg){
  var json = JSON.stringify(msg)
  console.log('redis>', json)
  redisPub.publish('neur0n', json)
  if(msg.type === 'irc.connected') {
    var session = sessions.get(msg.irc_session_id)
    session.channels.forEach(function(channel){
      console.log('!! rejoining ', channel)
      irc.join(session, channel)
    })
  }
}

setInterval(sessionsCheck, 60*1000)

function sessionsCheck(){
  sessions.list().forEach(function(session){
    console.log('checking session', session.id, session.state)
    if(session.state == 'error') {
      console.log('!! re-starting session', session.id)
      start(session)
    }
  })
}

function start(session) {
  irc.connect(session, new net.Socket(), function(err){restart(session, err)})
}
