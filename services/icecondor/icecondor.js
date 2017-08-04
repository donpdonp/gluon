// node
var net = require("net")
var uuid = require('node-uuid')
var wsock = require('websock');
var Url = require('url');

var settings = require('./settings');

// npm
var redisLib = require("redis"),
    redisSub = redisLib.createClient(),
    redisPub = redisLib.createClient()

var pubsub_channel = 'gluon'
var my_uuid = uuid.v4()

var uri = Url.parse(settings.api)
var stream_id
var usercache = {}
var opened

var ws = new wsock.connect(uri, {agent:{rejectUnauthorized:false}})
ws.on('open', function() {
  redis_pub({method: "icecondor.open"})
  console.log('icecondor connected')
  opened = new Date()
})

ws.on('message', function(data) {
  //process.stdout.write("ic: "+data)
  var msg = JSON.parse(data)
  if(msg.method == "hello") {
    var m = { id: "123",
              method: "auth.session",
              params: {device_key: settings.key}}
    ws.send(JSON.stringify(m))
  }
  if(msg.result) {
    if(msg.id == "123") {
      var m = { id: "456",
                method: "stream.follow",
                params: {type: "location", follow: true}}
      ws.send(JSON.stringify(m))
    } else if (msg.id == "456") {
      console.log('fw', JSON.stringify(msg.result))
      stream_id = msg.result.stream_id
      var added = msg.result.added[0]
      usercache[added.id] = added.username
    } else if (stream_id && msg.id == stream_id) {
      var username = usercache[msg.result.user_id]
      var ldate = new Date(msg.result.date)
      var ago = ((new Date()).getTime() - ldate.getTime())/1000/60/60
      console.log(username, msg.result.latitude, msg.result.longitude, ago, "hours ago")
      if (msg.result.latitude) {
        if (ago < 48) {
          redis_pub({method: "icecondor.location",
                    params: {username: username,
                             latitude: msg.result.latitude,
                             longitude: msg.result.longitude,
                             date: msg.result.date,
                             accuracy: msg.result.accuracy
                           }})
        } else {
          console.log(username, 'too old')
        }
      } else {
        console.log(username, 'cloaked')
      }
    } else {
      console.log('unknown response:', msg)
    }
  }
})

ws.on('error', function(data) {
  console.error(data)
})

ws.on('close', function() {
  redis_pub({method: "icecondor.closed"})
  var minutes = ((new Date()).getTime() - opened.getTime())/1000/60
  console.log("closed. duration "+minutes.toFixed(1)+"min")
})


redisSub.on("subscribe", function (channel, count) {
  console.log("redis subscribe "+channel)
})

redisSub.on("message", function (channel, message) {
  try {
    var payload = JSON.parse(message)
    if(payload.method && payload.method.match(/^db\./)) {
      dispatch(payload)
    }
  } catch(err) {
      console.log("redis json err %s", err);
  }
})

redisSub.subscribe(pubsub_channel)

function redis_pub(msg){
  msg["from"] = my_uuid
  msg["id"] = msg["id"] || uuid.v4()
  msg["key"] = settings.gluon_key
  var json = JSON.stringify(msg)
  console.log('redis>', json)
  redisPub.publish(pubsub_channel, json)
}

function dispatch(payload) {
  var cmd = payload.method.split('.')[1]
  if(cmd == 'delete') {
  }
}

