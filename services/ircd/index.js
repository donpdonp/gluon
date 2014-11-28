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

