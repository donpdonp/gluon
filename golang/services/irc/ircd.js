// node
var net = require("net")

// npm
var nano = require('nanomsg');
var pub_chan = nano.socket('bus');

// local
var sessions = require('./lib/sessions')
var irc = require('./lib/irc')(function(json){pub_chan.send(json)})

var addr = 'tcp://127.0.0.1:40899'
pub_chan.connect(addr);
console.log('connected ', addr)

pub_chan.on('data', function (buf) {
  var message = String(buf);
  console.log("<nano", message);
  try {
    var payload = JSON.parse(message)
    if(payload.method && payload.method.match(/^irc\./)) {
      dispatch(payload)
    }
  } catch(err) {
      console.log("json err %s", err);
  }
})

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
    pub({id: payload.id, result: session_list})
  }
  if(cmd == 'disconnect') {
    var session = sessions.get(payload.params.irc_session_id)
    if(session) {
      irc.disconnect(session)
      pub({id: payload.id, result: "irc session "+session.id+" disconnected."})
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
      var session = sessions.get(payload.params.irc_session_id)
      if(session) {
        irc.privmsg(session, payload.params.channel, ':'+payload.params.message)
      } else {
        console.log('privmsg: bad irc session id', payload.params.irc_session_id)
      }
    }
  }
}

function pub(msg){
  var json = JSON.stringify(msg)
  console.log('nano>', json)
  pub_chan.send(json)
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
