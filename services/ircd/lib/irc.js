// node
var fs = require('fs')
// npm
var IrcSocket = require('irc-socket')

module.exports = function(publish){
  var o = {}
  var channels = {}
  var logfiles = {}

  o.connect = function(session, socket) {
    var opts = {
        server: session.hostname,
        port: 6667,
        nicknames: [session.nick],
        realname: session.name
      }
    logfiles[session.id] = fs.openSync('logs/'+session.hostname, 'a')
    log(session, [new Date().toISOString(), '!#!'])
    log(session, [new Date().toISOString(), '!#!', 'Session begin', session.id, "!#!"])
    var irc = channels[session.id] = IrcSocket(opts, socket);
    session.state = 'connecting'

    irc.once('ready', function () {
      console.log("irc ready")
    })

    irc.on('data', function (message) {
      //log(session, [new Date().toISOString(), message])
      var ircmsg = /^:([^ ]+) ([^ ]+) :?([^ ]+)( :?(.*))?/.exec(message)
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
      var msg = 'session ' + session.id + ' closed.'
      if(err) { msg = msg + ' has error: ' + err }
      log(session, ['ircd', msg])
      if(session.state === 'closing') {
        delete channels[session.id]
      } else {
        session.state = 'error'
      }
    })

    irc.connect().then(function(a){console.log('connect good', a)},
                       function(a){console.log('connect bad', a)})
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
      channels[session.id].raw(msg)
    }
  }

  o.disconnect = function(session) {
    console.log('removing session '+session.id)
    var channel = channels[session.id]
    if(channel) {
      channel.end()
      delete channels[session.id]
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
        var capstr = ircmsg[5].match(/(.*)\s+:[^:]+$/)
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

      case "JOIN":
        var reply = {id: session.msg_id,
                     result: {
                       irc_session_id: session.id,
                       channel: ircmsg[3]
                     }
                    }
        if(session.channels.indexOf(ircmsg[3]) == -1) {
          console.log('adding channel', ircmsg[3])
          session.channels.push(ircmsg[3])
        }
        publish(reply)
        log(session, [new Date().toISOString(), '!#!', session.id, '/join', ircmsg[3], "!#!"])
        break

      case "PART":
        if(session.channels.indexOf(ircmsg[3]) >= 0) {
          delete session.channels[ircmsg[3]]
        }
        log(session, [new Date().toISOString(), '!#!', session.id, '/part', ircmsg[3], "!#!"])
        break

      case "PRIVMSG":
        var from_nick = ircmsg[1].split('!')[0]
        console.log('from_nick', from_nick)
        if(from_nick != session.nick) {
          var reply = {method:'irc.privmsg',
                       params: {
                         irc_session_id: session.id,
                         nick: from_nick,
                         channel: ircmsg[3],
                         message: ircmsg[5] }
                      }
          publish(reply)
        }
        break
    }
  }

  return o
}
