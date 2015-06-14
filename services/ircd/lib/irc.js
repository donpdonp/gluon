// npm
var IrcSocket = require('irc-socket')
var NetSocket = require("net").Socket

module.exports = (function(){
  var o = {}

  o.connect = function(session) {
    session.irc.connect()
  }

  o.add = function(session) {
    var opts = {
        server: session.hostname,
        port: 6667,
        nicknames: [session.nick],
        realname: session.name
      }
    var netSocket = new NetSocket();
    var irc = session.irc = IrcSocket(opts, netSocket);

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

    irc.on('closed', function (message) {
      console.log(server, 'closed')
    })

    irc.on('error', function(err){
    /* input: ':zrobo!~user
     * @75-175-104-74.ptld.qwest.net QUIT :Ping timeout: 240 seconds' ]
     * <irc ERROR :Closing Link: 75-175-104-74.ptld.qwest.net (Ping timeout: 240 seconds)
     */
      console.log(err)
      irc.end()
  /*{ [Error: read ETIMEDOUT] code: 'ETIMEDOUT', errno: 'ETIMEDOUT', syscall: 'read' }
   */
    })
  }


  o.join = function(session, channel) {
    var cmd = "JOIN "+channel
    o.say(session, cmd)
  }

  o.privmsg = function(session, channel, message) {
    var cmd = "PRIVMSG "+channel+" "+message
    o.say(network, cmd)
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
