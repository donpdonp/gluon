module.exports = (function(){
  var sessions = {}
  var o = {}

  o.generate = function(hostname, nick, name, msg_id) {
    var session = {
                    id: newId(36, 5),
                    state: 'new',
                    server: {caps: {}},
                    channels: [],
                    hostname: hostname,
                    network: hostname,
                    nick: nick,
                    name: name,
                    msg_id: msg_id
                  }
    sessions[session.id] = session
    return session
  }

  o.get = function(id) {
    return sessions[id]
  }

  o.list = function() {
    return Object.keys(sessions).map(function(key){return sessions[key]})
  }

  o.search = function(channel) {
    var matches = Object.keys(sessions).filter(function(sid){
      console.log('irc searching for', channel, sessions[sid].channels)
      return sessions[sid].channels.indexOf(channel) > -1
    })
    return sessions[matches[0]]
  }

  function newId(base, length) {
    var width = Math.pow(base,length) - Math.pow(base,length-1)
    var add = Math.floor(Math.random()*width)
    return (Math.pow(base,length-1)+add).toString(base)
  }

  return o
})()

