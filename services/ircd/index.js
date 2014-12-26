var redisLib = require("redis"),
    redisSub = redisLib.createClient(),
    redisPub = redisLib.createClient()

var IrcSocket = require('irc-socket');

var sessions = {}

var irc = IrcSocket({
    server: 'irc.freenode.net',
    port: 6667,
    nickname: 'zr0bo',
    realname: 'Node Simple Socket'
    });

irc.once('ready', function () {
  console.log("irc connected")
})

irc.on('data', function (message) {
  var ircmsg = /^:[^ ]+ (\d+) [^ ]+ (.*)/.exec(message)
  if(ircmsg && ircmsg[1] == "005") {
    var capstr = ircmsg[2].match(/(.*)\s+:[^:]+$/)
    var capabilities = split005(capstr[1])
  }
})

irc.connect();

redisSub.on("message", function (channel, message) {
  console.log("redis channel " + channel + ": " + message);
})

redisSub.subscribe("neur0n")

function split005(capstr) {
  console.log(capstr)
  var caps = capstr.split(' ')
  for(var idx in caps) {
    var kv = caps[idx].split('=')
    if(kv[1]) {
      var vs = kv[1].split(',')
      if(vs.length > 1) { kv[1] = vs}
    }
    console.log(kv)
  }
}
