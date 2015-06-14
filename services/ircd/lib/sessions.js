module.exports = (function(){
  var sessions = {}
  var o = {}

  o.generate = function(hostname, nick, name) {
    var session = { server: {caps: {}},
                    hostname: hostname,
                    nick: nick,
                    name: name }
    return session
  }

  o.add = function(network, session) {
    console.log('sessions.add', network)
    sessions[network] = session
  }

  o.get = function(network) {
    return sessions[network]
  }

  return o
})()

