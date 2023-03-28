// node
var fs = require('fs')
// npm
var IrcSocket = require('irc-socket')

// irc regex
var IrcMsgRegex = /^:([^ ]+) ([^ ]+) :?([^ ]+)( :?(.*))?/
var IrcExtraRegex = /(\S*)(\s+:?([^:]+))?$/

module.exports = function(publish){
  var o = {}
  var sockets = {}
  var logfiles = {}

  o.connect = function(session, socket) {
    var opts = {
        server: session.hostname,
        port: 6667,
        nicknames: [session.nick],
        realname: session.name
      }
    logfiles[session.id] = fs.openSync('logs/'+session.hostname, 'a')
    log(session, [])
    log(session, ['Session begin', "!#!"])
    var irc = sockets[session.id] = IrcSocket(opts, socket);
    session.state = 'connecting'

    irc.once('ready', function () {
      console.log(session.id, "irc ready")
    })

    irc.on('data', function (message) {
      //log(session, [new Date().toISOString(), message])
      var ircmsg = IrcMsgRegex.exec(message)
      if(ircmsg) {
        handle_irc_msg(session, ircmsg)
      }
    })

    irc.on('error', function(e) {
      log(session, [new Date().toISOString(), '!#!'])
      log(session, ['ircd', 'session', '#'+session.id, 'in error', e.code])
      session.state = 'error'
    })

    irc.on('close', function(err) {
      var msg = 'session ' + session.id + ' state closed. session was state '+session.state+'. setting to error for reconnect.' 
     //2023-02-09T03:07:23.697Z !#! vw6si ircd session vw6si closed. was state connected
      if(err) { msg = msg + ' error: ' + err }
      log(session, ['ircd', msg])
      session.state = 'error'
    })

    irc.connect().then(function(a){console.log(new Date(), 'connect good', a)},
                       function(a){console.log(new Date(), 'connect bad', a)})
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
    if(session) {
      console.log('irc-'+session.id+'-'+session.state+'>', msg)
      sockets[session.id].raw(msg)
    }
  }

  o.disconnect = function(session) {
    log(session, ['disconnect(). removing session'])
    var channel = sockets[session.id]
    if(channel) {
      channel.end()
      delete sockets[session.id]
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

  function log(session, words) {
    words.unshift('! '+session.id+' !')
    words.unshift(new Date().toISOString())
    var message = words.join(' ')+"\n"
    fs.writeSync(logfiles[session.id], message)
  }

  function rejoin(session) {
    console.log('rejoin', session.channels)
    session.channels.forEach(function(channel){
      o.join(session, channel)
    })
  }

  function handle_irc_msg(session, ircmsg){
    var command = ircmsg[2]
    switch(command) {
      case "001":
        console.log('irc 001 greeting. nick confirmed as', ircmsg[3])
        session['nick'] = ircmsg[3]
        session.state = 'connected'
        log(session, [new Date().toISOString(), '!#!', 'IRC connected', session.id, "!#!"])
        rejoin(session)
        break

      case "005":
        var capstr = ircmsg[5].match(IrcExtraRegex)
        var capabilities = split005(session.server.caps, capstr[1])
        break

      case "251":
        // 251 signals CAPS list is over
        session.network = session.server.caps.network
        var reply = {id: session.msg_id,
                     result: {
                       irc_session_id: session.id,
                       network: session.server.caps.network,
                       nick: session.nick
                     }
                    }
        publish(reply)
        break

//:daamien15!~daamien@116.104.86.89 JOIN #pdxtech
      case "JOIN":
        var reply = {id: session.msg_id,
                     result: {
                       irc_session_id: session.id,
                       channel: ircmsg[3]
                     }
                    }
        if(session.channels.indexOf(ircmsg[3]) == -1) {
          log(session, ['adding channel', ircmsg[3]])
          session.channels.push(ircmsg[3])
        }
        publish(reply)
        var user_parts = ircmsg[1].split('!')
        var ircbus = {method: "irc.join",
                      params: {irc_session_id: session.id,
                               user: {nick: user_parts[0], host: user_parts[1]},
                               channel: ircmsg[3]}}
        publish(ircbus)
        log(session, ['/join', ircmsg[3], "!#!"])
        break

//:Loqi!Loqi@2600:3c01::f03c:91ff:fef1:a349 KICK #pdxtech daamien15 :1116
      case "KICK":
        var user_parts = ircmsg[1].split('!')
        var channel = ircmsg[3]
        var extra = ircmsg[5].match(IrcExtraRegex)
        var ircbus = {method: "irc.kick",
                      params: {irc_session_id: session.id,
                               user: {nick: user_parts[0], host: user_parts[1]},
                               nick: extra[1],
                               reason: extra[3],
                               channel: ircmsg[3]}}
        publish(ircbus)
        break

      case "PART":
        if(session.channels.indexOf(ircmsg[3]) >= 0) {
          delete session.channels[ircmsg[3]]
        }
        log(session, ['/part', ircmsg[3], "!#!"])
        break

//:bkero!~bkero@osuosl/staff/bkero PRIVMSG #pdxtech :Only can be copied with a $38 device on Aliexpress.
      case "PRIVMSG":
        var from_nick = ircmsg[1].split('!')[0]
        log(session, [ircmsg])
        if(from_nick != session.nick) {
          var ircbus = {method:'irc.privmsg',
                        params: {
                          irc_session_id: session.id,
                          nick: from_nick,
                          channel: ircmsg[3],
                          message: ircmsg[5] }
                       }
          publish(ircbus)
        }
        break

//:Loqi!Loqi@2600:3c01::f03c:91ff:fef1:a349 MODE #pdxtech +v zz99
//:zrobo MODE zrobo :+Ri
      case "MODE":
        var user_parts = ircmsg[1].split('!')
        var extra = ircmsg[5].match(IrcExtraRegex)
        var ircbus = {method: "irc.mode",
                      params: {irc_session_id: session.id,
                               user: {nick: user_parts[0], host: user_parts[1]},
                               channel: ircmsg[3],
                               mode: extra[1],
                               nick: extra[3]
                             }}
        publish(ircbus)
          break

//:Guest21917!~Khisanth@sju13-4-88-161-81-249.fbx.proxad.net QUIT :Remote host closed the connection
    }
  }

  return o
}
