var redisLib = require("redis"),
    redisSub = redisLib.createClient(),
    redisPub = redisLib.createClient()

var IrcSocket = require('irc-socket');

var sessions = {}

function add_irc_session(server, nick, name) {
  var session = { server: {caps: {}} }

  var irc = session.irc = IrcSocket({
      server: server,
      port: 6667,
      nickname: nick,
      realname: name
      });

  irc.once('ready', function () {
    console.log("irc connected")
  })

  irc.on('data', function (message) {
    console.log(message)
    var ircmsg = /^:[^ ]+ ([^ ]+) ([^ ]+) (.*)/.exec(message)
    if(ircmsg) {
      if(ircmsg[1] == "005") {
        var capstr = ircmsg[3].match(/(.*)\s+:[^:]+$/)
        var capabilities = split005(session.server.caps, capstr[1])
      }
      if(ircmsg[1] == "251") {
        console.log('irc network detect', session.server.caps.network)
        sessions[session.server.caps.network] = session
        var reply = {type:'irc.connected', network: session.server.caps.network}
        redisPub.publish('neur0n', JSON.stringify(reply))
      }
      if(ircmsg[1] == "JOIN") {
        var reply = {type:'irc.joined', network: session.server.caps.network, channel: ircmsg[2]}
        redisPub.publish('neur0n', JSON.stringify(reply))
      }
    }
  })

  irc.on('close', function (message) {
    console.log(server, 'closed')
  })

  irc.connect();
}

redisSub.on("subscribe", function (channel, count) {
  console.log("redis subscribe "+channel)
})

redisSub.on("message", function (channel, message) {
  var payload = JSON.parse(message)
  console.log("redis<", channel, payload);
  if(payload.type && payload.type.match(/^irc\./)) { irc_dispatch(payload) }
})

redisSub.subscribe("neur0n")

function split005(scaps, capstr) {
  var caps = capstr.split(' ')
  for(var idx in caps) {
    var kv = caps[idx].split('=')
    if(kv[1]) {
      var vs = kv[1].split(',')
      if(vs.length > 1) { kv[1] = vs}
      scaps[kv[0].toLowerCase()] = kv[1]
    }
  }
}

function irc_dispatch(payload) {
  // manage irc sessions
  var cmd = payload.type.split('.')[1]
  console.log('irc command', cmd)
  if(cmd == 'connect') {
    add_irc_session(payload.server, payload.nick, payload.nick)
  }
  if(cmd == 'join') {
    irc_join(payload.network, payload.channel)
  }
  if(cmd == 'privmsg') {
    irc_privmsg(payload.network, payload.channel, payload.message)
  }
}

function irc_join(network, channel) {
  var cmd = "JOIN "+channel
  console.log(cmd)
  sessions[network].irc.raw(cmd)
}

function irc_privmsg(network, channel, message) {
  var cmd = "PRIVMSG "+channel+" "+message
  console.log(cmd)
  sessions[network].irc.raw(cmd)
}
