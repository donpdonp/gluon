var redisLib = require("redis"),
    redisSub = redisLib.createClient(),
    redisPub = redisLib.createClient()

var IrcSocket = require('irc-socket');

var sessions = {}

function add_irc_session(server, nick, name) {
  var session = sessions[server+':'+nick] = { server: {caps: {}} }

  var irc = IrcSocket({
      server: server,
      port: 6667,
      nickname: nick,
      realname: name
      });

  irc.once('ready', function () {
    console.log("irc connected")
  })

  irc.on('data', function (message) {
    var ircmsg = /^:[^ ]+ (\d+) [^ ]+ (.*)/.exec(message)
    if(ircmsg && ircmsg[1] == "005") {
      var capstr = ircmsg[2].match(/(.*)\s+:[^:]+$/)
      var capabilities = split005(session.server.caps, capstr[1])
    }
    console.log(session.server.caps)
  })

  irc.connect();
}

redisSub.on("subscribe", function (channel, count) {
  console.log("redis subscribe "+channel)
})

redisSub.on("message", function (channel, message) {
  var payload = JSON.parse(message)
  console.log("redis<", channel, payload);
  if(payload.type && payload.type.substr(0,5) == 'irc.') { irc_dispatch(payload) }
})

redisSub.subscribe("neur0n")

function split005(scaps, capstr) {
  var caps = capstr.split(' ')
  for(var idx in caps) {
    var kv = caps[idx].split('=')
    if(kv[1]) {
      var vs = kv[1].split(',')
      if(vs.length > 1) { kv[1] = vs}
      scaps[kv[0]] = kv[1]
    }
  }
}

function irc_dispatch(payload) {
  // manage irc sessions
  var cmd = payload.type.split('.')[1]
  console.log('irc command', cmd)
}