var redisLib = require("redis"),
    redisSub = redisLib.createClient(),
    redisPub = redisLib.createClient()

var IrcSocket = require('irc-socket');
var irc = IrcSocket({
    server: 'irc.freenode.net',
    port: 6667,
    nickname: 'zr0bo',
    realname: 'Node Simple Socket'
    });

irc.once('ready', function () {
  console.log("connected")
  irc.end();
})

irc.connect();

redisSub.on("message", function (channel, message) {
  console.log("redis channel " + channel + ": " + message);
})

redisSub.subscribe("neur0n")
