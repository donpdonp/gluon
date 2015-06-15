// npm
var IrcSocket = require('irc-socket')

module.exports = (function(){
  var o = {}

  o.connect = function(session, socket) {
    var opts = {
        server: session.hostname,
        port: 6667,
        nicknames: [session.nick],
        realname: session.name
      }
    var irc = session.irc = IrcSocket(opts, socket);

    irc.once('ready', function () {
      console.log("irc connected")
    })

    irc.on('data', function (message) {
      console.log('<irc', message)
      var ircmsg = /^:([^ ]+) ([^ ]+) ([^ ]+)( :?(.*))?/.exec(message)
      if(ircmsg) {
        handle_irc_msg(session, ircmsg)
      }
    })

    irc.connect()
  }


  o.join = function(session, channel) {
    var cmd = "JOIN "+channel
    o.say(session, cmd)
  }

  o.privmsg = function(session, channel, message) {
    var cmd = "PRIVMSG "+channel+" "+message
    o.say(session, cmd)
  }

  o.say = function(session, msg) {
    console.log('irc>', msg)
    if(session) {
      session.irc.raw(msg)
    }
  }

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

  function handle_irc_msg(session, ircmsg){
      if(ircmsg[2] == "001") {
        console.log('irc 001 greeting. nick confirmed as', ircmsg[3])
        session['nick'] = ircmsg[3]
      }

      if(ircmsg[2] == "005") {
        var capstr = ircmsg[5].match(/(.*)\s+:[^:]+$/)
        var capabilities = split005(session.server.caps, capstr[1])
      }
      if(ircmsg[2] == "251") {
        console.log('irc network detect', session.server.caps.network)
        session.connected(session.server.caps.network, session)
        var reply = {type:'irc.connected', network: session.server.caps.network, nick: session.nick}
        session.publish(reply)
      }
      if(ircmsg[2] == "JOIN") {
        var reply = {type:'irc.joined', network: session.server.caps.network, channel: ircmsg[3]}
        session.publish(reply)
      }
      if(ircmsg[2] == "PRIVMSG") {
        var from_nick = ircmsg[1].split('!')[0]
        console.log('from_nick', from_nick)
        if(from_nick != session.nick) {
          var reply = {type:'irc.privmsg',
                       network: session.server.caps.network,
                       nick: from_nick,
                       channel: ircmsg[3],
                       message: ircmsg[5] }
          session.publish(reply)
        }
      }
  }

  return o
})()
