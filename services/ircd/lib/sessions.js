module.exports = (function(){
  var sessions = {}
  var o = {}

  o.generate = function(hostname, nick, name) {
    var session = {
                    id: newId(36, 6),
                    state: 'new',
                    server: {caps: {}},
                    channels: [],
                    hostname: hostname,
                    nick: nick,
                    name: name
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

  function newId(base, length) {
    var width = Math.pow(base,length) - Math.pow(base,length-1)
    var add = Math.floor(Math.random()*width)
    return (Math.pow(base,length-1)+add).toString(base)
  }

  return o
})()

