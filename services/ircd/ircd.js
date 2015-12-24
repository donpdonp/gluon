// node
var net = require("net")
var uuid = require('node-uuid')

// npm
var redisLib = require("redis"),
    redisSub = redisLib.createClient(),
    redisPub = redisLib.createClient()

var pubsub_channel = 'gluon'
var my_uuid = uuid.v4()

// local
var sessions = require('./lib/sessions')
var irc = require('./lib/irc')(redis_pub)

redisSub.on("subscribe", function (channel, count) {
  console.log("redis subscribe "+channel)
})

redisSub.on("message", function (channel, message) {
  console.log("<redis", message);
  try {
    var payload = JSON.parse(message)
    if(payload.method && payload.method.match(/^irc\./)) {
      dispatch(payload)
    }
  } catch(err) {
      console.log("redis json err %s", err);
  }
})

redisSub.subscribe(pubsub_channel)

function dispatch(payload) {
  // manage irc sessions
  var cmd = payload.method.split('.')[1]
  if(cmd == 'connect') {
    var session = sessions.generate(payload.params.server, payload.params.nick,
                                    payload.params.nick, payload.id)
    start(session)
  }
  if(cmd == 'list') {
    var session_list = sessions.list()
    console.log("irc sessions:", session_list)
    redis_pub({id: payload.id, result: session_list})
  }
  if(cmd == 'disconnect') {
    var session = sessions.get(payload.params.irc_session_id)
    if(session) {
      irc.disconnect(session)
      redis_pub({id: payload.id, result: "irc session "+session.id+" disconnected."})
    } else {
      console.log('disconnect: bad irc session id', payload.params.irc_session_id)
    }
  }
  if(cmd == 'join') {
    var session = sessions.get(payload.params.irc_session_id)
    if(session) {
      irc.join(session, payload.params.channel)
    } else {
      console.log('join: bad irc session id', payload.params.irc_session_id)
    }
  }
  if(cmd == 'privmsg') {
    if(!payload.params.nick) {
      var session
      if(payload.params.channel) {
        session = sessions.search(payload.params.channel)
      }
      if(payload.params.irc_session_id) {
        session = sessions.get(payload.params.irc_session_id)
      }
      if(session) {
        irc.privmsg(session, payload.params.channel, ':'+payload.params.message)
      } else {
        console.log('privmsg: bad irc session id', payload.params.irc_session_id)
      }
    }
  }
}

function redis_pub(msg){
  msg["from"] = my_uuid
  msg["id"] = uuid.v4()
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

setInterval(sessionsCheck, 60*1000)

function sessionsCheck(){
  sessions.list().forEach(function(session){
    if(session.state == 'error') {
      console.log('!! re-starting session', session.id)
      start(session)
    }
  })
}

function start(session) {
  console.log('ircd', 'session', '#'+session.id, 'start', session.hostname)
  irc.connect(session, new net.Socket(), function(err){restart(session, err)})
}
